// DON'T EDIT THIS FILE. These tests check your answers against real
// services. Start them first:
//
//	docker compose -f 05-clients/compose.yaml up -d --wait
package clients

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

var (
	redisAddr  = env("REDIS_ADDR", "localhost:56379")
	pgURL      = env("CLIENTS_DATABASE_URL", "postgres://clients:clients@localhost:55434/clients?sslmode=disable")
	natsURL    = env("NATS_URL", "nats://localhost:54222")
	s3Endpoint = env("S3_ENDPOINT", "localhost:59000")
)

const startHint = "Start the services: docker compose -f 05-clients/compose.yaml up -d --wait"

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func ctxT(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func uniq(prefix string) string { return prefix + "-" + strings.ToLower(rand.Text()[:10]) }

func within(t *testing.T, d time.Duration, hint string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() { defer close(done); f() }()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("too slow (over %v). %s", d, hint)
	}
}

// ═══ Redis ════════════════════════════════════════════════════

func testRedis(t *testing.T) *redis.Client {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(ctxT(t)).Err(); err != nil {
		t.Fatalf("Redis isn't reachable at %s: %v\n%s", redisAddr, err, startHint)
	}
	t.Cleanup(func() { rdb.Close() })
	return rdb
}

func TestRedis01_Connect(t *testing.T) {
	testRedis(t)
	rdb, err := NewRedis(ctxT(t), redisAddr)
	if err != nil {
		t.Fatalf("NewRedis: %v", err)
	}
	defer rdb.Close()
	if err := rdb.Ping(ctxT(t)).Err(); err != nil {
		t.Errorf("client from NewRedis can't ping: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if bad, err := NewRedis(ctx, "localhost:1"); err == nil {
		bad.Close()
		t.Error("NewRedis with a wrong address returned no error. Did you PING?")
	}
}

func TestRedis02_CacheAside(t *testing.T) {
	rdb := testRedis(t)
	ctx := ctxT(t)
	id := int(time.Now().UnixNano() % 1_000_000_000)
	key := fmt.Sprintf("user:%d", id)
	t.Cleanup(func() { rdb.Del(context.Background(), key) })

	var loads atomic.Int32
	cache := NewUserCache(rdb, time.Minute, func(_ context.Context, uid int) (User, error) {
		loads.Add(1)
		return User{ID: uid, Name: "Myke"}, nil
	})

	for i := range 3 {
		u, err := cache.Get(ctx, id)
		if err != nil {
			t.Fatalf("Get #%d: %v", i+1, err)
		}
		if u.ID != id || u.Name != "Myke" {
			t.Errorf("Get #%d = %+v", i+1, u)
		}
	}
	if n := loads.Load(); n != 1 {
		t.Errorf("load was called %d times for 3 Gets, want 1 (the rest should come from Redis)", n)
	}

	raw, err := rdb.Get(ctx, key).Result()
	var cached User
	if err != nil || json.Unmarshal([]byte(raw), &cached) != nil || cached.Name != "Myke" {
		t.Errorf("Redis key %q should hold the user as JSON, got %q (err %v)", key, raw, err)
	}
	if ttl := rdb.TTL(ctx, key).Val(); ttl <= 0 || ttl > time.Minute {
		t.Errorf("cached key TTL = %v, want between 0 and 1m", ttl)
	}

	if err := cache.Invalidate(ctx, id); err != nil {
		t.Fatalf("Invalidate: %v", err)
	}
	cache.Get(ctx, id)
	if n := loads.Load(); n != 2 {
		t.Errorf("after Invalidate, Get should load again (loads = %d, want 2)", n)
	}

	failing := NewUserCache(rdb, time.Minute, func(context.Context, int) (User, error) {
		return User{}, errors.New("database down")
	})
	badID := id + 1
	t.Cleanup(func() { rdb.Del(context.Background(), fmt.Sprintf("user:%d", badID)) })
	if _, err := failing.Get(ctx, badID); err == nil {
		t.Error("when load fails, Get should return the error")
	}
	if rdb.Exists(ctx, fmt.Sprintf("user:%d", badID)).Val() != 0 {
		t.Error("when load fails, nothing should be cached")
	}
}

func TestRedis03_RateLimit(t *testing.T) {
	rdb := testRedis(t)
	ctx := ctxT(t)
	key := uniq("ratelimit")
	t.Cleanup(func() { rdb.Del(context.Background(), key) })

	for i := 1; i <= 6; i++ {
		ok, err := Allow(ctx, rdb, key, 5, time.Second)
		if err != nil {
			t.Fatalf("Allow #%d: %v", i, err)
		}
		if want := i <= 5; ok != want {
			t.Errorf("request #%d: allowed = %v, want %v", i, ok, want)
		}
	}
	if ttl := rdb.PTTL(ctx, key).Val(); ttl <= 0 {
		t.Errorf("rate limit key has no expiry (TTL %v); it would block forever", ttl)
	}
	time.Sleep(1100 * time.Millisecond)
	if ok, _ := Allow(ctx, rdb, key, 5, time.Second); !ok {
		t.Error("after the window passed, requests should be allowed again")
	}
}

func TestRedis04_Lock(t *testing.T) {
	rdb := testRedis(t)
	ctx := ctxT(t)
	name := uniq("lock")
	t.Cleanup(func() { rdb.Del(context.Background(), name) })

	token, ok, err := AcquireLock(ctx, rdb, name, 10*time.Second)
	if err != nil || !ok || token == "" {
		t.Fatalf("first AcquireLock = (%q, %v, %v), want a token and true", token, ok, err)
	}
	if _, ok, _ := AcquireLock(ctx, rdb, name, 10*time.Second); ok {
		t.Error("second AcquireLock succeeded while the lock is held")
	}
	if released, _ := ReleaseLock(ctx, rdb, name, "wrong-token"); released {
		t.Error("ReleaseLock with someone else's token released the lock")
	}
	if rdb.Exists(ctx, name).Val() != 1 {
		t.Fatal("the lock disappeared after a release with the wrong token")
	}
	if released, err := ReleaseLock(ctx, rdb, name, token); !released || err != nil {
		t.Errorf("ReleaseLock with the right token = (%v, %v), want (true, nil)", released, err)
	}
	if _, ok, _ := AcquireLock(ctx, rdb, name, 200*time.Millisecond); !ok {
		t.Error("couldn't acquire the lock after releasing it")
	}
	time.Sleep(300 * time.Millisecond)
	if _, ok, _ := AcquireLock(ctx, rdb, name, time.Second); !ok {
		t.Error("lock didn't expire after its TTL")
	}

	// 20 goroutines race for the same lock: exactly one may win.
	race := uniq("lock-race")
	t.Cleanup(func() { rdb.Del(context.Background(), race) })
	var wins atomic.Int32
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if _, ok, _ := AcquireLock(ctx, rdb, race, 10*time.Second); ok {
				wins.Add(1)
			}
		})
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Errorf("%d goroutines got the lock, want exactly 1", wins.Load())
	}
}

