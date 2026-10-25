package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type store interface {
	Incr(ctx context.Context) (int64, error)
	Count(ctx context.Context) (int64, error)
	Ping(ctx context.Context) error
	Kind() string
}

type memoryStore struct {
	n atomic.Int64
}

func (m *memoryStore) Incr(context.Context) (int64, error)  { return m.n.Add(1), nil }
func (m *memoryStore) Count(context.Context) (int64, error) { return m.n.Load(), nil }
func (m *memoryStore) Ping(context.Context) error           { return nil }
func (m *memoryStore) Kind() string                         { return "memory" }

type postgresStore struct {
	pool *pgxpool.Pool
}

// connectPostgres retries for up to a minute, because in Compose and
// Kubernetes the database often starts after the app.
func connectPostgres(ctx context.Context, url string, log *slog.Logger) (*postgresStore, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	deadline := time.Now().Add(time.Minute)
	for attempt := 1; ; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err = pool.Ping(pingCtx)
		cancel()
		if err == nil {
			break
		}
		if time.Now().After(deadline) || ctx.Err() != nil {
			pool.Close()
			return nil, fmt.Errorf("connect to database: %w", err)
		}
		log.Warn("database not reachable yet, retrying", "attempt", attempt, "error", err)
		time.Sleep(2 * time.Second)
	}

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS visits (id int PRIMARY KEY, count bigint NOT NULL);
		INSERT INTO visits (id, count) VALUES (1, 0) ON CONFLICT (id) DO NOTHING;`)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	log.Info("connected to database")
	return &postgresStore{pool: pool}, nil
}

func (p *postgresStore) Incr(ctx context.Context) (int64, error) {
	var n int64
	err := p.pool.QueryRow(ctx, `UPDATE visits SET count = count + 1 WHERE id = 1 RETURNING count`).Scan(&n)
	return n, err
}

func (p *postgresStore) Count(ctx context.Context) (int64, error) {
	var n int64
	err := p.pool.QueryRow(ctx, `SELECT count FROM visits WHERE id = 1`).Scan(&n)
	return n, err
}

func (p *postgresStore) Ping(ctx context.Context) error { return p.pool.Ping(ctx) }
func (p *postgresStore) Kind() string                   { return "postgres" }
func (p *postgresStore) Close()                         { p.pool.Close() }
