# Go learning path

**Doing it intensively? Follow [BOOTCAMP.md](BOOTCAMP.md)**, a 14-day
day-by-day plan through everything below. Otherwise, do these in order at
your own pace.

Write code in the files marked "YOUR CODE GOES HERE" and check it with the tests.

| # | Folder | Topic | Check with |
|---|---|---|---|
| 0 | `00-thinking/` | How to think like a programmer: method, decomposition, edge cases, debugging, refactoring | `go test ./00-thinking/ -v -run Receipt` (then Shift, Bank, Roman, Buggy, Messy) |
| 1 | `01-basics/` | Types, if, loops, strings, slices, maps, structs, pointers, errors, interfaces | `go test ./01-basics/ -v` |
| 2 | `02-expense-cli/` | Build a CRUD app: memory → JSON → CLI flags → packages → PostgreSQL → HTTP API | the ✅ criteria in its README |
| 3 | `03-advanced/` | Interfaces in depth, generics, goroutines, channels, select, mutex, worker pools, context | `go test ./03-advanced/ -v -run Interfaces` (then `Generics`, then `Concurrency`) |
| 4 | `04-dsa/` | Data structures and algorithms: arrays, stacks, linked lists, sorting, recursion, trees, heaps, graphs, DP, LRU cache, trie | `go test ./04-dsa/ -v -run Arrays` (then `Stacks`, `LinkedList`, ...) |
| 5 | `05-clients/` | Connecting to Redis, Postgres, HTTP APIs, NATS and S3 | `go test ./05-clients/ -v -run Redis` (then Postgres, HTTP, NATS, S3) |
| 6 | `~/fraudguard` | A production backend and distributed systems | `CURRICULUM.md` |
| 7 | `~/devops-path` | DevOps: Linux, Docker, Compose, CI/CD, Kubernetes, Helm, monitoring, Terraform | `./check.sh 0` (then 2, 3, ...) |

Suggested order at your own pace: `00-thinking` (read the guide, do 1–2) →
`01-basics` → rest of `00-thinking` → `03-advanced` Parts 1–2 →
`02-expense-cli` Stages A–C → `03-advanced` Part 3 → `02-expense-cli` D–F →
`05-clients` → FraudGuard → DevOps.

DSA runs **alongside** everything else: after `01-basics`, do one or two
DSA exercises a day next to whatever project you're on.

When you're stuck for more than 20 minutes, ask for a **hint**, not the answer.
