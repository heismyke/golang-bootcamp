package transaction_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"fraudguard/internal/fraud"
	"fraudguard/internal/transaction"
)

func newTestService(t *testing.T) (*transaction.Service, *transaction.MemoryRepository, transaction.Account) {
	t.Helper()
	repo := transaction.NewMemoryRepository()
	svc := transaction.NewService(repo, fraud.DefaultEngine())
	acct, err := svc.CreateAccount(context.Background(), transaction.NewAccount{ExternalRef: "user_1", HomeCountry: "jp"})
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	// Age the account so the new_account rule doesn't affect every test.
	repo.SetAccountCreatedAt(acct.ID, time.Now().Add(-365*24*time.Hour))
	return svc, repo, acct
}

func purchase(accountID, key string) transaction.NewTransaction {
	return transaction.NewTransaction{
		IdempotencyKey:   key,
		AccountID:        accountID,
		AmountMinor:      2_500,
		Currency:         "jpy",
		MerchantID:       "m_123",
		MerchantCategory: "Grocery",
		Country:          "JP",
	}
}

func TestIngestApprovesOrdinaryPurchase(t *testing.T) {
	svc, _, acct := newTestService(t)

	got, created, err := svc.Ingest(context.Background(), purchase(acct.ID, "k1"))
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if !created {
		t.Error("created = false, want true")
	}
	if got.Decision != fraud.Approve || got.RiskScore != 0 {
		t.Errorf("got decision %s score %d, want approve 0", got.Decision, got.RiskScore)
	}
	if got.Currency != "JPY" || got.MerchantCategory != "grocery" {
		t.Errorf("input was not normalized: currency %q category %q", got.Currency, got.MerchantCategory)
	}
}

func TestIngestIsIdempotent(t *testing.T) {
	svc, _, acct := newTestService(t)
	ctx := context.Background()

	first, _, err := svc.Ingest(ctx, purchase(acct.ID, "same-key"))
	if err != nil {
		t.Fatal(err)
	}
	second, created, err := svc.Ingest(ctx, purchase(acct.ID, "same-key"))
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Error("replay reported created = true")
	}
	if first.ID != second.ID {
		t.Errorf("replay returned a different transaction: %s vs %s", first.ID, second.ID)
	}
}

func TestIngestConcurrentReplaysCreateOnce(t *testing.T) {
	svc, _, acct := newTestService(t)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		created int
	)
	for range 50 {
		wg.Go(func() {
			_, c, err := svc.Ingest(context.Background(), purchase(acct.ID, "retry-storm"))
			if err != nil {
				t.Error(err)
				return
			}
			if c {
				mu.Lock()
				created++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if created != 1 {
		t.Errorf("created %d transactions, want exactly 1", created)
	}
}

func TestIngestFlagsVelocity(t *testing.T) {
	svc, _, acct := newTestService(t)
	ctx := context.Background()

	var last transaction.Transaction
	for i := range 6 {
		var err error
		last, _, err = svc.Ingest(ctx, purchase(acct.ID, fmt.Sprintf("burst-%d", i)))
		if err != nil {
			t.Fatal(err)
		}
	}
	if !hasReason(last, "velocity") {
		t.Errorf("6th transaction in an hour should trigger velocity, reasons: %v", last.Reasons)
	}
}

func TestIngestRejectsKeyReuseAcrossAccounts(t *testing.T) {
	svc, _, acct := newTestService(t)
	ctx := context.Background()
	other, err := svc.CreateAccount(ctx, transaction.NewAccount{ExternalRef: "user_2", HomeCountry: "JP"})
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := svc.Ingest(ctx, purchase(acct.ID, "shared")); err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.Ingest(ctx, purchase(other.ID, "shared"))
	if !errors.Is(err, transaction.ErrConflict) {
		t.Errorf("err = %v, want ErrConflict", err)
	}
}

func TestIngestValidation(t *testing.T) {
	svc, _, acct := newTestService(t)

	in := purchase(acct.ID, "")
	in.AmountMinor = -5
	in.Country = "Japan"
	in.OccurredAt = time.Now().Add(time.Hour)

	_, _, err := svc.Ingest(context.Background(), in)
	var ve *transaction.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v, want ValidationError", err)
	}
	for _, field := range []string{"Idempotency-Key", "amount_minor", "country", "occurred_at"} {
		if _, ok := ve.Fields[field]; !ok {
			t.Errorf("missing validation error for %s (got %v)", field, ve.Fields)
		}
	}
}

func TestIngestUnknownAccount(t *testing.T) {
	svc, _, _ := newTestService(t)
	_, _, err := svc.Ingest(context.Background(), purchase("missing", "k"))
	if !errors.Is(err, transaction.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestListTransactionsPaginates(t *testing.T) {
	svc, _, acct := newTestService(t)
	ctx := context.Background()

	base := time.Now().Add(-10 * time.Hour)
	for i := range 7 {
		in := purchase(acct.ID, fmt.Sprintf("p-%d", i))
		in.OccurredAt = base.Add(time.Duration(i) * time.Hour)
		if _, _, err := svc.Ingest(ctx, in); err != nil {
			t.Fatal(err)
		}
	}

	var (
		seen   []string
		cursor string
		pages  int
	)
	for {
		page, err := svc.ListTransactions(ctx, acct.ID, transaction.Page{Limit: 3, Cursor: cursor})
		if err != nil {
			t.Fatal(err)
		}
		pages++
		for _, tx := range page.Items {
			seen = append(seen, tx.IdempotencyKey)
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}

	want := "p-6,p-5,p-4,p-3,p-2,p-1,p-0"
	if got := strings.Join(seen, ","); got != want || pages != 3 {
		t.Errorf("got %s over %d pages, want %s over 3 pages", got, pages, want)
	}
}

func hasReason(t transaction.Transaction, rule string) bool {
	for _, r := range t.Reasons {
		if strings.HasPrefix(r, rule+":") {
			return true
		}
	}
	return false
}
