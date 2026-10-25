package fraud

import (
	"testing"
	"time"
)

func TestRules(t *testing.T) {
	established := Features{HomeCountry: "JP", AccountAge: 365 * 24 * time.Hour}

	tests := []struct {
		name string
		rule Rule
		tx   Transaction
		f    Features
		want bool
	}{
		{"high amount below threshold", HighAmount{ThresholdMinor: 100}, Transaction{AmountMinor: 100}, established, false},
		{"high amount above threshold", HighAmount{ThresholdMinor: 100}, Transaction{AmountMinor: 101}, established, true},
		{"velocity at limit", Velocity{MaxPerHour: 5}, Transaction{}, withHistory(established, History{TxLastHour: 4}), false},
		{"velocity over limit", Velocity{MaxPerHour: 5}, Transaction{}, withHistory(established, History{TxLastHour: 5}), true},
		{"same country", CountryMismatch{}, Transaction{Country: "JP"}, established, false},
		{"foreign country", CountryMismatch{}, Transaction{Country: "BR"}, established, true},
		{"spike with no history", AmountSpike{Multiplier: 5}, Transaction{AmountMinor: 1_000_000}, established, false},
		{"spike within range", AmountSpike{Multiplier: 5}, Transaction{AmountMinor: 500}, withHistory(established, History{AvgAmount30dMinor: 100}), false},
		{"spike over range", AmountSpike{Multiplier: 5}, Transaction{AmountMinor: 501}, withHistory(established, History{AvgAmount30dMinor: 100}), true},
		{"normal merchant", HighRiskMerchant{Categories: map[string]bool{"crypto": true}}, Transaction{MerchantCategory: "grocery"}, established, false},
		{"risky merchant", HighRiskMerchant{Categories: map[string]bool{"crypto": true}}, Transaction{MerchantCategory: "crypto"}, established, true},
		{"old account", NewAccount{MaxAge: time.Hour}, Transaction{}, established, false},
		{"new account", NewAccount{MaxAge: time.Hour}, Transaction{}, Features{AccountAge: time.Minute}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := tt.rule.Evaluate(tt.tx, tt.f)
			if got != tt.want {
				t.Errorf("Evaluate() hit = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEngineDecisions(t *testing.T) {
	engine := DefaultEngine()
	established := Features{HomeCountry: "JP", AccountAge: 365 * 24 * time.Hour}

	tests := []struct {
		name      string
		tx        Transaction
		f         Features
		wantScore int
		want      Decision
	}{
		{
			name: "ordinary purchase is approved",
			tx:   Transaction{AmountMinor: 3_000, Currency: "JPY", Country: "JP", MerchantCategory: "grocery"},
			f:    established,
			want: Approve,
		},
		{
			name:      "foreign gambling goes to review",
			tx:        Transaction{AmountMinor: 3_000, Currency: "JPY", Country: "BR", MerchantCategory: "gambling"},
			f:         established,
			wantScore: 45,
			want:      Review,
		},
		{
			name:      "card testing burst on new account is declined",
			tx:        Transaction{AmountMinor: 100, Currency: "JPY", Country: "US", MerchantCategory: "gift_cards"},
			f:         Features{HomeCountry: "JP", AccountAge: time.Hour, History: History{TxLastHour: 9}},
			wantScore: 95,
			want:      Decline,
		},
		{
			name:      "score is capped at 100",
			tx:        Transaction{AmountMinor: 10_000_000, Currency: "JPY", Country: "US", MerchantCategory: "crypto"},
			f:         Features{HomeCountry: "JP", AccountAge: time.Hour, History: History{TxLastHour: 9, AvgAmount30dMinor: 100}},
			wantScore: 100,
			want:      Decline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.Evaluate(tt.tx, tt.f)
			if got.Score != tt.wantScore || got.Decision != tt.want {
				t.Errorf("Evaluate() = score %d decision %s, want score %d decision %s (signals: %+v)",
					got.Score, got.Decision, tt.wantScore, tt.want, got.Signals)
			}
		})
	}
}

func withHistory(f Features, h History) Features {
	f.History = h
	return f
}