func TestRedis05_Leaderboard(t *testing.T) {
	rdb := testRedis(t)
	ctx := ctxT(t)
	board := uniq("board")
	t.Cleanup(func() { rdb.Del(context.Background(), board) })

	for _, s := range []struct {
		p   string
		pts float64
	}{{"alice", 10}, {"bob", 25}, {"alice", 20}, {"carol", 5}} {
		if err := AddScore(ctx, rdb, board, s.p, s.pts); err != nil {
			t.Fatalf("AddScore: %v", err)
		}
	}
	top, err := Top(ctx, rdb, board, 2)
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	if want := []Entry{{"alice", 30}, {"bob", 25}}; !slices.Equal(top, want) {
		t.Errorf("Top(2) = %v, want %v", top, want)
	}
	if all, _ := Top(ctx, rdb, board, 10); len(all) != 3 {
		t.Errorf("Top(10) returned %d entries, want 3", len(all))
	}
}

func TestRedis06_Queue(t *testing.T) {
	rdb := testRedis(t)
	ctx := ctxT(t)
	q := uniq("queue")
	t.Cleanup(func() { rdb.Del(context.Background(), q) })

	for _, job := range []string{"a", "b", "c"} {
		if err := Enqueue(ctx, rdb, q, job); err != nil {
			t.Fatalf("Enqueue: %v", err)
		}
	}
	for _, want := range []string{"a", "b", "c"} {
		job, ok, err := Dequeue(ctx, rdb, q, time.Second)
		if err != nil || !ok || job != want {
			t.Errorf("Dequeue = (%q, %v, %v), want (%q, true, nil): oldest job first", job, ok, err, want)
		}
	}
	start := time.Now()
	job, ok, err := Dequeue(ctx, rdb, q, time.Second)
	if err != nil || ok {
		t.Errorf("Dequeue on an empty queue = (%q, %v, %v), want (\"\", false, nil)", job, ok, err)
	}
	if time.Since(start) < 900*time.Millisecond {
		t.Error("Dequeue on an empty queue returned immediately; it should wait for the timeout (BRPOP)")
	}
}

