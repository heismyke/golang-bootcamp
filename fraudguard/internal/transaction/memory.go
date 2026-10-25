package transaction

import (
	"cmp"
	"context"
	"crypto/rand"
	"slices"
	"sync"
	"time"

	"fraudguard/internal/fraud"
)

// MemoryRepository is an in-memory Repository for tests and experiments.
// RunInTx holds one global lock, so transactions are fully serialized, and
// inserts become visible only when fn succeeds.
type MemoryRepository struct {
	mu       sync.Mutex
	accounts map[string]Account
	txs      map[string]Transaction
	byKey    map[string]string // idempotency key -> transaction ID
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		accounts: make(map[string]Account),
		txs:      make(map[string]Transaction),
		byKey:    make(map[string]string),
	}
}

func (m *MemoryRepository) CreateAccount(_ context.Context, in NewAccount) (Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.accounts {
		if a.ExternalRef == in.ExternalRef {
			return Account{}, ErrConflict
		}
	}
	a := Account{ID: rand.Text(), ExternalRef: in.ExternalRef, HomeCountry: in.HomeCountry, CreatedAt: time.Now().UTC()}
	m.accounts[a.ID] = a
	return a, nil
}

// SetAccountCreatedAt lets tests simulate accounts of a given age.
func (m *MemoryRepository) SetAccountCreatedAt(id string, at time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.accounts[id]
	a.CreatedAt = at
	m.accounts[id] = a
}

func (m *MemoryRepository) GetAccount(_ context.Context, id string) (Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.accounts[id]
	if !ok {
		return Account{}, ErrNotFound
	}
	return a, nil
}

func (m *MemoryRepository) GetTransaction(_ context.Context, id string) (Transaction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.txs[id]
	if !ok {
		return Transaction{}, ErrNotFound
	}
	return t, nil
}

func (m *MemoryRepository) ListTransactions(_ context.Context, accountID string, after *Cursor, limit int) ([]Transaction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var out []Transaction
	for _, t := range m.txs {
		if t.AccountID != accountID {
			continue
		}
		if after != nil && !before(t, *after) {
			continue
		}
		out = append(out, t)
	}
	slices.SortFunc(out, func(a, b Transaction) int {
		return cmp.Or(b.OccurredAt.Compare(a.OccurredAt), cmp.Compare(b.ID, a.ID))
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// before reports whether t sorts after the cursor in newest-first order.
func before(t Transaction, c Cursor) bool {
	if !t.OccurredAt.Equal(c.OccurredAt) {
		return t.OccurredAt.Before(c.OccurredAt)
	}
	return t.ID < c.ID
}

func (m *MemoryRepository) RunInTx(_ context.Context, fn func(TxRepository) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx := &memoryTx{m: m}
	if err := fn(tx); err != nil {
		return err
	}
	for _, t := range tx.pending {
		m.txs[t.ID] = t
		m.byKey[t.IdempotencyKey] = t.ID
	}
	return nil
}

// memoryTx runs with m.mu already held by RunInTx.
type memoryTx struct {
	m       *MemoryRepository
	pending []Transaction
}

func (tx *memoryTx) LockAccount(_ context.Context, id string) (Account, error) {
	a, ok := tx.m.accounts[id]
	if !ok {
		return Account{}, ErrNotFound
	}
	return a, nil
}

func (tx *memoryTx) FindByIdempotencyKey(_ context.Context, key string) (Transaction, error) {
	if id, ok := tx.m.byKey[key]; ok {
		return tx.m.txs[id], nil
	}
	for _, t := range tx.pending {
		if t.IdempotencyKey == key {
			return t, nil
		}
	}
	return Transaction{}, ErrNotFound
}

func (tx *memoryTx) History(_ context.Context, accountID string, at time.Time) (fraud.History, error) {
	var (
		h         fraud.History
		sum30d    int64
		count30d  int64
		countries = make(map[string]bool)
	)
	for _, t := range tx.all() {
		if t.AccountID != accountID || !t.OccurredAt.After(at.Add(-30*24*time.Hour)) {
			continue
		}
		sum30d += t.AmountMinor
		count30d++
		if t.OccurredAt.After(at.Add(-time.Hour)) {
			h.TxLastHour++
		}
		if t.OccurredAt.After(at.Add(-24 * time.Hour)) {
			h.AmountLast24hMinor += t.AmountMinor
			countries[t.Country] = true
		}
	}
	h.DistinctCountries24h = len(countries)
	if count30d > 0 {
		h.AvgAmount30dMinor = sum30d / count30d
	}
	return h, nil
}

func (tx *memoryTx) InsertTransaction(_ context.Context, t Transaction) (Transaction, error) {
	if _, err := tx.FindByIdempotencyKey(context.Background(), t.IdempotencyKey); err == nil {
		return Transaction{}, ErrConflict
	}
	t.ID = rand.Text()
	t.CreatedAt = time.Now().UTC()
	tx.pending = append(tx.pending, t)
	return t, nil
}

func (tx *memoryTx) all() []Transaction {
	out := make([]Transaction, 0, len(tx.m.txs)+len(tx.pending))
	for _, t := range tx.m.txs {
		out = append(out, t)
	}
	return append(out, tx.pending...)
}
