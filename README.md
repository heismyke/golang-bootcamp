# Golang Bootcamp

A hands-on path from Go basics to production backend engineering, distributed systems and DevOps.

| Folder | What it is |
|---|---|
| [`go-basics/`](go-basics/) | Start here. Thinking like a programmer, Go basics, an expense CLI, advanced Go (interfaces, generics, concurrency), data structures and algorithms, and connecting to Redis, Postgres, NATS and S3. Includes the [2-week bootcamp plan](go-basics/BOOTCAMP.md). |
| [`fraudguard/`](fraudguard/) | A production-style Go REST API (stdlib `net/http`, PostgreSQL, Docker), with a levelled [curriculum](fraudguard/CURRICULUM.md) of features to build. |
| [`devops-path/`](devops-path/) | DevOps levels 0–10: Linux and shell, Docker, Compose, Kubernetes (kind), Helm, Prometheus/Grafana, Terraform and CI, each with an automatic checker. |

## How to use

Each exercise is a stub with tests. Write the code until the tests pass:

```bash
cd go-basics
go test ./01-basics/ -run Test01
```

Follow [`go-basics/BOOTCAMP.md`](go-basics/BOOTCAMP.md) day by day.
