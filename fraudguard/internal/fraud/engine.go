// Package fraud contains the risk-scoring engine. It is pure business logic:
// no database, no HTTP, no clock. Everything it needs is passed in, which makes
// it fast and trivial to test.
package fraud

import "time"

// Decision is the action the caller should take on a transaction.
type Decision string

const (
	Approve Decision = "approve"
	Review  Decision = "review"
	Decline Decision = "decline"
)

// Transaction is the subset of a payment the engine scores.
type Transaction struct {
	AmountMinor      int64
	Currency         string
	Country          string
	MerchantCategory string
}

// History holds aggregates over an account's previous transactions.
type History struct {
	TxLastHour           int
	AmountLast24hMinor   int64
	DistinctCountries24h int
	AvgAmount30dMinor    int64
}

// Features is everything known about the account at scoring time.
type Features struct {
	HomeCountry string
	AccountAge  time.Duration
	History
}

// Signal explains why a rule contributed to the score.
type Signal struct {
	Rule   string
	Score  int
	Reason string
}

// Rule inspects a transaction and reports whether it looks risky.
type Rule interface {
	Name() string
	Evaluate(tx Transaction, f Features) (Signal, bool)
}

type Result struct {
	Score    int
	Decision Decision
	Signals  []Signal
}

// Engine sums the scores of every rule that fires and maps the total to a
// decision using two thresholds.
type Engine struct {
	rules            []Rule
	reviewThreshold  int
	declineThreshold int
}

func NewEngine(reviewThreshold, declineThreshold int, rules ...Rule) *Engine {
	return &Engine{rules: rules, reviewThreshold: reviewThreshold, declineThreshold: declineThreshold}
}

// DefaultEngine returns the rule set used in production.
func DefaultEngine() *Engine {
	return NewEngine(40, 80,
		HighAmount{ThresholdMinor: 500_000, Score: 30},
		Velocity{MaxPerHour: 5, Score: 35},
		CountryMismatch{Score: 25},
		AmountSpike{Multiplier: 5, Score: 30},
		HighRiskMerchant{Categories: map[string]bool{"gambling": true, "crypto": true, "gift_cards": true}, Score: 20},
		NewAccount{MaxAge: 24 * time.Hour, Score: 15},
	)
}

func (e *Engine) Evaluate(tx Transaction, f Features) Result {
	var res Result
	for _, r := range e.rules {
		if s, hit := r.Evaluate(tx, f); hit {
			res.Score += s.Score
			res.Signals = append(res.Signals, s)
		}
	}
	res.Score = min(res.Score, 100)

	switch {
	case res.Score >= e.declineThreshold:
		res.Decision = Decline
	case res.Score >= e.reviewThreshold:
		res.Decision = Review
	default:
		res.Decision = Approve
	}
	return res
}
