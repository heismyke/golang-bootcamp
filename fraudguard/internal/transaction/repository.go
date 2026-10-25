package transaction

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"fraudguard/internal/fraud"
)

// Repository is the storage port the Service depends on. Postgres implements
// it in production; MemoryRepository implements it for tests.
type Repository interface {
	CreateAccount(ctx context.Context, in NewAccount) (Account, error)
	GetAccount(ctx context.Context, id string) (Account, error)
	GetTransaction(ctx context.Context, id string) (Transaction, error)
	// ListTransactions returns an account's transactions newest first,
	// starting strictly after the cursor when one is given.
	ListTransactions(ctx context.Context, accountID string, after *Cursor, limit int) ([]Transaction, error)
	// RunInTx runs fn inside a database transaction. If fn returns an error
	// everything it did is rolled back.
	RunInTx(ctx context.Context, fn func(TxRepository) error) error
}

// TxRepository holds the operations that must happen atomically while a
// transaction is scored.
type TxRepository interface {
	// LockAccount loads the account and blocks other transactions from
	// locking it until this one finishes.
	LockAccount(ctx context.Context, id string) (Account, error)
	FindByIdempotencyKey(ctx context.Context, key string) (Transaction, error)
	History(ctx context.Context, accountID string, at time.Time) (fraud.History, error)
	InsertTransaction(ctx context.Context, t Transaction) (Transaction, error)
}

// Cursor marks a position in a keyset-paginated listing. Keyset pagination
// stays fast on deep pages, unlike OFFSET, which scans every skipped row.
type Cursor struct {
	OccurredAt time.Time
	ID         string
}

func (c Cursor) Encode() string {
	raw := c.OccurredAt.UTC().Format(time.RFC3339Nano) + "|" + c.ID
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(s string) (Cursor, error) {
	invalid := &ValidationError{Fields: map[string]string{"cursor": "is invalid"}}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, invalid
	}
	ts, id, ok := strings.Cut(string(raw), "|")
	if !ok || id == "" {
		return Cursor{}, invalid
	}
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return Cursor{}, invalid
	}
	return Cursor{OccurredAt: t, ID: id}, nil
}