// ═══ Postgres ═════════════════════════════════════════════════

const testSchema = `
CREATE TABLE IF NOT EXISTS bank_accounts (id text PRIMARY KEY, balance bigint NOT NULL CHECK (balance >= 0));
CREATE TABLE IF NOT EXISTS events (id bigserial PRIMARY KEY, kind text NOT NULL, payload text NOT NULL,
                                   created_at timestamptz NOT NULL DEFAULT now());`

func testPG(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := ctxT(t)
	pool, err := pgxpool.New(ctx, pgURL)
	if err == nil {
		err = pool.Ping(ctx)
	}
	if err != nil {
		t.Fatalf("Postgres isn't reachable: %v\n%s", err, startHint)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(ctx, testSchema); err != nil {
		t.Fatal(err)
	}
	return pool
}

func TestPostgres07_Connect(t *testing.T) {
	testPG(t)
	pool, err := NewPostgres(ctxT(t), pgURL)
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctxT(t)); err != nil {
		t.Errorf("pool can't ping: %v", err)
	}
	if got := pool.Config().MaxConns; got != 10 {
		t.Errorf("MaxConns = %d, want 10", got)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if bad, err := NewPostgres(ctx, "postgres://x:y@localhost:1/z?sslmode=disable"); err == nil {
		bad.Close()
		t.Error("NewPostgres with a wrong address returned no error. Did you ping?")
	}
}

func TestPostgres08_Migrate(t *testing.T) {
	admin := testPG(t)
	ctx := ctxT(t)
	// Run Migrate inside an empty schema, so it must really create the tables.
	schema := strings.ReplaceAll(uniq("test"), "-", "_")
	admin.Exec(ctx, "CREATE SCHEMA "+schema)
	t.Cleanup(func() { admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE") })

	cfg, _ := pgxpool.ParseConfig(pgURL)
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	for i := range 2 {
		if err := Migrate(ctx, pool); err != nil {
			t.Fatalf("Migrate run #%d: %v (it must be safe to run twice: IF NOT EXISTS)", i+1, err)
		}
	}
	var tables int
	pool.QueryRow(ctx, `SELECT count(*) FROM information_schema.tables
		WHERE table_schema = $1 AND table_name IN ('bank_accounts', 'events')`, schema).Scan(&tables)
	if tables != 2 {
		t.Fatalf("found %d of the 2 tables after Migrate", tables)
	}
	if _, err := pool.Exec(ctx, "INSERT INTO bank_accounts (id, balance) VALUES ('x', -1)"); err == nil {
		t.Error("a negative balance was accepted; add CHECK (balance >= 0)")
	}
	if _, err := pool.Exec(ctx, "INSERT INTO events (kind, payload) VALUES ('k', 'p')"); err != nil {
		t.Errorf("events insert with defaults failed: %v (id and created_at need defaults)", err)
	}
}

func TestPostgres09_Accounts(t *testing.T) {
	pool := testPG(t)
	ctx := ctxT(t)
	id := uniq("acct")

	if err := CreateAccount(ctx, pool, id, 500); err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if err := CreateAccount(ctx, pool, id, 1); !errors.Is(err, ErrAccountExists) {
		t.Errorf("duplicate CreateAccount = %v, want ErrAccountExists", err)
	}
	if b, err := Balance(ctx, pool, id); b != 500 || err != nil {
		t.Errorf("Balance = (%d, %v), want (500, nil)", b, err)
	}
	if _, err := Balance(ctx, pool, "missing-"+id); !errors.Is(err, ErrNoAccount) {
		t.Errorf("Balance(missing) error = %v, want ErrNoAccount", err)
	}
	evil := "x'; DROP TABLE bank_accounts; --"
	CreateAccount(ctx, pool, evil, 1)
	if _, err := pool.Exec(ctx, "SELECT 1 FROM bank_accounts LIMIT 1"); err != nil {
		t.Fatalf("bank_accounts is gone: SQL injection! Use $1 placeholders. (%v)", err)
	}
	pool.Exec(ctx, "DELETE FROM bank_accounts WHERE id = $1", evil)
}

func TestPostgres10_Transfer(t *testing.T) {
	pool := testPG(t)
	ctx := ctxT(t)
	a, b := uniq("a"), uniq("b")
	for _, id := range []string{a, b} {
		if _, err := pool.Exec(ctx, "INSERT INTO bank_accounts (id, balance) VALUES ($1, 1000)", id); err != nil {
			t.Fatal(err)
		}
	}
	bal := func(id string) int64 {
		var n int64
		pool.QueryRow(ctx, "SELECT balance FROM bank_accounts WHERE id = $1", id).Scan(&n)
		return n
	}

	if err := Transfer(ctx, pool, a, b, 300); err != nil {
		t.Fatalf("Transfer: %v", err)
	}
	if bal(a) != 700 || bal(b) != 1300 {
		t.Fatalf("after transferring 300: a=%d b=%d, want 700 and 1300", bal(a), bal(b))
	}
	cases := []struct {
		from, to string
		amount   int64
		want     error
	}{
		{a, b, 701, ErrInsufficientFunds},
		{a, b, 0, ErrInvalidAmount},
		{a, b, -5, ErrInvalidAmount},
		{a, "missing", 1, ErrNoAccount},
		{"missing", a, 1, ErrNoAccount},
	}
	for _, c := range cases {
		if err := Transfer(ctx, pool, c.from, c.to, c.amount); !errors.Is(err, c.want) {
			t.Errorf("Transfer(%s, %s, %d) = %v, want %v", c.from, c.to, c.amount, err, c.want)
		}
	}
	if bal(a) != 700 || bal(b) != 1300 {
		t.Fatalf("failed transfers changed balances: a=%d b=%d", bal(a), bal(b))
	}

	within(t, 30*time.Second, "transfers hang. Deadlock? Lock rows in a consistent order.", func() {
		var wg sync.WaitGroup
		var failures atomic.Int32
		var firstErr atomic.Value
		for i := range 100 {
			wg.Go(func() {
				from, to := a, b
				if i%2 == 1 {
					from, to = b, a
				}
				if err := Transfer(ctx, pool, from, to, 10); err != nil {
					failures.Add(1)
					firstErr.CompareAndSwap(nil, err.Error())
				}
			})
		}
		wg.Wait()
		if n := failures.Load(); n > 0 {
			t.Errorf("%d of 100 concurrent transfers failed, e.g.: %v", n, firstErr.Load())
		}
	})
	if bal(a)+bal(b) != 2000 {
		t.Errorf("money was created or destroyed: a=%d b=%d (total should be 2000)", bal(a), bal(b))
	}
	if bal(a) != 700 {
		t.Errorf("after 50 transfers each way, a = %d, want 700", bal(a))
	}
}

func TestPostgres11_BulkInsert(t *testing.T) {
	pool := testPG(t)
	ctx := ctxT(t)
	kind := uniq("bulk")
	t.Cleanup(func() { pool.Exec(context.Background(), "DELETE FROM events WHERE kind = $1", kind) })

	events := make([]Event, 50_000)
	for i := range events {
		events[i] = Event{Kind: kind, Payload: fmt.Sprintf(`{"n":%d}`, i), CreatedAt: time.Now()}
	}
	within(t, 3*time.Second, "use pool.CopyFrom instead of one INSERT per row.", func() {
		n, err := BulkInsertEvents(ctx, pool, events)
		if err != nil || n != 50_000 {
			t.Errorf("BulkInsertEvents = (%d, %v), want (50000, nil)", n, err)
		}
	})
	var count int
	pool.QueryRow(ctx, "SELECT count(*) FROM events WHERE kind = $1", kind).Scan(&count)
	if count != 50_000 {
		t.Errorf("found %d rows in events, want 50000", count)
	}
}

// ═══ HTTP client (no services needed) ═════════════════════════

func weatherServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request, attempt int)) (*httptest.Server, *atomic.Int32) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, int(attempts.Add(1)))
	}))
	t.Cleanup(srv.Close)
	return srv, &attempts
}

func okWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"city":%q,"temp_c":21.5}`, r.URL.Query().Get("city"))
}

func client(url string) *WeatherClient {
	return &WeatherClient{BaseURL: url, HTTP: &http.Client{Timeout: time.Second}, MaxRetries: 3, Backoff: 10 * time.Millisecond}
}

func TestHTTP12_Success(t *testing.T) {
	srv, _ := weatherServer(t, func(w http.ResponseWriter, r *http.Request, _ int) {
		if r.URL.Path != "/weather" || r.Method != http.MethodGet {
			http.Error(w, `{"error":"wrong path or method"}`, 400)
			return
		}
		okWeather(w, r)
	})
	got, err := client(srv.URL).Get(context.Background(), "São Paulo")
	if err != nil || got.City != "São Paulo" || got.TempC != 21.5 {
		t.Errorf("Get = (%+v, %v), want city São Paulo (URL-encoded correctly) and 21.5", got, err)
	}
}

func TestHTTP12_RetriesServerErrors(t *testing.T) {
	srv, attempts := weatherServer(t, func(w http.ResponseWriter, r *http.Request, n int) {
		if n <= 2 {
			http.Error(w, `{"error":"overloaded"}`, http.StatusServiceUnavailable)
			return
		}
		okWeather(w, r)
	})
	start := time.Now()
	got, err := client(srv.URL).Get(context.Background(), "Tokyo")
	if err != nil || got.City != "Tokyo" {
		t.Fatalf("Get = (%+v, %v), want success after retries", got, err)
	}
	if n := attempts.Load(); n != 3 {
		t.Errorf("made %d attempts, want 3 (two 503s, then success)", n)
	}
	if time.Since(start) < 30*time.Millisecond {
		t.Error("retried without waiting; back off 10ms, then 20ms")
	}
}

func TestHTTP12_NoRetryOnClientError(t *testing.T) {
	srv, attempts := weatherServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"unknown city"}`)
	})
	_, err := client(srv.URL).Get(context.Background(), "Atlantis")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 || apiErr.Message != "unknown city" {
		t.Errorf("error = %v, want *APIError{404, \"unknown city\"}", err)
	}
	if n := attempts.Load(); n != 1 {
		t.Errorf("made %d attempts for a 404, want 1 (don't retry client errors)", n)
	}
}

