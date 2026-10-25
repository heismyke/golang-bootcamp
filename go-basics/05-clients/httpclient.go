package clients

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ═════════════════════════════════════════════════════════════
// PART 3: CALLING OTHER HTTP APIS
// Backends call payment providers, maps, email services, other internal
// services... and networks fail. A production HTTP client needs timeouts,
// retries with backoff, and careful error handling.
// ═════════════════════════════════════════════════════════════

type Weather struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_c"`
}

// APIError is returned when the API answers with an error status that
// shouldn't be retried. Message comes from the response body {"error": "..."}.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string { return fmt.Sprintf("api error %d: %s", e.StatusCode, e.Message) }

type WeatherClient struct {
	BaseURL    string        // e.g. "https://api.example.com"
	HTTP       *http.Client  // has its own Timeout per attempt
	MaxRetries int           // extra attempts after the first one
	Backoff    time.Duration // wait before the 1st retry; doubles each time
}

// ─── 12. A robust API client ──────────────────────────────────
// Get calls GET {BaseURL}/weather?city={city} and decodes the JSON.
//
//   - Build the URL with net/url so "São Paulo" is encoded correctly.
//   - Use http.NewRequestWithContext, so the caller can cancel.
//   - Retry on network errors (including timeouts), 5xx, and 429 Too Many
//     Requests, up to MaxRetries times. Wait Backoff, then 2×Backoff, 4×...
//   - On 429 with a "Retry-After: N" header (seconds), wait N seconds instead.
//   - Do NOT retry other 4xx errors (404, 400...): retrying won't help.
//     Return *APIError with the status and the "error" field of the body.
//   - If ctx is cancelled (even while waiting to retry), stop immediately
//     and return ctx.Err().
//   - Always close resp.Body, or connections leak.
//   - After the last failed retry, return an error that mentions the last
//     problem.
//
// Hint: wait with select { case <-time.After(d): case <-ctx.Done(): }
func (c *WeatherClient) Get(ctx context.Context, city string) (Weather, error) {
	return Weather{}, ErrTODO
}
