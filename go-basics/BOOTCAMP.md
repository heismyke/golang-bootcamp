# 14-day backend bootcamp

Two weeks, back to back, about 10 focused hours a day. It covers the core
of everything in this folder, FraudGuard and the DevOps path, in an order
where each day builds on the one before.

**Be realistic:** two weeks gives you a solid, hands-on foundation and a
portfolio project, not mastery. The full curriculum takes months. The
bootcamp picks the parts that matter most and leaves the rest for after
(listed at the end).

---

## The daily routine

| Time | Block | What |
|---|---|---|
| 08:30–09:30 | **DSA warm-up** | 2 exercises from `04-dsa` (method: paper first, then code) |
| 09:30–12:30 | **Block 1** | the day's main topic |
| 12:30–13:30 | lunch, away from the screen | |
| 13:30–17:30 | **Block 2** | exercises and building |
| 17:30–18:30 | break, walk | |
| 18:30–20:00 | **Block 3** | finish, stretch goals, or catch up |
| 20:00–20:30 | **Review** | run the day's checks, commit, write `JOURNAL.md`, send me your results |

**Rules**
1. **Type every line yourself.** No copy-pasting solutions, no AI-written code.
   Ask me for hints, explanations and code reviews instead.
2. **Use the 5-step method** (`00-thinking/GUIDE.md`) for every exercise.
3. **Stuck 30 minutes?** Ask for a hint. Don't lose an afternoon.
4. **Commit at the end of every day**, with a real message.
5. **Behind schedule?** Skip the stretch goals, never the "done when" checks.
   Day 7's afternoon is a buffer.
6. **Sleep 7+ hours.** Learning is consolidated during sleep; cutting it
   makes days 8–14 slower.

**Before day 1 (setup, about 1 hour):**
```bash
cd ~/go-basics && git init && git add . && git commit -m "Start bootcamp"
docker compose -f 05-clients/compose.yaml pull       # download images now
code ~/go-basics
```

---

## WEEK 1: Go, problem solving, data and APIs

