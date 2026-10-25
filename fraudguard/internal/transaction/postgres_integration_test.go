//go:build integration

// Run with: make test-integration (requires the database from docker compose).
package transaction_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"

	"fraudguard/internal/fraud"
	"fraudguard/internal/platform/postgres"
	"fraudguard/internal/transaction"
)

func newPostgresService(t *testing.T) (*transaction.Service, transaction.Account) {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := postgres.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := postgres.Migrate(ctx, pool, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatal(err)
	}

	svc := transaction.NewService(transaction.NewPostgresRepository(pool), fraud.DefaultEngine())
	acct, err := svc.CreateAccount(ctx, transaction.NewAccount{ExternalRef: "it_" + rand.Text(), HomeCountry: "JP"})
	if err != nil {
		t.Fatal(err)
	}
	return svc, acct
}

func TestPostgresConcurrentReplaysCreateOnce(t *testing.T) {
	svc, acct := newPostgresService(t)
	key := "it_" + rand.Text()

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		created int
		ids     = make(map[string]bool)
	)
	for range 20 {
		wg.Go(func() {
			tx, c, err := svc.Ingest(context.Background(), purchase(acct.ID, key))
			if err != nil {
				t.Error(err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			ids[tx.ID] = true
			if c {
				created++
			}
		})
	}
	wg.Wait()
	if created != 1 || len(ids) != 1 {
		t.Errorf("created %d rows with %d distinct IDs, want 1 and 1", created, len(ids))
	}
}

// The account row lock serializes concurrent requests, so exactly the
// transactions beyond the 5-per-hour limit are flagged, regardless of timing.
func TestPostgresVelocityUnderConcurrency(t *testing.T) {
	svc, acct := newPostgresService(t)
	const n = 10

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		flagged int
	)
	for i := range n {
		wg.Go(func() {
			tx, _, err := svc.Ingest(context.Background(), purchase(acct.ID, fmt.Sprintf("it_%s_%d", acct.ID, i)))
			if err != nil {
				t.Error(err)
				return
			}
			if hasReason(tx, "velocity") {
				mu.Lock()
				flagged++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if flagged != n-5 {
		t.Errorf("flagged %d transactions for velocity, want %d", flagged, n-5)
	}
}

func TestPostgresNotFound(t *testing.T) {
	svc, _ := newPostgresService(t)
	ctx := context.Background()
	for _, id := range []string{"not-a-uuid", "00000000-0000-0000-0000-000000000000"} {
		if _, err := svc.GetTransaction(ctx, id); err == nil {
			t.Errorf("GetTransaction(%q) succeeded, want not found", id)
		}
	}
}
