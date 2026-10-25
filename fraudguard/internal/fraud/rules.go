package fraud

import (
	"fmt"
	"time"
)

// HighAmount fires when a single transaction exceeds a fixed threshold.
type HighAmount struct {
	ThresholdMinor int64
	Score          int
}

func (r HighAmount) Name() string { return "high_amount" }

func (r HighAmount) Evaluate(tx Transaction, _ Features) (Signal, bool) {
	if tx.AmountMinor <= r.ThresholdMinor {
		return Signal{}, false
	}
	return Signal{r.Name(), r.Score, fmt.Sprintf("amount %d exceeds %d", tx.AmountMinor, r.ThresholdMinor)}, true
}

// Velocity fires when an account transacts unusually often. Card-testing
// attacks typically show up as bursts of small transactions.
type Velocity struct {
	MaxPerHour int
	Score      int
}

func (r Velocity) Name() string { return "velocity" }

func (r Velocity) Evaluate(_ Transaction, f Features) (Signal, bool) {
	count := f.TxLastHour + 1 // include the transaction being scored
	if count <= r.MaxPerHour {
		return Signal{}, false
	}
	return Signal{r.Name(), r.Score, fmt.Sprintf("%d transactions in the last hour (max %d)", count, r.MaxPerHour)}, true
}

// CountryMismatch fires when the transaction happens outside the account's
// home country.
type CountryMismatch struct {
	Score int
}

func (r CountryMismatch) Name() string { return "country_mismatch" }

func (r CountryMismatch) Evaluate(tx Transaction, f Features) (Signal, bool) {
	if f.HomeCountry == "" || tx.Country == f.HomeCountry {
		return Signal{}, false
	}
	return Signal{r.Name(), r.Score, fmt.Sprintf("country %s differs from home country %s", tx.Country, f.HomeCountry)}, true
}

// AmountSpike fires when a transaction is much larger than the account's
// recent average.
type AmountSpike struct {
	Multiplier int64
	Score      int
}

func (r AmountSpike) Name() string { return "amount_spike" }

func (r AmountSpike) Evaluate(tx Transaction, f Features) (Signal, bool) {
	if f.AvgAmount30dMinor == 0 || tx.AmountMinor <= f.AvgAmount30dMinor*r.Multiplier {
		return Signal{}, false
	}
	return Signal{r.Name(), r.Score, fmt.Sprintf("amount %d is over %dx the 30-day average %d", tx.AmountMinor, r.Multiplier, f.AvgAmount30dMinor)}, true
}

// HighRiskMerchant fires for merchant categories commonly used to cash out
// stolen funds.
type HighRiskMerchant struct {
	Categories map[string]bool
	Score      int
}

func (r HighRiskMerchant) Name() string { return "high_risk_merchant" }

func (r HighRiskMerchant) Evaluate(tx Transaction, _ Features) (Signal, bool) {
	if !r.Categories[tx.MerchantCategory] {
		return Signal{}, false
	}
	return Signal{r.Name(), r.Score, fmt.Sprintf("merchant category %q is high risk", tx.MerchantCategory)}, true
}

// NewAccount fires for accounts younger than MaxAge.
type NewAccount struct {
	MaxAge time.Duration
	Score  int
}

func (r NewAccount) Name() string { return "new_account" }

func (r NewAccount) Evaluate(_ Transaction, f Features) (Signal, bool) {
	if f.AccountAge >= r.MaxAge {
		return Signal{}, false
	}
	return Signal{r.Name(), r.Score, fmt.Sprintf("account is younger than %s", r.MaxAge)}, true
}