func TestHTTP12_GivesUp(t *testing.T) {
	srv, attempts := weatherServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		http.Error(w, `{"error":"broken"}`, http.StatusInternalServerError)
	})
	c := client(srv.URL)
	c.MaxRetries = 2
	if _, err := c.Get(context.Background(), "Tokyo"); err == nil {
		t.Error("Get succeeded against a server that always fails")
	}
	if n := attempts.Load(); n != 3 {
		t.Errorf("made %d attempts with MaxRetries=2, want 3", n)
	}
}

func TestHTTP12_RetryAfter(t *testing.T) {
	srv, _ := weatherServer(t, func(w http.ResponseWriter, r *http.Request, n int) {
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"error":"slow down"}`)
			return
		}
		okWeather(w, r)
	})
	start := time.Now()
	if _, err := client(srv.URL).Get(context.Background(), "Tokyo"); err != nil {
		t.Fatalf("Get after a 429: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 900*time.Millisecond {
		t.Errorf("retried after %v; the server said Retry-After: 1 (second)", elapsed)
	}
}

func TestHTTP12_TimeoutIsRetried(t *testing.T) {
	srv, attempts := weatherServer(t, func(w http.ResponseWriter, r *http.Request, n int) {
		if n == 1 {
			time.Sleep(500 * time.Millisecond)
		}
		okWeather(w, r)
	})
	c := client(srv.URL)
	c.HTTP = &http.Client{Timeout: 100 * time.Millisecond}
	if _, err := c.Get(context.Background(), "Tokyo"); err != nil {
		t.Fatalf("Get: %v (a timeout should be retried)", err)
	}
	if n := attempts.Load(); n != 2 {
		t.Errorf("made %d attempts, want 2", n)
	}
}

func TestHTTP12_Cancel(t *testing.T) {
	srv, _ := weatherServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		http.Error(w, `{"error":"down"}`, http.StatusServiceUnavailable)
	})
	c := client(srv.URL)
	c.Backoff = 5 * time.Second
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(100*time.Millisecond, cancel)
	start := time.Now()
	_, err := c.Get(ctx, "Tokyo")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("took %v to notice the cancellation; stop waiting when ctx is done", elapsed)
	}
}

func TestHTTP12_BadJSON(t *testing.T) {
	srv, attempts := weatherServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		fmt.Fprint(w, "<html>not json</html>")
	})
	_, err := client(srv.URL).Get(context.Background(), "Tokyo")
	if attempts.Load() == 0 {
		t.Fatal("Get never called the API")
	}
	if err == nil {
		t.Error("Get returned no error for a response that isn't JSON")
	}
}

// ═══ NATS ═════════════════════════════════════════════════════

func testNATS(t *testing.T) *nats.Conn {
	t.Helper()
	nc, err := nats.Connect(natsURL, nats.Timeout(3*time.Second))
	if err != nil {
		t.Fatalf("NATS isn't reachable at %s: %v\n%s", natsURL, err, startHint)
	}
	t.Cleanup(nc.Close)
	return nc
}

func TestNATS13_Connect(t *testing.T) {
	testNATS(t)
	nc, err := NewNATS(natsURL)
	if err != nil {
		t.Fatalf("NewNATS: %v", err)
	}
	defer nc.Close()
	if nc.Opts.Name != "go-basics" {
		t.Errorf("connection name = %q, want \"go-basics\"", nc.Opts.Name)
	}
	if bad, err := NewNATS("nats://localhost:1"); err == nil {
		bad.Close()
		t.Error("NewNATS with a wrong address returned no error")
	}
}

func TestNATS14_PubSub(t *testing.T) {
	nc := testNATS(t)
	got := make(chan Order, 10)
	sub, err := SubscribeOrders(nc, func(o Order) { got <- o })
	if err != nil {
		t.Fatalf("SubscribeOrders: %v", err)
	}
	t.Cleanup(func() { sub.Unsubscribe() })

	raw, _ := nc.SubscribeSync("orders.created")
	t.Cleanup(func() { raw.Unsubscribe() })

	nc.Publish("orders.created", []byte("this is not json"))
	want := []Order{{"o1", 1200}, {"o2", 450}, {"o3", 99}}
	for _, o := range want {
		if err := PublishOrder(nc, o); err != nil {
			t.Fatalf("PublishOrder: %v", err)
		}
	}
	nc.Flush()

	var received []Order
	timeout := time.After(3 * time.Second)
	for len(received) < len(want) {
		select {
		case o := <-got:
			received = append(received, o)
		case <-timeout:
			t.Fatalf("received %d of %d orders within 3s: %v", len(received), len(want), received)
		}
	}
	if !slices.Equal(received, want) {
		t.Errorf("received %v, want %v", received, want)
	}
	raw.NextMsg(time.Second) // the garbage message
	if msg, err := raw.NextMsg(time.Second); err != nil || !strings.Contains(string(msg.Data), `"amount_yen":1200`) {
		t.Errorf("PublishOrder should send JSON like {\"id\":\"o1\",\"amount_yen\":1200}, got %v", msg)
	}
}

func TestNATS15_RequestReply(t *testing.T) {
	nc := testNATS(t)
	sub, err := ServePrices(nc, map[string]int{"coffee": 400, "cake": 650})
	if err != nil {
		t.Fatalf("ServePrices: %v", err)
	}
	nc.Flush()
	if p, err := GetPrice(nc, "coffee", time.Second); p != 400 || err != nil {
		t.Errorf("GetPrice(coffee) = (%d, %v), want (400, nil)", p, err)
	}
	if _, err := GetPrice(nc, "tea", time.Second); !errors.Is(err, ErrUnknownProduct) {
		t.Errorf("GetPrice(tea) error = %v, want ErrUnknownProduct", err)
	}
	sub.Unsubscribe()
	nc.Flush()
	if _, err := GetPrice(nc, "coffee", 500*time.Millisecond); err == nil {
		t.Error("GetPrice with no responder returned no error")
	}
}

func TestNATS16_QueueGroup(t *testing.T) {
	nc := testNATS(t)
	var mu sync.Mutex
	perJob := map[string]int{}
	perWorker := map[int]int{}
	subs, err := StartWorkers(nc, 3, func(worker int, job string) {
		mu.Lock()
		defer mu.Unlock()
		perJob[job]++
		perWorker[worker]++
	})
	if err != nil {
		t.Fatalf("StartWorkers: %v", err)
	}
	t.Cleanup(func() {
		for _, s := range subs {
			s.Unsubscribe()
		}
	})
	if len(subs) != 3 {
		t.Fatalf("StartWorkers returned %d subscriptions, want 3", len(subs))
	}
	nc.Flush()
	for i := range 300 {
		nc.Publish("jobs", fmt.Appendf(nil, "job-%d", i))
	}
	nc.Flush()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(perJob)
		mu.Unlock()
		if n == 300 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(perJob) != 300 {
		t.Fatalf("%d of 300 jobs were handled", len(perJob))
	}
	for job, n := range perJob {
		if n != 1 {
			t.Fatalf("%s was handled %d times; in a queue group each message goes to ONE member", job, n)
		}
	}
	if len(perWorker) < 2 {
		t.Errorf("only %d worker(s) got jobs, want the load spread across workers", len(perWorker))
	}
}

// ═══ S3 ═══════════════════════════════════════════════════════

func testS3(t *testing.T) (*minio.Client, string) {
	t.Helper()
	ctx := ctxT(t)
	c, err := minio.New(s3Endpoint, &minio.Options{Creds: credentials.NewStaticV4("s3admin", "s3secret", "")})
	if err == nil {
		_, err = c.ListBuckets(ctx)
	}
	if err != nil {
		t.Fatalf("The S3 server isn't reachable at %s: %v\n%s", s3Endpoint, err, startHint)
	}
	bucket := uniq("test")
	if err := c.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx := context.Background()
		for obj := range c.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: true}) {
			c.RemoveObject(ctx, bucket, obj.Key, minio.RemoveObjectOptions{})
		}
		c.RemoveBucket(ctx, bucket)
	})
	return c, bucket
}

func TestS3_17_Connect(t *testing.T) {
	testS3(t)
	c, err := NewS3(ctxT(t), s3Endpoint, "s3admin", "s3secret")
	if err != nil || c == nil {
		t.Fatalf("NewS3: %v", err)
	}
	if _, err := NewS3(ctxT(t), s3Endpoint, "s3admin", "wrong-password"); err == nil {
		t.Error("NewS3 with a wrong password returned no error. Did you check with ListBuckets?")
	}
}

func TestS3_18_EnsureBucket(t *testing.T) {
	c, _ := testS3(t)
	ctx := ctxT(t)
	bucket := uniq("ensure")
	t.Cleanup(func() { c.RemoveBucket(context.Background(), bucket) })
	for i := range 2 {
		if err := EnsureBucket(ctx, c, bucket); err != nil {
			t.Fatalf("EnsureBucket call #%d: %v (it must be safe to call twice)", i+1, err)
		}
	}
	if ok, _ := c.BucketExists(ctx, bucket); !ok {
		t.Error("bucket doesn't exist after EnsureBucket")
	}
}

func TestS3_19_UploadDownload(t *testing.T) {
	c, bucket := testS3(t)
	ctx := ctxT(t)
	data := []byte("hello from go-basics")
	if err := Upload(ctx, c, bucket, "docs/hello.txt", data, "text/plain"); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	info, err := c.StatObject(ctx, bucket, "docs/hello.txt", minio.StatObjectOptions{})
	if err != nil || info.ContentType != "text/plain" {
		t.Errorf("uploaded object: content type %q (err %v), want text/plain", info.ContentType, err)
	}
	got, err := Download(ctx, c, bucket, "docs/hello.txt")
	if err != nil || string(got) != string(data) {
		t.Errorf("Download = (%q, %v), want %q", got, err, data)
	}
	if _, err := Download(ctx, c, bucket, "docs/missing.txt"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Download(missing) error = %v, want ErrNotFound", err)
	}
}

func TestS3_20_PresignAndList(t *testing.T) {
	c, bucket := testS3(t)
	ctx := ctxT(t)
	for _, key := range []string{"reports/2026/b.csv", "reports/2026/a.csv", "reports/2025/z.csv", "images/logo.png"} {
		c.PutObject(ctx, bucket, key, strings.NewReader("data:"+key), -1, minio.PutObjectOptions{})
	}
	keys, err := ListKeys(ctx, c, bucket, "reports/")
	want := []string{"reports/2025/z.csv", "reports/2026/a.csv", "reports/2026/b.csv"}
	if err != nil || !slices.Equal(keys, want) {
		t.Errorf("ListKeys(reports/) = (%v, %v), want %v", keys, err, want)
	}

	url, err := PresignedURL(ctx, c, bucket, "images/logo.png", time.Minute)
	if err != nil || url == "" {
		t.Fatalf("PresignedURL: %v", err)
	}
	resp, err := http.Get(url) // no credentials: the signature in the URL is enough
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || string(body) != "data:images/logo.png" {
		t.Errorf("downloading the presigned URL gave %d %q", resp.StatusCode, body)
	}
}
