# FraudGuard Curriculum: Go Backend → Distributed Systems

A path from Go fundamentals to distributed systems, built on one real codebase.
Each level ends with something you can show an interviewer. By the end you
will have a multi-service, event-driven fraud platform with metrics, tracing,
load-test results and design docs, which is exactly what roles like
"Backend Engineer – Real-Time Fraud Defense" ask for.

## How to work

1. Run `git init`, then commit the starting code. Do **each assignment on its
   own branch**, open a PR against yourself on GitHub, and merge it. Your
   commit history becomes part of your portfolio.
2. Every assignment has **acceptance criteria**. It isn't done until they all
   pass and `make test vet` is green.
3. Write tests **first** when you can. Interviewers notice.
4. Don't paste code you can't explain. If you use AI help, make sure you can
   rewrite the result from memory.
5. Keep a `JOURNAL.md`: one paragraph per assignment covering what you
   learned and what surprised you. It's the raw material for interview answers.

Rough time: 1–2 hours a day, about 4–6 months total. Levels 0–3 make you
employable. Levels 4–7 make you stand out.

---

## Level 0 — Go fundamentals (2–3 weeks)

**Study:** [A Tour of Go](https://go.dev/tour), then *Learning Go* (Jon
Bodner, 2nd ed.) or [Go by Example](https://gobyexample.com). Read
[Effective Go](https://go.dev/doc/effective_go) once now and again after
Level 3.

**0.1 — Read this codebase.** Trace `POST /v1/transactions` from `main.go`
to the SQL and back. Draw it on paper. Answer these in `JOURNAL.md`:
- Why does `fraud` not import `transaction`? What would break if it did?
- What does `defer` do in `migrate.go`, and why is the unlock run with
  `context.Background()`?
- `Repository` is an interface. Name two places that benefit from that.
- What happens if two requests with the same `Idempotency-Key` arrive at
  the same instant? Find the exact lines that make it safe.

**0.2 — Your first rule.** Add a `MultiCountry` rule: fire when
`DistinctCountries24h` is 3 or more (already computed, currently unused).
✅ Table-driven tests in `engine_test.go`, rule added to `DefaultEngine`.

**0.3 — Late-night rule.** Fire for transactions between 01:00 and 05:00
*in the account's home country's timezone*. You'll need to pass
`OccurredAt` into `fraud.Transaction` and map country → `time.Location`.
✅ Tests cover JP vs US accounts for the same UTC instant.

**0.4 — Go concepts drill.** Without looking anything up, write small
programs in `scratch/` (gitignored) for: slices vs arrays (append
aliasing!), maps, structs and methods, pointer vs value receivers,
interfaces, errors with `%w` / `errors.Is` / `errors.As`, goroutines,
channels, `select`, `sync.WaitGroup`, `sync.Mutex`, `context` cancellation.

---

## Level 0.5 — APIs from scratch, no framework (2–3 weeks)

Frameworks like Gin and Echo are thin layers over `net/http`. Once you've
built each layer yourself, you can use any framework, or none, and explain
what it does under the hood. Work in a separate repo, `~/http-from-scratch`,
except for 0.5.4 and 0.5.5.

**0.5.1 — HTTP over raw TCP.** Using only `net` and `bufio` (**no
`net/http`**), write a server that:
- listens with `net.Listen("tcp", ":9000")` and handles each connection in its own goroutine
- parses the request line (`POST /echo HTTP/1.1`), the headers, and a body using `Content-Length`
- writes a valid response by hand: status line, headers, blank line, body
- routes `GET /hello` → `200 hello`, `POST /echo` → echoes the body, anything else → `404`
- returns `400` for malformed requests, rejects bodies over 1 MB, and sets read deadlines (`conn.SetReadDeadline`) so a slow client can't hold a connection forever
- supports keep-alive: several requests on one connection

✅ Works with `curl -v` and with `curl --http1.1 --keepalive-time 5 url1 url2`
(both requests over one connection). In `JOURNAL.md`, explain what `net/http`
was doing for you all along.

**0.5.2 — Your own router.** Back on `net/http`, but **without** Go 1.22's
`"GET /items/{id}"` patterns: register handlers with plain
`http.Handle("/", yourRouter)`. Build a router that supports:
- method + path matching: `r.Handle("GET", "/accounts/:id/transactions", h)`
- path parameters, read in handlers with `router.Param(r, "id")` (store them in the request `context`)
- `404` for unknown paths and `405 Method Not Allowed` with an `Allow` header for a known path with the wrong method
- route groups with a prefix: `v1 := r.Group("/v1")`

✅ Table-driven tests covering static routes, params, 404 vs 405, trailing
slashes, and conflicting routes (`/a/:id` vs `/a/new`).
Then **read [chi's router source](https://github.com/go-chi/chi/blob/master/tree.go)**
and write up how their radix tree differs from your approach.

**0.5.3 — Middleware and a mini framework.** Add to your router:
- the middleware type `func(http.Handler) http.Handler` and `r.Use(mw...)`, applied in the right order
- write these from memory: request ID, logging (you'll need a `ResponseWriter` wrapper to capture the status), panic recovery, CORS, request timeout
- helpers: `BindJSON(r, &dst)` (size limit, unknown fields rejected, one object only) and `JSON(w, status, v)`
- a consistent error format and a central `func(w, r, err)` that maps errors to status codes

✅ A test proves the middleware order: recovery catches a panic, *and* the
logger still records the request as a 500.

**0.5.4 — Rebuild FraudGuard's HTTP layer (reverse engineering).**
```bash
git checkout -b rebuild-httpapi
mv internal/httpapi/server_test.go /tmp/ && rm internal/httpapi/*.go && mv /tmp/server_test.go internal/httpapi/
```
`server_test.go` is now your **specification**. Rewrite `server.go`,
`middleware.go` and `respond.go` without looking at the originals, until
`go test ./internal/httpapi/` passes. Then diff your version against `main`
(`git diff main -- internal/httpapi`) and write down every difference and
which version is better and why.

**0.5.5 — A full resource from scratch, every layer.** Add **merchants** to
FraudGuard, designing each layer yourself:
`POST /v1/merchants`, `GET /v1/merchants/{id}`, `GET /v1/merchants?category=&cursor=`,
`PATCH /v1/merchants/{id}` (partial update: how do you tell "field missing"
from "field set to empty"?), and `DELETE /v1/merchants/{id}` (soft delete).
You write the migration, model and validation, the repository (Postgres +
memory), the service, the handlers, and tests at every layer.
✅ Write the API contract (`docs/merchants-api.md`: request and response
examples, status codes, error cases) **before** writing any code.

---

## Level 1 — SQL and PostgreSQL (2–3 weeks)

**Study:** [Use The Index, Luke](https://use-the-index-luke.com) (free; the
best book on indexes), and the PostgreSQL docs chapters on *Queries*,
*Indexes* and *Concurrency Control*.

Run `make seed`, then `make psql`. Save every answer in `sql/exercises.sql`
with a comment explaining it.

**1.1 — Fundamentals**
1. The 10 largest transactions in the last 7 days, with the account's `external_ref`. (JOIN)
2. Number of transactions and total amount per merchant category. (GROUP BY)
3. Categories whose decline rate is above 25%. (HAVING, `avg((decision='decline')::int)`)
4. Accounts that have **never** had a declined transaction. (Solve it twice: `NOT EXISTS` and `LEFT JOIN … IS NULL`.)
5. Daily transaction count for the last 30 days, **including days with zero**. (`generate_series` + LEFT JOIN)

**1.2 — Window functions (the interview favorite)**
6. Each account's 3 largest transactions. (`row_number() OVER (PARTITION BY …)`)
7. For each transaction, the time since that account's previous transaction. (`lag`)
8. Each account's 7-day rolling spend per day. (`sum() OVER (… RANGE BETWEEN interval '6 days' PRECEDING AND CURRENT ROW)`)
9. Transactions more than 5x larger than the account's running average up to that point.

**1.3 — CTEs and fraud-style analytics**
10. "Impossible travel": the same account in two different countries within 2 hours.
11. Card testing: accounts with 5 or more transactions under ¥500 within any 10-minute window.
12. Recursive CTE: build an `account_links` table (shared device IDs) and find every account connected to a given one. Fraud rings are graphs.

**1.4 — Performance**
13. Run `EXPLAIN (ANALYZE, BUFFERS)` on the `History` query from `postgres.go`. Which index does it use? Paste the plan and explain each node.
14. Find a query from 1.1–1.3 that does a Seq Scan, add an index in a new migration (`0002_…sql`), and show the before/after plans and timings.
15. Show why `OFFSET 150000` is slow compared with keyset pagination on the same data.

**1.5 — Transactions and isolation.** Open two `psql` sessions and
demonstrate a lost update under `READ COMMITTED`. Then show how
`SELECT … FOR UPDATE` prevents it, and what `SERIALIZABLE` does instead.
Write it up in `docs/isolation.md`. You'll be asked about this in interviews.

---

## Level 2 — Building features properly (3–4 weeks)

**2.1 — Idempotency done right.** Right now, replaying a key with a
*different* body silently returns the old result. Store a SHA-256 hash of
the request body and return `422 idempotency_key_reused` when it differs
(this is how Stripe does it).
✅ Service tests and HTTP tests for same-body and different-body replays.

**2.2 — Merchants and blocklists.** Add a `merchants` table and a
`blocklist` table (by merchant, country, or account) with CRUD endpoints. A
blocklist hit means an instant `decline` with score 100.
✅ Migration, repository methods (Postgres + memory), handlers, tests.

**2.3 — Analyst review workflow.** Transactions with `review` need a human.
Add `POST /v1/transactions/{id}/resolve` with
`{"resolution":"confirmed_fraud"|"legitimate","note":"…"}`.
- Model it as a **state machine**: `pending_review → confirmed_fraud | legitimate`. Any other transition returns 409.
- Write an `audit_log` row **in the same DB transaction** as the state change.

✅ Concurrent resolves of the same transaction: exactly one wins.

**2.4 — Multi-currency correctness.** `HighAmount` compares minor units, but
¥500,000 and $5,000.00 are both `500000`. JPY has 0 decimal places, USD has
2, BHD has 3. Add a currency table (ISO 4217 exponents) and an FX-rate table,
and make every amount rule compare in a base currency.
✅ Tests prove JPY, USD and BHD transactions of equal value score equally.

**2.5 — Configurable rules.** Move rule thresholds into a `rule_config`
table, cached in memory and refreshed every 30 seconds (goroutine + ticker +
`sync.RWMutex` or `atomic.Pointer`). Thresholds change without a redeploy.
✅ Test the cache refresh with a fake clock. No data races (`make test-race`).

**2.6 — Authentication.** API keys per client, stored as hashes, never in
plaintext. Write the middleware, return 401/403, and add a per-client rate
limit (token bucket, `golang.org/x/time/rate`) that returns `429` with a
`Retry-After` header.

---

## Level 3 — Testing and code quality (1–2 weeks, then ongoing)

**3.1** Replace the "run `make db-up` first" integration tests with
[testcontainers-go](https://golang.testcontainers.org) so `go test` starts
its own Postgres.

**3.2** Add a **fuzz test** (`go test -fuzz`) for `DecodeCursor` and for
`NewTransaction.Validate`. Fix anything it finds.

**3.3** Add **benchmarks** for `Engine.Evaluate` (`b.Loop()`), and report
`ns/op` and `allocs/op`. Reduce allocations, then show the before and after.

**3.4** Add [golangci-lint](https://golangci-lint.run) with a sensible
config, plus a **GitHub Actions** pipeline: lint → unit tests (race) →
integration tests (Postgres service container) → Docker build. Put a green
badge in the README.

**3.5** Generate an **OpenAPI 3** spec (hand-written YAML is fine) and add a
test that fails when the routes and the spec drift apart.

---

## Level 4 — Observability (1–2 weeks)

You can't run what you can't see. Fraud teams live on dashboards.

**4.1 — Metrics.** Add a Prometheus `/metrics` endpoint with:
`http_request_duration_seconds` (histogram by route *pattern*, not raw path,
because of cardinality), `fraud_decisions_total{decision}`,
`fraud_rule_hits_total{rule}`, and DB pool stats. Add Prometheus and Grafana
to `docker-compose.yml` with a provisioned dashboard showing p50/p95/p99
latency and the decline rate.

**4.2 — Tracing.** Instrument with OpenTelemetry: one span per HTTP request,
child spans for the DB transaction and rule evaluation, and the trace ID in
every log line. Run Jaeger in compose. Screenshot a trace for the README.

**4.3 — Profiling.** Expose `net/http/pprof` on a separate *internal* port.
Under load, capture a CPU profile and a heap profile, and write up the top
three hotspots.

---

## Level 5 — Performance and concurrency (2 weeks)

**5.1 — Load test.** Use [k6](https://k6.io) or
[vegeta](https://github.com/tsenart/vegeta) to drive realistic traffic (many
accounts, a few hot ones). Record max throughput at p99 under 50 ms. This
number goes on your resume.

**5.2 — Hot-account problem.** One account with 1,000 req/s serializes on
its row lock. Measure it. Then move velocity counters into **Redis**
(sliding window with sorted sets or `INCR` + `EXPIRE` buckets, done
atomically with a Lua script) and compare. Write `docs/adr/0001-velocity-in-redis.md`
explaining the trade-off: what consistency did you give up, and why is it
acceptable for fraud?

**5.3 — Timeouts and budgets.** Give the scoring path a hard 100 ms
budget (`context.WithTimeout`). If Redis is slow, **degrade** (skip that
feature, mark the result `degraded: true`) rather than fail the payment.
Test it by injecting latency.

**5.4 — Worker pool.** Write `cmd/backfill`: re-score the last 30 days of
transactions with the current rules, using a bounded worker pool (N
goroutines, `errgroup`), batching, progress logging, and clean Ctrl-C
handling that can resume where it stopped.

---

## Level 6 — Distributed systems (4–6 weeks)

**Study:** *Designing Data-Intensive Applications* (Kleppmann), chapters 5,
7, 8, 9 and 11. It's the most valuable backend book there is.

**6.1 — Events with a transactional outbox.** Every scored transaction must
publish a `transaction.scored` event to **Kafka** (or Redpanda, which is
lighter to run). The trap: writing to the DB and then publishing to Kafka
isn't atomic. If the process crashes between the two, the event is lost.
Solve it with the **outbox pattern**: insert the event into an `outbox`
table in the same DB transaction, and have a relay goroutine publish and
mark rows sent.
✅ Kill the process mid-flight (`kill -9`) during a load test. No event is lost.

**6.2 — A second service: case management.** `cmd/cases` consumes
`transaction.scored`, opens a "case" for every `review` and `decline`, and
exposes its own API and database (database per service).
- Kafka delivers **at least once**, so make the consumer **idempotent** (dedupe table keyed by event ID).
- Failing messages go to a **dead-letter topic** after N retries with exponential backoff and jitter.

✅ Replaying the whole topic from offset 0 creates no duplicate cases.

**6.3 — gRPC.** Expose scoring over gRPC as well as HTTP (protobuf
contract in `api/proto`). Have `cases` call `fraudguard` over gRPC to fetch
transaction details, with deadlines, retries for idempotent calls only, and
a **circuit breaker** (`sony/gobreaker`).

**6.4 — Feedback loop.** When an analyst resolves a case as
`confirmed_fraud`, `cases` publishes `case.resolved`. FraudGuard consumes it
and adds the merchant/device to a "recent fraud" signal used by a new rule.
You now have an **event-driven, eventually consistent** system. Draw the
sequence diagram and explain what a user sees during the inconsistency
window.

**6.5 — Horizontal scale.** Run 3 replicas of FraudGuard behind
nginx/Envoy in compose. Prove that migrations run once (advisory lock),
that the outbox relay doesn't double-publish (`FOR UPDATE SKIP LOCKED`, so
each replica claims different rows), and that graceful shutdown during a
rolling restart drops zero requests (verify with k6).

**6.6 — Deploy to Kubernetes.** Write manifests or a Helm chart for
[kind](https://kind.sigs.k8s.io): Deployment with liveness/readiness probes
pointed at `/healthz` and `/readyz`, resource limits, HPA on CPU, ConfigMap
and Secret, a PodDisruptionBudget, and a migration Job.

---

## Level 7 — Capstone: make it hireable (1–2 weeks)

1. **README as a landing page.** Architecture diagram (excalidraw), a GIF
   of the Grafana dashboard under load, load-test numbers, and a "design
   decisions" section that links your ADRs.
2. **Three ADRs** (`docs/adr/`): outbox vs CDC (Debezium), Redis velocity vs
   DB lock, Kafka vs NATS. Each covers context, options, decision and
   consequences.
3. **A chaos write-up.** Kill Postgres, Redis and Kafka one at a time under
   load. Document what happened, what users saw, and what you fixed.
4. **A blog post** on dev.to or Zenn (popular in Japan): *"Preventing
   double-charges and velocity bypass with idempotency keys and row
   locks in Go"*. Recruiters and engineers read these.
5. **Resume bullets** (fill in your real numbers):
   - Built a real-time fraud-scoring service in Go and PostgreSQL handling
     **N req/s at p99 < X ms**, with idempotent APIs and race-free per-account
     velocity checks.
   - Designed an event-driven architecture (Kafka, transactional outbox,
     idempotent consumers, DLQ) with zero event loss under `kill -9` chaos tests.
   - Instrumented with Prometheus, Grafana and OpenTelemetry, and cut p99
     latency **Y%** via profiling and moving hot counters to Redis.

---

## Interview prep checklist

You should be able to explain each of these out loud, using this project as the example:

- [ ] Goroutines vs threads; how the scheduler works (G, M, P)
- [ ] Channels vs mutexes, and when to use each
- [ ] How `context` cancellation propagates, and why every I/O call takes one
- [ ] Interfaces: implicit satisfaction, "accept interfaces, return structs"
- [ ] Error wrapping and sentinel vs typed errors
- [ ] Indexes: B-tree, composite index column order, covering indexes
- [ ] Isolation levels and the anomalies each one allows
- [ ] Idempotency: why, how, and where it breaks
- [ ] At-least-once vs exactly-once delivery (and why "exactly-once" is mostly marketing)
- [ ] Outbox pattern, saga pattern, CQRS
- [ ] CAP and PACELC, using your system's choices as the example
- [ ] Retries: backoff, jitter, retry storms, circuit breakers
- [ ] Rate limiting algorithms: token bucket, sliding window
- [ ] Horizontal scaling and what has to be stateless
- [ ] Design a fraud system that handles 50k TPS (whiteboard version of this repo)

Good luck — 頑張って！
