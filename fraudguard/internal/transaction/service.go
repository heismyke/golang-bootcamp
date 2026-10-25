package transaction

import (
	"context"
	"errors"
	"fmt"
	"time"

	"fraudguard/internal/fraud"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type Service struct {
	repo   Repository
	engine *fraud.Engine
	now    func() time.Time
}

func NewService(repo Repository, engine *fraud.Engine) *Service {
	return &Service{repo: repo, engine: engine, now: time.Now}
}

func (s *Service) CreateAccount(ctx context.Context, in NewAccount) (Account, error) {
	if err := in.Validate(); err != nil {
		return Account{}, err
	}
	acct, err := s.repo.CreateAccount(ctx, in)
	if errors.Is(err, ErrConflict) {
		return Account{}, fmt.Errorf("account with external_ref %q already exists: %w", in.ExternalRef, ErrConflict)
	}
	return acct, err
}

func (s *Service) GetAccount(ctx context.Context, id string) (Account, error) {
	acct, err := s.repo.GetAccount(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return Account{}, fmt.Errorf("account %q: %w", id, ErrNotFound)
	}
	return acct, err
}

func (s *Service) GetTransaction(ctx context.Context, id string) (Transaction, error) {
	t, err := s.repo.GetTransaction(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return Transaction{}, fmt.Errorf("transaction %q: %w", id, ErrNotFound)
	}
	return t, err
}

func (s *Service) ListTransactions(ctx context.Context, accountID string, p Page) (TransactionPage, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = defaultPageSize
	}
	limit = min(limit, maxPageSize)

	var after *Cursor
	if p.Cursor != "" {
		c, err := DecodeCursor(p.Cursor)
		if err != nil {
			return TransactionPage{}, err
		}
		after = &c
	}

	if _, err := s.GetAccount(ctx, accountID); err != nil {
		return TransactionPage{}, err
	}

	// Fetch one extra row to learn whether another page exists.
	items, err := s.repo.ListTransactions(ctx, accountID, after, limit+1)
	if err != nil {
		return TransactionPage{}, err
	}
	page := TransactionPage{Items: items}
	if len(items) > limit {
		page.Items = items[:limit]
		last := page.Items[limit-1]
		page.NextCursor = Cursor{OccurredAt: last.OccurredAt, ID: last.ID}.Encode()
	}
	if page.Items == nil {
		page.Items = []Transaction{}
	}
	return page, nil
}

// Ingest scores a transaction and records it. It is idempotent on
// IdempotencyKey: replaying a request returns the original result with
// created == false, so clients can safely retry after a timeout.
func (s *Service) Ingest(ctx context.Context, in NewTransaction) (t Transaction, created bool, err error) {
	now := s.now().UTC()
	if in.OccurredAt.IsZero() {
		in.OccurredAt = now
	}
	if err := in.Validate(now); err != nil {
		return Transaction{}, false, err
	}

	err = s.repo.RunInTx(ctx, func(r TxRepository) error {
		// Locking the account row serializes concurrent transactions for the
		// same account. Without it, two simultaneous requests would both see
		// the same history and a velocity attack could slip through.
		acct, err := r.LockAccount(ctx, in.AccountID)
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("account %q: %w", in.AccountID, ErrNotFound)
		}
		if err != nil {
			return err
		}

		existing, err := r.FindByIdempotencyKey(ctx, in.IdempotencyKey)
		switch {
		case err == nil:
			if existing.AccountID != acct.ID {
				return fmt.Errorf("idempotency key already used for a different account: %w", ErrConflict)
			}
			t = existing
			return nil
		case !errors.Is(err, ErrNotFound):
			return err
		}

		hist, err := r.History(ctx, acct.ID, in.OccurredAt)
		if err != nil {
			return err
		}
		result := s.engine.Evaluate(
			fraud.Transaction{
				AmountMinor:      in.AmountMinor,
				Currency:         in.Currency,
				Country:          in.Country,
				MerchantCategory: in.MerchantCategory,
			},
			fraud.Features{
				HomeCountry: acct.HomeCountry,
				AccountAge:  in.OccurredAt.Sub(acct.CreatedAt),
				History:     hist,
			},
		)

		reasons := make([]string, 0, len(result.Signals))
		for _, sig := range result.Signals {
			reasons = append(reasons, sig.Rule+": "+sig.Reason)
		}

		t, err = r.InsertTransaction(ctx, Transaction{
			IdempotencyKey:   in.IdempotencyKey,
			AccountID:        acct.ID,
			AmountMinor:      in.AmountMinor,
			Currency:         in.Currency,
			MerchantID:       in.MerchantID,
			MerchantCategory: in.MerchantCategory,
			Country:          in.Country,
			OccurredAt:       in.OccurredAt,
			RiskScore:        result.Score,
			Decision:         result.Decision,
			Reasons:          reasons,
		})
		if errors.Is(err, ErrConflict) {
			return fmt.Errorf("idempotency key already used for a different account: %w", ErrConflict)
		}
		if err != nil {
			return err
		}
		created = true
		return nil
	})
	if err != nil {
		return Transaction{}, false, err
	}
	return t, created, nil
}