### Day 1: Thinking like a programmer + Go basics I
**Goal:** learn the problem-solving method and Go's core syntax.
- **DSA warm-up:** read `04-dsa/README.md` (Big-O cheat sheet); `Arrays_TwoSum`
- **Block 1:** read `00-thinking/GUIDE.md` and `worked_example.go` carefully.
  Do [A Tour of Go](https://go.dev/tour) "Basics" (all 3 sections).
- **Block 2:** `01-basics` exercises 1–6 (variables through strings)
- **Block 3:** `00-thinking` exercise 1: `FormatReceipt`

**Done when:**
```bash
go test ./01-basics/ -run 'Test0[1-6]'
go test ./00-thinking/ -run Receipt
```

### Day 2: Go basics II
**Goal:** slices, maps, structs, pointers, errors, interfaces.
- **DSA:** `Arrays_IsPalindrome`, `Arrays_MaxSubarraySum`
- **Block 1:** Tour of Go "Methods and interfaces"; `01-basics` 7–9
- **Block 2:** `01-basics` 10–12
- **Block 3:** `00-thinking` exercise 2: `ShiftMinutes` (write 10 edge cases first!)

**Done when:** `go test ./01-basics/` all pass, and `go test ./00-thinking/ -run Shift`

### Day 3: Problem solving, debugging, interfaces in depth
- **DSA:** `Arrays_LongestUniqueSubstring`, `Arrays_MergeSorted`, `Arrays_GroupAnagrams`
- **Block 1:** `00-thinking` exercises 3 (`RunBank`) and 4 (Roman numerals)
- **Block 2:** `00-thinking` exercise 5 (debugging) and 6 (refactoring)
- **Block 3:** `03-advanced` Part 1 (interfaces 1–6)

**Done when:** `go test ./00-thinking/` and `go test ./03-advanced/ -run Interfaces`

### Day 4: Generics and concurrency I
**Goal:** goroutines and channels; the thing Go is famous for.
- **DSA:** `Stacks_ValidBrackets`, `Stacks_Queue`
- **Block 1:** `03-advanced` Part 2 (generics 7–11)
- **Block 2:** read Tour of Go "Concurrency"; `03-advanced` 12–14
- **Block 3:** `03-advanced` 15–16 (pipelines, fan-in). Draw the goroutines and channels on paper first.

**Done when:** `go test ./03-advanced/ -run 'Generics|Concurrency1[2-6]'`

### Day 5: Concurrency II + build your first app
- **DSA:** `Stacks_EvalRPN`, `Stacks_DailyTemperatures`
- **Block 1:** `03-advanced` 17–20 (select, worker pool, context, sync.Once)
- **Block 2:** `02-expense-cli` **Stage A** (interactive CRUD in memory)
- **Block 3:** `02-expense-cli` **Stage B** (save to JSON)

**Done when:** `go test ./03-advanced/` all pass; the CLI adds, lists, updates
and deletes, and survives a restart.

### Day 6: Real CLI, packages and testing
- **DSA:** `LinkedList_Reverse`, `LinkedList_Middle`, `LinkedList_HasCycle`
- **Block 1:** `02-expense-cli` **Stage C** (flag-based commands, summary)
- **Block 2:** `02-expense-cli` **Stage D** (packages, Store interface, table-driven tests)
- **Block 3:** read [Go by Example: testing](https://gobyexample.com/testing-and-benchmarking);
  add tests until `go test -cover ./02-expense-cli/...` shows 80% or more

**Done when:** `go build -o expense ./02-expense-cli && ./expense summary`
works, and `go test ./02-expense-cli/...` passes.

### Day 7: SQL (morning) + buffer (afternoon)
- **DSA:** `LinkedList_Merge`, `LinkedList_RemoveNthFromEnd`
- **Block 1:** `cd ~/fraudguard && make db-up && make seed && make psql`.
  Do `CURRICULUM.md` Level 1, exercises 1–10 (joins, GROUP BY, window functions, CTEs)
- **Block 2:** `02-expense-cli` **Stage E** (Postgres store), then exercises 13–14 (EXPLAIN, indexes)
- **Block 3: buffer.** Catch up on anything unfinished from days 1–6. If
  you're caught up: rest. Week 2 is harder.

**Done when:** `sql/exercises.sql` has 10+ answered queries, and the CLI works with `-store=postgres`.

---

## WEEK 2: Real backends, services and DevOps

### Day 8: HTTP APIs from scratch
**Goal:** understand HTTP all the way down, then build a REST API without a framework.
- **DSA:** `SearchSort_BinarySearch`, `SearchSort_FirstTrue`
- **Block 1:** FraudGuard `CURRICULUM.md` **0.5.1**: an HTTP server on raw TCP
- **Block 2:** `02-expense-cli` **Stage F**: a REST API with `net/http`
- **Block 3:** **0.5.2**: your own router with path parameters, 404 vs 405

**Done when:** all Stage F endpoints work with `curl` and return correct
status codes; your raw TCP server answers `curl -v localhost:9000/hello`.

### Day 9: Connecting to services I: Redis and Postgres
- **DSA:** `SearchSort_MergeSort`, `SearchSort_QuickSort`
- **Block 1:** read `05-clients/README.md`; try Redis by hand in `redis-cli`; exercises 1–3
- **Block 2:** Redis 4–6 (locks, leaderboard, queue)
- **Block 3:** Postgres 7–11. The deadlock in exercise 10 is the most important lesson of the day.

**Done when:** `go test ./05-clients/ -run 'Redis|Postgres'`

### Day 10: Connecting to services II + reading a production codebase
- **DSA:** `Recursion_Power`, `Recursion_Permutations`, `Recursion_Subsets`
- **Block 1:** `05-clients` HTTP client (12), then NATS (13–16)
- **Block 2:** `05-clients` S3 (17–20)
- **Block 3:** FraudGuard **Level 0.1**: trace a request through the code,
  answer the questions in `JOURNAL.md`

**Done when:** `go test ./05-clients/` all pass; `JOURNAL.md` 0.1 answered.

### Day 11: Working in a production codebase
**Goal:** add features to real code the way you would at a job.
- **DSA:** `Trees_InsertAndInOrder`, `Trees_Contains`, `Trees_LevelOrder`
- **Block 1:** FraudGuard **0.2** (`MultiCountry` rule) and **2.1** (idempotency body hash), on branches, with tests
- **Block 2:** FraudGuard **2.6**: API keys plus a **Redis rate limiter**
  (reuse your `Allow` from `05-clients`). Add Redis to its `docker-compose.yml`.
- **Block 3:** FraudGuard **3.2** (fuzz test) and **3.3** (benchmark)

**Done when:** `make test` and `make test-integration` pass in `~/fraudguard`,
and a client over the rate limit gets `429` with `Retry-After`.

### Day 12: DevOps I: Linux, Docker, Compose
- **DSA:** `Heap_MinHeap`, `Heap_KthLargest`
- **Block 1:** `~/devops-path` **Level 0**: `top_ips.sh`, `status_counts.sh`,
  `errors_by_hour.sh`, `backup.sh`
- **Block 2:** **Level 2**: Dockerfile for the visits app
- **Block 3:** **Level 3**: Compose with nginx, Postgres and 3 replicas

**Done when:** `./check.sh 2` and `./check.sh 3` pass (and Level 0 mostly passes).

### Day 13: DevOps II: CI/CD and Kubernetes
- **DSA:** `Graphs_ShortestPath`, `Graphs_NumIslands`
- **Block 1:** **Level 4**: push `devops-path` to GitHub, write the CI
  workflow, and get a green run that pushes the image
- **Block 2:** **Level 5**: install kind and kubectl, deploy the app and Postgres
- **Block 3:** get the Level 5 zero-downtime rolling restart check passing

**Done when:** `./check.sh 4` passes; `./check.sh 5` passes (or you can
explain exactly which check fails and why).

### Day 14: Capstone and interview day
- **DSA:** `DP_Fib`, `DP_ClimbStairs`, `DP_CoinChange`, then `Design_LRUCache`
- **Block 1:** **Ship FraudGuard:** push it to GitHub with a CI workflow
  (test, build, push image), and a README with what you added this week
- **Block 2:** **Mock interview with me** (paste "start the mock interview"):
  - 30 min: a coding problem, solved out loud with the 5-step method
  - 30 min: Go questions (goroutines, interfaces, errors, context)
  - 45 min: system design: "Design a rate-limited payment fraud API"
  - 15 min: feedback
- **Block 3:** write `RETRO.md`: what you can now do, what's still weak,
  and your plan for the next 4 weeks

**Done when:** FraudGuard is on GitHub with green CI, and you've done the mock interview.

---

## After the bootcamp

What the bootcamp skipped, in the order I'd do it (2–3 hours a day):

| Weeks | Work |
|---|---|
| 1–2 | Remaining `04-dsa` exercises, then NeetCode 150 in Go (2/day, ongoing) |
| 2–3 | FraudGuard Level 0.5.3–0.5.5, Level 2 (rest), Level 4 (observability) |
| 3–4 | `devops-path` Levels 6–8 (Helm, monitoring, Terraform) |
| 5–8 | FraudGuard Levels 5–6: Redis velocity, outbox + Kafka/NATS, gRPC, 3 replicas |
| 9+ | FraudGuard Level 7 capstone + `devops-path` Level 10: deploy it for real, write it up, apply for jobs |

## Progress tracker

Tick these off as you go:

- [ ] Day 1: method + basics 1–6 + receipt
- [ ] Day 2: basics 7–12 + shift
- [ ] Day 3: thinking 3–6 + interfaces
- [ ] Day 4: generics + concurrency 12–16
- [ ] Day 5: concurrency 17–20 + CLI A–B
- [ ] Day 6: CLI C–D + tests
- [ ] Day 7: SQL + CLI E
- [ ] Day 8: raw TCP + REST API + router
- [ ] Day 9: Redis + Postgres clients
- [ ] Day 10: HTTP + NATS + S3 clients + FraudGuard 0.1
- [ ] Day 11: FraudGuard features + Redis rate limiter
- [ ] Day 12: shell + Docker + Compose
- [ ] Day 13: CI/CD + Kubernetes
- [ ] Day 14: ship + mock interview + retro
