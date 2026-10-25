# Assignment 5: Connecting to real services

Almost every backend job is "receive a request, talk to some other
systems, respond." This assignment teaches you to talk to the five kinds of
systems you'll meet everywhere: a **cache** (Redis), a **database**
(Postgres), **other HTTP APIs**, a **message broker** (NATS), and **object
storage** (S3). 20 exercises, tested against real servers running in Docker.

## Setup

```bash
cd ~/go-basics
docker compose -f 05-clients/compose.yaml up -d --wait     # start everything
docker compose -f 05-clients/compose.yaml ps               # check it's healthy
go test ./05-clients/ -v -run Redis                         # one part at a time
docker compose -f 05-clients/compose.yaml down -v          # stop and wipe data
```

| Service | Address | Explore it by hand |
|---|---|---|
| Redis | `localhost:56379` | `docker compose -f 05-clients/compose.yaml exec redis redis-cli` |
| Postgres | `localhost:55434` (user/pass/db: `clients`) | `docker compose -f 05-clients/compose.yaml exec postgres psql -U clients` |
| NATS | `localhost:54222` | monitoring UI: http://localhost:58222 |
| S3 (SeaweedFS) | `localhost:59000` (key `s3admin`, secret `s3secret`) | any S3 tool, e.g. `aws s3 --endpoint-url http://localhost:59000 ls` |

**Before writing Go for a service, try its commands by hand.** For example, in
`redis-cli`: `SET greeting hello EX 10`, `GET greeting`, `TTL greeting`,
`INCR counter`, `LPUSH q a b`, `BRPOP q 5`, `ZADD board 10 alice`. Once you
know what the commands do, the Go client is just a way to send them.

## Parts

| Part | File | Exercises | You learn |
|---|---|---|---|
| 1. Redis | `redis.go` | 1–6 | connecting, cache-aside, rate limiting, distributed locks (Lua), sorted sets, queues |
| 2. Postgres | `postgres.go` | 7–11 | connection pools, migrations, error mapping, SQL injection, transactions, **deadlocks**, bulk COPY |
| 3. HTTP APIs | `httpclient.go` | 12 | timeouts, retries with exponential backoff, `Retry-After`, cancellation, error types |
| 4. NATS | `nats.go` | 13–16 | publish/subscribe, request/reply, queue groups (load balancing) |
| 5. S3 | `s3.go` | 17–20 | buckets, upload/download, lazy errors, presigned URLs, listing |

The HTTP part needs no Docker (the tests start fake servers), so you can do it anywhere.

## Rules that apply to every client

1. **Create clients once, at startup, and share them.** Redis clients, pgx
   pools, NATS connections and S3 clients are all safe for concurrent use.
   Never create one per request.
2. **Check the connection at startup** (ping), so misconfiguration fails fast.
3. **Pass `context.Context` to every call**, so requests can time out and be cancelled.
4. **Set timeouts.** A call with no timeout can hang forever and take your service down with it.
5. **Close what you open** (`defer rows.Close()`, `resp.Body.Close()`, `client.Close()` on shutdown).
6. **Map library errors to your own** (`redis.Nil` → "not cached",
   `pgx.ErrNoRows` → `ErrNotFound`), so the rest of your code doesn't depend on
   a specific library.

## Which one do I use?

| Need | Use | Why not the others |
|---|---|---|
| The source of truth for business data | **Postgres** | Redis can lose data; S3 can't query |
| Make repeated reads fast | **Redis** cache in front of Postgres | |
| Count things across replicas (rate limits) | **Redis** `INCR` | atomic and shared by all replicas |
| Only one replica runs a job | **Redis** lock (or a Postgres advisory lock) | |
| Tell other services something happened | **NATS** / Kafka | the sender shouldn't wait for, or know, the listeners |
| Spread work over many workers | **NATS** queue group / Redis list / Kafka | |
| Store files | **S3** | databases get huge and slow with blobs |
| Get data from another company | **HTTP client** with retries | |

## After this

Use what you learned in FraudGuard: Redis for velocity counters (Curriculum
5.2) and rate limiting (2.6), and NATS or Kafka for events (6.1–6.2).
