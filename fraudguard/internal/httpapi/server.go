// Package httpapi exposes the service over HTTP/JSON.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"fraudguard/internal/transaction"
)

type Server struct {
	svc   *transaction.Service
	log   *slog.Logger
	ready func(context.Context) error
}

// NewHandler builds the routing tree. ready reports whether dependencies
// (the database) are reachable, for the readiness probe.
func NewHandler(svc *transaction.Service, log *slog.Logger, ready func(context.Context) error) http.Handler {
	s := &Server{svc: svc, log: log, ready: ready}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
	mux.HandleFunc("POST /v1/accounts", s.createAccount)
	mux.HandleFunc("GET /v1/accounts/{id}", s.getAccount)
	mux.HandleFunc("GET /v1/accounts/{id}/transactions", s.listTransactions)
	mux.HandleFunc("POST /v1/transactions", s.createTransaction)
	mux.HandleFunc("GET /v1/transactions/{id}", s.getTransaction)

	// Outermost first: every request gets an ID, is logged, and is protected
	// from panics.
	return requestID(logRequests(log, recoverPanics(log, mux)))
}

// healthz is the liveness probe: the process is up and serving.
func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readyz is the readiness probe: the service can handle traffic. Load
// balancers stop routing here while it fails.
func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.ready(ctx); err != nil {
		s.log.WarnContext(ctx, "readiness check failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	var in transaction.NewAccount
	if err := decodeJSON(w, r, &in); err != nil {
		s.writeError(w, r, err)
		return
	}
	acct, err := s.svc.CreateAccount(r.Context(), in)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, acct)
}

func (s *Server) getAccount(w http.ResponseWriter, r *http.Request) {
	acct, err := s.svc.GetAccount(r.Context(), r.PathValue("id"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, acct)
}

func (s *Server) listTransactions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := transaction.Page{Cursor: q.Get("cursor")}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			s.writeError(w, r, &badRequestError{"limit must be a positive integer"})
			return
		}
		page.Limit = n
	}
	result, err := s.svc.ListTransactions(r.Context(), r.PathValue("id"), page)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	var in transaction.NewTransaction
	if err := decodeJSON(w, r, &in); err != nil {
		s.writeError(w, r, err)
		return
	}
	in.IdempotencyKey = r.Header.Get("Idempotency-Key")

	tx, created, err := s.svc.Ingest(r.Context(), in)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
		w.Header().Set("Idempotent-Replayed", "true")
	}
	writeJSON(w, status, tx)
}

func (s *Server) getTransaction(w http.ResponseWriter, r *http.Request) {
	tx, err := s.svc.GetTransaction(r.Context(), r.PathValue("id"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, tx)
}
