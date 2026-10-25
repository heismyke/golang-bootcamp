package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutes(t *testing.T) {
	a := newApp(&memoryStore{}, "test-host", slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(a.routes())
	defer srv.Close()

	get := func(method, path string) (int, string) {
		req, _ := http.NewRequest(method, srv.URL+path, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b)
	}

	if code, body := get("GET", "/"); code != 200 || !strings.Contains(body, "test-host") {
		t.Errorf("GET / = %d %s", code, body)
	}
	get("POST", "/visits")
	if code, body := get("POST", "/visits"); code != 200 || !strings.Contains(body, `"count":2`) {
		t.Errorf("POST /visits = %d %s, want count 2", code, body)
	}
	if code, _ := get("GET", "/readyz"); code != 200 {
		t.Errorf("GET /readyz = %d, want 200", code)
	}
	if code, _ := get("GET", "/nope"); code != 404 {
		t.Errorf("GET /nope = %d, want 404", code)
	}
	code, body := get("GET", "/metrics")
	if code != 200 || !strings.Contains(body, `http_requests_total{path="/visits",code="200"} 2`) || !strings.Contains(body, "visits_total 2") {
		t.Errorf("GET /metrics = %d\n%s", code, body)
	}

	a.draining.Store(true)
	if code, _ := get("GET", "/readyz"); code != 503 {
		t.Errorf("GET /readyz while draining = %d, want 503", code)
	}
}
