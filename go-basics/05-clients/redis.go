// Package clients: connecting Go to the services every backend talks to.
// YOUR CODE GOES IN redis.go, postgres.go, httpclient.go, nats.go and s3.go.
//
// Start the services first (from the go-basics folder):
//
//	docker compose -f 05-clients/compose.yaml up -d --wait
//
// Then test one part at a time:
//
//	go test ./05-clients/ -v -run Redis
//
// Parts: Redis, Postgres, HTTP, NATS, S3. Read 05-clients/README.md first.
package clients

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrTODO is returned by functions you haven't written yet.
var ErrTODO = errors.New("TODO: not implemented")

// ═════════════════════════════════════════════════════════════
// PART 1: REDIS
// An in-memory key-value store, used for caching, rate limiting, locks,
// queues and leaderboards. Docs: https://redis.io/docs/latest/commands/
// and https://redis.uptrace.dev/guide/ (the go-redis client).
// Explore by hand first: docker compose -f 05-clients/compose.yaml exec redis redis-cli
// ═════════════════════════════════════════════════════════════

// ─── 1. Connecting ────────────────────────────────────────────
// NewRedis creates a client for addr ("localhost:56379") and checks the
// connection with PING. Creating a client doesn't connect: always ping, so
// a wrong address fails at startup, not on the first user request.
func NewRedis(ctx context.Context, addr string) (*redis.Client, error) {
	return nil, ErrTODO
}

// ─── 2. Cache-aside ───────────────────────────────────────────
// The most common caching pattern:
//   1. look in the cache
//   2. hit → return it
//   3. miss → load from the slow source (database, API), store it in the
//      cache with a TTL, return it

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UserCache struct {
	rdb  *redis.Client
	ttl  time.Duration
	load func(ctx context.Context, id int) (User, error) // the slow source
}

func NewUserCache(rdb *redis.Client, ttl time.Duration, load func(context.Context, int) (User, error)) *UserCache {
	return &UserCache{rdb: rdb, ttl: ttl, load: load}
}

// Get returns the user, from Redis if cached under key "user:<id>" (stored
// as JSON), otherwise from load, caching the result with the TTL.
// A cache miss is the error redis.Nil; any OTHER Redis error should be
// returned, not treated as a miss. If load fails, return its error and
// cache nothing.
func (c *UserCache) Get(ctx context.Context, id int) (User, error) {
	return User{}, ErrTODO
}

// Invalidate removes the user from the cache (call it after an update so
// readers don't see stale data).
func (c *UserCache) Invalidate(ctx context.Context, id int) error {
	return ErrTODO
}

// ─── 3. Rate limiting ─────────────────────────────────────────
// Allow reports whether one more request is allowed for key, using a fixed
// window: at most `limit` requests per `window`. Every replica of your API
// shares the same Redis, so the limit holds across all of them.
//
// INCR the key; if the result is 1 (first request in this window), set
// EXPIRE to the window. Allowed if the count is <= limit.
// Bonus question: what goes wrong if the process crashes between INCR and
// EXPIRE? Fix it with a pipeline in a MULTI transaction (rdb.TxPipelined).
func Allow(ctx context.Context, rdb *redis.Client, key string, limit int, window time.Duration) (bool, error) {
	return false, ErrTODO
}

// ─── 4. Distributed lock ──────────────────────────────────────
// Only one process (out of many replicas) should run a job at a time, e.g.
// a nightly report. A Redis lock does that.

// AcquireLock tries to take lock `name` for ttl. It returns a random token
// and true if it got the lock, or false if someone else holds it.
// Use SET with NX (only if not exists) and an expiry, in ONE command
// (rdb.SetNX). The TTL means a crashed holder can't block everyone forever.
// Token: crypto/rand.Text().
func AcquireLock(ctx context.Context, rdb *redis.Client, name string, ttl time.Duration) (token string, ok bool, err error) {
	return "", false, ErrTODO
}

// ReleaseLock deletes the lock only if it still holds OUR token (the lock
// may have expired and been taken by someone else). "Check the value, then
// delete" must be atomic, so it has to be a Lua script:
//
//	if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("DEL", KEYS[1]) else return 0 end
//
// Returns true if we released it. Hint: redis.NewScript(...).Run(...)
func ReleaseLock(ctx context.Context, rdb *redis.Client, name, token string) (bool, error) {
	return false, ErrTODO
}

// ─── 5. Leaderboard (sorted sets) ─────────────────────────────

type Entry struct {
	Player string
	Score  float64
}

// AddScore adds points to a player's score on a board (ZINCRBY).
func AddScore(ctx context.Context, rdb *redis.Client, board, player string, points float64) error {
	return ErrTODO
}

// Top returns the n highest scores, highest first (ZREVRANGE ... WITHSCORES,
// or rdb.ZRevRangeWithScores).
func Top(ctx context.Context, rdb *redis.Client, board string, n int) ([]Entry, error) {
	return nil, ErrTODO
}

// ─── 6. Job queue (lists) ─────────────────────────────────────

// Enqueue adds a job to the queue (LPUSH).
func Enqueue(ctx context.Context, rdb *redis.Client, queue, job string) error {
	return ErrTODO
}

// Dequeue waits up to timeout for the oldest job (BRPOP). It returns
// ok=false (and no error) if the timeout passes with nothing to do.
// Hint: on timeout, BRPOP returns redis.Nil.
func Dequeue(ctx context.Context, rdb *redis.Client, queue string, timeout time.Duration) (job string, ok bool, err error) {
	return "", false, ErrTODO
}
