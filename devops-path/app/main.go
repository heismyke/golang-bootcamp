// Command visits is a small web service used as the workload for the DevOps
// track. You don't need to change it: your job is to build, ship, run and
// monitor it.
//
// Configuration (environment variables):
//
//	PORT            port to listen on (default 8080)
//	DATABASE_URL    Postgres connection string; if empty, counts are kept in memory
//	SHUTDOWN_DELAY  how long to keep serving after SIGTERM while reporting
//	                not-ready, so load balancers can stop sending traffic (default 0s)
//	LOG_LEVEL       debug, info, warn or error (default info)
//
// Flags:
//
//	-healthcheck    call /healthz on the local server and exit 0 (healthy) or 1,
//	                for Docker HEALTHCHECK in images without curl or a shell
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

// version is set at build time: go build -ldflags "-X main.version=1.2.3"
var version = "dev"

func main() {
	healthcheck := flag.Bool("healthcheck", false, "check the local server's /healthz and exit")
	flag.Parse()

	port := getenv("PORT", "8080")
	if *healthcheck {
		os.Exit(runHealthcheck(port))
	}

	if err := run(port); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run(port string) error {
	var level slog.Level
	if err := level.UnmarshalText([]byte(getenv("LOG_LEVEL", "info"))); err != nil {
		return fmt.Errorf("LOG_LEVEL: %w", err)
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	shutdownDelay, err := time.ParseDuration(getenv("SHUTDOWN_DELAY", "0s"))
	if err != nil {
		return fmt.Errorf("SHUTDOWN_DELAY: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var st store = &memoryStore{}
	if url := os.Getenv("DATABASE_URL"); url != "" {
		pg, err := connectPostgres(ctx, url, log)
		if err != nil {
			return err
		}
		defer pg.Close()
		st = pg
	} else {
		log.Warn("DATABASE_URL not set, keeping counts in memory")
	}

	hostname, _ := os.Hostname()
	app := newApp(st, hostname, log)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           app.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Info("listening", "port", port, "version", version, "hostname", hostname, "storage", st.Kind())
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	// Report not-ready first and keep serving for a moment, so load
	// balancers and Kubernetes stop routing new requests here before we stop.
	app.draining.Store(true)
	log.Info("shutdown signal received", "delay", shutdownDelay)
	time.Sleep(shutdownDelay)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Info("stopped")
	return nil
}

type app struct {
	store    store
	hostname string
	log      *slog.Logger
	metrics  *metrics
	draining atomic.Bool
}

func newApp(st store, hostname string, log *slog.Logger) *app {
	return &app{store: st, hostname: hostname, log: log, metrics: newMetrics()}
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", a.index)
	mux.HandleFunc("GET /visits", a.getVisits)
	mux.HandleFunc("POST /visits", a.addVisit)
	mux.HandleFunc("GET /healthz", a.healthz)
	mux.HandleFunc("GET /readyz", a.readyz)
	mux.HandleFunc("GET /metrics", a.serveMetrics)
	return a.instrument(mux)
}

func (a *app) index(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"message":  "hello from visits",
		"version":  version,
		"hostname": a.hostname,
	})
}

func (a *app) getVisits(w http.ResponseWriter, r *http.Request) {
	n, err := a.store.Count(r.Context())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": n, "hostname": a.hostname})
}

func (a *app) addVisit(w http.ResponseWriter, r *http.Request) {
	n, err := a.store.Incr(r.Context())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": n, "hostname": a.hostname})
}

// healthz is the liveness probe: the process is running.
func (a *app) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readyz is the readiness probe: the process can serve traffic right now.
func (a *app) readyz(w http.ResponseWriter, r *http.Request) {
	if a.draining.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "shutting down"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := a.store.Ping(ctx); err != nil {
		a.log.Warn("readiness check failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "storage": a.store.Kind()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready", "storage": a.store.Kind()})
}

func (a *app) fail(w http.ResponseWriter, r *http.Request, err error) {
	a.log.Error("request failed", "path", r.URL.Path, "error", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func runHealthcheck(port string) int {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/healthz")
	if err != nil {
		fmt.Fprintln(os.Stderr, "unhealthy:", err)
		return 1
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "unhealthy: status", resp.StatusCode)
		return 1
	}
	return 0
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
