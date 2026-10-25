package transaction

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"fraudguard/internal/fraud"
)

// querier is satisfied by both *pgxpool.Pool and pgx.Tx, so the same query
// code runs inside or outside a transaction.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

const (
	accountColumns     = `id, external_ref, home_country, created_at`
	transactionColumns = `id, idempotency_key, account_id, amount_minor, currency, merchant_id,
		merchant_category, country, occurred_at, risk_score, decision, reasons, created_at`
)

func (r *PostgresRepository) CreateAccount(ctx context.Context, in NewAccount) (Account, error) {
	return scanAccount(r.pool.QueryRow(ctx,
		`INSERT INTO accounts (external_ref, home_country) VALUES ($1, $2) RETURNING `+accountColumns,
		in.ExternalRef, in.HomeCountry,
	))
}

func (r *PostgresRepository) GetAccount(ctx context.Context, id string) (Account, error) {
	return scanAccount(r.pool.QueryRow(ctx, `SELECT `+accountColumns+` FROM accounts WHERE id = $1`, id))
}

func (r *PostgresRepository) GetTransaction(ctx context.Context, id string) (Transaction, error) {
	return scanTransaction(r.pool.QueryRow(ctx, `SELECT `+transactionColumns+` FROM transactions WHERE id = $1`, id))
}

func (r *PostgresRepository) ListTransactions(ctx context.Context, accountID string, after *Cursor, limit int) ([]Transaction, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if after == nil {
		rows, err = r.pool.Query(ctx, `
			SELECT `+transactionColumns+` FROM transactions
			WHERE account_id = $1
			ORDER BY occurred_at DESC, id DESC
			LIMIT $2`, accountID, limit)
	} else {
		// Row-value comparison matches the (occurred_at DESC, id DESC) index
		// exactly, so Postgres can seek straight to the cursor position.
		rows, err = r.pool.Query(ctx, `
			SELECT `+transactionColumns+` FROM transactions
			WHERE account_id = $1 AND (occurred_at, id) < ($2, $3)
			ORDER BY occurred_at DESC, id DESC
			LIMIT $4`, accountID, after.OccurredAt, after.ID, limit)
	}
	if err != nil {
		return nil, mapErr(err)
	}
	txs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Transaction, error) {
		return scanTransaction(row)
	})
	return txs, mapErr(err)
}

func (r *PostgresRepository) RunInTx(ctx context.Context, fn func(TxRepository) error) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(&postgresTx{q: tx})
	})
}

type postgresTx struct {
	q querier
}

func (t *postgresTx) LockAccount(ctx context.Context, id string) (Account, error) {
	return scanAccount(t.q.QueryRow(ctx, `SELECT `+accountColumns+` FROM accounts WHERE id = $1 FOR UPDATE`, id))
}

func (t *postgresTx) FindByIdempotencyKey(ctx context.Context, key string) (Transaction, error) {
	return scanTransaction(t.q.QueryRow(ctx, `SELECT `+transactionColumns+` FROM transactions WHERE idempotency_key = $1`, key))
}

func (t *postgresTx) History(ctx context.Context, accountID string, at time.Time) (fraud.History, error) {
	// One pass over the account's last 30 days computes every feature.
	// FILTER narrows each aggregate to its own time window.
	var h fraud.History
	err := t.q.QueryRow(ctx, `
		SELECT
			count(*)                       FILTER (WHERE occurred_at > $2::timestamptz - interval '1 hour'),
			coalesce(sum(amount_minor)     FILTER (WHERE occurred_at > $2::timestamptz - interval '24 hours'), 0),
			count(DISTINCT country)        FILTER (WHERE occurred_at > $2::timestamptz - interval '24 hours'),
			coalesce(avg(amount_minor), 0)::bigint
		FROM transactions
		WHERE account_id = $1
		  AND occurred_at > $2::timestamptz - interval '30 days'`,
		accountID, at,
	).Scan(&h.TxLastHour, &h.AmountLast24hMinor, &h.DistinctCountries24h, &h.AvgAmount30dMinor)
	return h, mapErr(err)
}

func (t *postgresTx) InsertTransaction(ctx context.Context, tx Transaction) (Transaction, error) {
	return scanTransaction(t.q.QueryRow(ctx, `
		INSERT INTO transactions (idempotency_key, account_id, amount_minor, currency, merchant_id,
			merchant_category, country, occurred_at, risk_score, decision, reasons)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING `+transactionColumns,
		tx.IdempotencyKey, tx.AccountID, tx.AmountMinor, tx.Currency, tx.MerchantID,
		tx.MerchantCategory, tx.Country, tx.OccurredAt, tx.RiskScore, string(tx.Decision), tx.Reasons,
	))
}

func scanAccount(row pgx.Row) (Account, error) {
	var a Account
	err := row.Scan(&a.ID, &a.ExternalRef, &a.HomeCountry, &a.CreatedAt)
	return a, mapErr(err)
}

func scanTransaction(row pgx.Row) (Transaction, error) {
	var t Transaction
	var decision string
	err := row.Scan(&t.ID, &t.IdempotencyKey, &t.AccountID, &t.AmountMinor, &t.Currency, &t.MerchantID,
		&t.MerchantCategory, &t.Country, &t.OccurredAt, &t.RiskScore, &decision, &t.Reasons, &t.CreatedAt)
	t.Decision = fraud.Decision(decision)
	return t, mapErr(err)
}

// mapErr translates driver errors into domain errors.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return ErrConflict
		case "22P02": // invalid_text_representation, e.g. an ID that isn't a UUID
			return ErrNotFound
		}
	}
	return err
}
