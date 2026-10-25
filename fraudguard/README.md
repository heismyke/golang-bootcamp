# FraudGuard

A real-time transaction fraud-scoring service in Go. Clients send a payment,
FraudGuard scores it against the account's history in a few milliseconds, and
returns `approve`, `review`, or `decline` with human-readable reasons.

```
POST /v1/transactions   Idempotency-Key: order-8812
{"account_id":"…","amount_minor":900000,"currency":"JPY","merchant_id":"m_9",
 "merchant_category":"crypto","country":"US"}

201 Created
{"id":"…","risk_score":90,"decision":"decline","reasons":[
  "high_amount: amount 900000 exceeds 500000",
  "country_mismatch: country US differs from home country JP",
  "high_risk_merchant: merchant category \"crypto\" is high risk", …]}
```

## Quick start

Requirements: Go 1.27+, Docker.

```bash
make db-up              # Postgres on localhost:55432
make run                # API on localhost:8088 (runs migrations on start)
make test               # unit tests, no database needed
make test-integration   # adds Postgres concurrency tests
make seed               # ~200k transactions for SQL practice
make psql               # SQL shell
make up                 # the whole stack in Docker
```

## API

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/accounts` | Create an account (`external_ref`, `home_country`) |
| `GET` | `/v1/accounts/{id}` | Fetch an account |
| `GET` | `/v1/accounts/{id}/transactions?limit=&cursor=` | Newest-first, keyset-paginated |
| `POST` | `/v1/transactions` | Score and record a transaction. Requires `Idempotency-Key` |
| `GET` | `/v1/transactions/{id}` | Fetch a transaction |
| `GET` | `/healthz`, `/readyz` | Liveness and readiness probes |

Errors share one shape:
`{"error":{"code":"validation_failed","message":"…","fields":{…},"request_id":"…"}}`

## Architecture

```
cmd/api                  wiring: config → db → migrations → service → HTTP server
internal/
  config                 env-based configuration
  httpapi                routing, JSON, middleware (request ID, logging, panic recovery)
  transaction            use cases (Service), storage port (Repository),
                         Postgres + in-memory adapters
  fraud                  pure scoring engine: rules → score → decision
  platform/postgres      connection pool, embedded migrations
```

Dependencies point inward: `httpapi → transaction → fraud`. The fraud engine
knows nothing about HTTP or SQL, and the service knows nothing about HTTP.

## Design decisions

- **Idempotency keys.** Payment clients retry on timeouts. Without
  idempotency, a retry would score and store the payment twice. A replay
  returns the original result (`200` + `Idempotent-Replayed: true`).
- **Row lock per account.** `SELECT … FOR UPDATE` on the account serializes
  concurrent payments for that account, so a burst of parallel requests can't
  all read the same stale history and dodge the velocity rule. The integration
  test `TestPostgresVelocityUnderConcurrency` proves it.
- **Money as integer minor units.** `amount_minor bigint`, never floats.
- **Keyset pagination.** `(occurred_at, id) < ($cursor)` uses the index
  directly. `OFFSET` gets slower with every page.
- **One aggregate query for features.** `FILTER` clauses compute every
  history window in a single index range scan.
- **Graceful shutdown.** On SIGTERM the server stops accepting connections
  and drains in-flight requests before exiting.
- **Distroless, non-root container** with a static binary.

See [CURRICULUM.md](CURRICULUM.md) for the learning path built on this codebase.
