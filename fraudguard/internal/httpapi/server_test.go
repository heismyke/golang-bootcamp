package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fraudguard/internal/fraud"
	"fraudguard/internal/httpapi"
	"fraudguard/internal/transaction"
)

func newTestServer(t *testing.T, ready error) *httptest.Server {
	t.Helper()
	svc := transaction.NewService(transaction.NewMemoryRepository(), fraud.DefaultEngine())
	h := httpapi.NewHandler(svc, slog.New(slog.DiscardHandler), func(context.Context) error { return ready })
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, srv *httptest.Server, method, path, body string, headers map[string]string) (*http.Response, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp, out
}

func TestTransactionLifecycle(t *testing.T) {
	srv := newTestServer(t, nil)

	resp, acct := do(t, srv, "POST", "/v1/accounts", `{"external_ref":"user_42","home_country":"JP"}`, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create account: status %d, body %v", resp.StatusCode, acct)
	}
	accountID := acct["id"].(string)

	body := `{"account_id":"` + accountID + `","amount_minor":1200,"currency":"JPY",
		"merchant_id":"m_1","merchant_category":"restaurants","country":"JP"}`
	headers := map[string]string{"Idempotency-Key": "order-1"}

	resp, tx := do(t, srv, "POST", "/v1/transactions", body, headers)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create transaction: status %d, body %v", resp.StatusCode, tx)
	}
	if tx["decision"] == nil || resp.Header.Get("X-Request-ID") == "" {
		t.Errorf("missing decision or request ID: %v", tx)
	}

	resp, replay := do(t, srv, "POST", "/v1/transactions", body, headers)
	if resp.StatusCode != http.StatusOK || resp.Header.Get("Idempotent-Replayed") != "true" {
		t.Errorf("replay: status %d, replayed header %q", resp.StatusCode, resp.Header.Get("Idempotent-Replayed"))
	}
	if replay["id"] != tx["id"] {
		t.Errorf("replay returned id %v, want %v", replay["id"], tx["id"])
	}

	resp, got := do(t, srv, "GET", "/v1/transactions/"+tx["id"].(string), "", nil)
	if resp.StatusCode != http.StatusOK || got["idempotency_key"] != "order-1" {
		t.Errorf("get transaction: status %d, body %v", resp.StatusCode, got)
	}

	resp, list := do(t, srv, "GET", "/v1/accounts/"+accountID+"/transactions?limit=10", "", nil)
	if resp.StatusCode != http.StatusOK || len(list["items"].([]any)) != 1 {
		t.Errorf("list: status %d, body %v", resp.StatusCode, list)
	}
}

func TestErrorResponses(t *testing.T) {
	srv := newTestServer(t, nil)

	tests := []struct {
		name, method, path, body string
		headers                  map[string]string
		wantStatus               int
		wantCode                 string
	}{
		{"malformed json", "POST", "/v1/accounts", `{"external_ref":`, nil, 400, "bad_request"},
		{"unknown field", "POST", "/v1/accounts", `{"external_ref":"a","home_country":"JP","admin":true}`, nil, 400, "bad_request"},
		{"invalid account", "POST", "/v1/accounts", `{"external_ref":"","home_country":"Japan"}`, nil, 422, "validation_failed"},
		{"missing idempotency key", "POST", "/v1/transactions", `{"account_id":"x","amount_minor":1,"currency":"JPY","merchant_id":"m","merchant_category":"c","country":"JP"}`, nil, 422, "validation_failed"},
		{"unknown account", "POST", "/v1/transactions", `{"account_id":"x","amount_minor":1,"currency":"JPY","merchant_id":"m","merchant_category":"c","country":"JP"}`, map[string]string{"Idempotency-Key": "k"}, 404, "not_found"},
		{"missing transaction", "GET", "/v1/transactions/nope", "", nil, 404, "not_found"},
		{"bad limit", "GET", "/v1/accounts/x/transactions?limit=abc", "", nil, 400, "bad_request"},
		{"bad cursor", "GET", "/v1/accounts/x/transactions?cursor=!!!", "", nil, 422, "validation_failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := do(t, srv, tt.method, tt.path, tt.body, tt.headers)
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status %d, want %d (body %v)", resp.StatusCode, tt.wantStatus, body)
			}
			errObj, _ := body["error"].(map[string]any)
			if errObj["code"] != tt.wantCode {
				t.Errorf("error code %v, want %s", errObj["code"], tt.wantCode)
			}
		})
	}
}

func TestReadiness(t *testing.T) {
	if resp, _ := do(t, newTestServer(t, nil), "GET", "/readyz", "", nil); resp.StatusCode != http.StatusOK {
		t.Errorf("healthy: status %d, want 200", resp.StatusCode)
	}
	if resp, _ := do(t, newTestServer(t, errors.New("db down")), "GET", "/readyz", "", nil); resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("unhealthy: status %d, want 503", resp.StatusCode)
	}
}
