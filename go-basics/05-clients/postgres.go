package clients

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ═════════════════════════════════════════════════════════════
// PART 2: POSTGRES
// Docs: https://pkg.go.dev/github.com/jackc/pgx/v5 (see pgxpool, and the
// pgx wiki for transactions and COPY).
// Explore by hand: docker compose -f 05-clients/compose.yaml exec postgres psql -U clients
// ═════════════════════════════════════════════════════════════

var (
	ErrAccountExists     = errors.New("account already exists")
	ErrNoAccount         = errors.New("no such account")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("amount must be positive")
)

// ─── 7. Connection pool ───────────────────────────────────────
// NewPostgres opens a pool and pings it. Set sensible limits on the
// config (pgxpool.ParseConfig): MaxConns 10, MaxConnIdleTime 5 minutes.
// A pool shares a few connections among many goroutines. Never open a
// new connection per request.
func NewPostgres(ctx context.Context, url string) (*pgxpool.Pool, error) {
	return nil, ErrTODO
}

// ─── 8. Schema ────────────────────────────────────────────────
// Migrate creates the tables if they don't exist (safe to run many times):
//
//	bank_accounts (id text primary key, balance bigint not null, check balance >= 0)
//	events        (id bigserial primary key, kind text not null,
//	               payload text not null, created_at timestamptz not null default now())
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	return ErrTODO
}

// ─── 9. Basic queries and error mapping ───────────────────────

// CreateAccount inserts an account. A duplicate ID returns ErrAccountExists.
// Hint: errors.As(err, &pgErr) with *pgconn.PgError, code "23505".
// ALWAYS use placeholders ($1, $2), never build SQL with fmt.Sprintf:
// that's how SQL injection happens.
func CreateAccount(ctx context.Context, pool *pgxpool.Pool, id string, balance int64) error {
	return ErrTODO
}

// Balance returns an account's balance, or ErrNoAccount (pgx.ErrNoRows).
func Balance(ctx context.Context, pool *pgxpool.Pool, id string) (int64, error) {
	return 0, ErrTODO
}

// ─── 10. Transactions ─────────────────────────────────────────
// Transfer moves money between two accounts ATOMICALLY: either both
// balances change or neither does.
//
//   - amount <= 0 → ErrInvalidAmount
//   - either account missing → ErrNoAccount
//   - not enough money → ErrInsufficientFunds (and nothing changes)
//
// Use pgx.BeginFunc (or Begin + defer Rollback + Commit). Lock both rows
// with SELECT ... FOR UPDATE before checking the balance.
//
// The test runs 100 transfers concurrently, A→B and B→A at the same time.
// If you lock "from" then "to", two transfers in opposite directions can
// each hold one lock and wait for the other forever: a DEADLOCK (Postgres
// detects it and kills one). Fix: always lock the two rows in the same
// order (e.g. sorted by ID).
func Transfer(ctx context.Context, pool *pgxpool.Pool, from, to string, amount int64) error {
	return ErrTODO
}

// ─── 11. Bulk insert ──────────────────────────────────────────

type Event struct {
	Kind      string
	Payload   string
	CreatedAt time.Time
}

// BulkInsertEvents inserts all events and returns how many were inserted.
// The test inserts 50,000. One INSERT per row takes many seconds; use
// pool.CopyFrom (Postgres COPY protocol) to do it in well under a second.
func BulkInsertEvents(ctx context.Context, pool *pgxpool.Pool, events []Event) (int64, error) {
	return 0, ErrTODO
}
