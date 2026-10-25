// Package transaction implements the core use case: receive a payment, score
// it for fraud, and record the decision.
package transaction

import (
	"strings"
	"time"

	"fraudguard/internal/fraud"
)

type Account struct {
	ID          string    `json:"id"`
	ExternalRef string    `json:"external_ref"`
	HomeCountry string    `json:"home_country"`
	CreatedAt   time.Time `json:"created_at"`
}

type NewAccount struct {
	ExternalRef string `json:"external_ref"`
	HomeCountry string `json:"home_country"`
}

func (n *NewAccount) Validate() error {
	n.ExternalRef = strings.TrimSpace(n.ExternalRef)
	n.HomeCountry = strings.ToUpper(strings.TrimSpace(n.HomeCountry))

	var v validator
	v.check(n.ExternalRef != "" && len(n.ExternalRef) <= 255, "external_ref", "is required (max 255 chars)")
	v.check(isLetters(n.HomeCountry, 2), "home_country", "must be a 2-letter ISO 3166-1 code")
	return v.err()
}

type Transaction struct {
	ID               string         `json:"id"`
	IdempotencyKey   string         `json:"idempotency_key"`
	AccountID        string         `json:"account_id"`
	AmountMinor      int64          `json:"amount_minor"`
	Currency         string         `json:"currency"`
	MerchantID       string         `json:"merchant_id"`
	MerchantCategory string         `json:"merchant_category"`
	Country          string         `json:"country"`
	OccurredAt       time.Time      `json:"occurred_at"`
	RiskScore        int            `json:"risk_score"`
	Decision         fraud.Decision `json:"decision"`
	Reasons          []string       `json:"reasons"`
	CreatedAt        time.Time      `json:"created_at"`
}

type NewTransaction struct {
	IdempotencyKey   string    `json:"-"` // from the Idempotency-Key header
	AccountID        string    `json:"account_id"`
	AmountMinor      int64     `json:"amount_minor"`
	Currency         string    `json:"currency"`
	MerchantID       string    `json:"merchant_id"`
	MerchantCategory string    `json:"merchant_category"`
	Country          string    `json:"country"`
	OccurredAt       time.Time `json:"occurred_at"`
}

// maxClockSkew tolerates clients whose clocks run slightly ahead of ours.
const maxClockSkew = 5 * time.Minute

func (n *NewTransaction) Validate(now time.Time) error {
	n.Currency = strings.ToUpper(strings.TrimSpace(n.Currency))
	n.Country = strings.ToUpper(strings.TrimSpace(n.Country))
	n.MerchantCategory = strings.ToLower(strings.TrimSpace(n.MerchantCategory))
	n.MerchantID = strings.TrimSpace(n.MerchantID)

	var v validator
	v.check(n.IdempotencyKey != "" && len(n.IdempotencyKey) <= 255, "Idempotency-Key", "header is required (max 255 chars)")
	v.check(n.AccountID != "", "account_id", "is required")
	v.check(n.AmountMinor > 0, "amount_minor", "must be greater than 0")
	v.check(isLetters(n.Currency, 3), "currency", "must be a 3-letter ISO 4217 code")
	v.check(n.MerchantID != "", "merchant_id", "is required")
	v.check(n.MerchantCategory != "", "merchant_category", "is required")
	v.check(isLetters(n.Country, 2), "country", "must be a 2-letter ISO 3166-1 code")
	v.check(!n.OccurredAt.After(now.Add(maxClockSkew)), "occurred_at", "cannot be in the future")
	return v.err()
}

type Page struct {
	Limit  int
	Cursor string
}

type TransactionPage struct {
	Items      []Transaction `json:"items"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

func isLetters(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
