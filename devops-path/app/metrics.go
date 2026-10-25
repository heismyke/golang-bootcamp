package main

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"
)

// metrics exposes counters in the Prometheus text format, written by hand
// so you can see exactly what Prometheus scrapes.
type metrics struct {
	mu        sync.Mutex
	requests  map[string]int64   // `path="/",code="200"` → count
	durations map[string]float64 // path → total seconds
	started   time.Time
}

func newMetrics() *metrics {
	return &metrics{requests: map[string]int64{}, durations: map[string]float64{}, started: time.Now()}
}

var knownPaths = map[string]bool{"/": true, "/visits": true, "/healthz": true, "/readyz": true, "/metrics": true}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// instrument records metrics and logs every request.
func (a *app) instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		elapsed := time.Since(start)

		path := r.URL.Path
		if !knownPaths[path] {
			path = "other" // keep label cardinality bounded
		}
		a.metrics.mu.Lock()
		a.metrics.requests[fmt.Sprintf("path=%q,code=\"%d\"", path, rec.status)]++
		a.metrics.durations[path] += elapsed.Seconds()
		a.metrics.mu.Unlock()

		if path != "/healthz" && path != "/readyz" && path != "/metrics" {
			a.log.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status,
				"duration_ms", float64(elapsed.Microseconds())/1000, "remote", r.RemoteAddr)
		}
	})
}

func (a *app) serveMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	m := a.metrics
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Fprintln(w, "# HELP http_requests_total Total HTTP requests by path and status code.")
	fmt.Fprintln(w, "# TYPE http_requests_total counter")
	for _, k := range slices.Sorted(maps.Keys(m.requests)) {
		fmt.Fprintf(w, "http_requests_total{%s} %d\n", k, m.requests[k])
	}
	fmt.Fprintln(w, "# HELP http_request_duration_seconds_sum Total time spent serving requests by path.")
	fmt.Fprintln(w, "# TYPE http_request_duration_seconds_sum counter")
	for _, k := range slices.Sorted(maps.Keys(m.durations)) {
		fmt.Fprintf(w, "http_request_duration_seconds_sum{path=%q} %s\n", k, strconv.FormatFloat(m.durations[k], 'f', 6, 64))
	}
	if n, err := a.store.Count(r.Context()); err == nil {
		fmt.Fprintln(w, "# HELP visits_total Visits recorded.")
		fmt.Fprintln(w, "# TYPE visits_total gauge")
		fmt.Fprintf(w, "visits_total %d\n", n)
	}
	fmt.Fprintln(w, "# HELP app_info Build information.")
	fmt.Fprintln(w, "# TYPE app_info gauge")
	fmt.Fprintf(w, "app_info{version=%q,hostname=%q} 1\n", version, a.hostname)
	fmt.Fprintln(w, "# HELP process_uptime_seconds Seconds since the process started.")
	fmt.Fprintln(w, "# TYPE process_uptime_seconds gauge")
	fmt.Fprintf(w, "process_uptime_seconds %.0f\n", time.Since(m.started).Seconds())
}
