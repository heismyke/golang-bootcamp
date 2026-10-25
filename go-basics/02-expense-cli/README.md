# Assignment 2: Expense tracker CLI (CRUD)

You build a command-line app that tracks your spending: **C**reate, **R**ead,
**U**pdate and **D**elete expenses. You write all the code in `main.go` (and
later, other files you create). Each stage builds on the last one.

An expense looks like this:

```go
type Expense struct {
	ID         int
	AmountYen  int       // money is always an integer, never a float
	Category   string    // food, transport, rent, ...
	Note       string
	Date       time.Time
}
```

---

## Stage A: Interactive and in memory (structs, slices, loops, switch)

The program shows a prompt, reads commands until you type `quit`, and keeps
expenses in a slice (`[]Expense`). They are lost when the program exits.
That's fine for now.

```
> add 1200 food lunch with team
added #1
> add 450 transport train
added #2
> list
#1  ¥1,200  food       lunch with team   2026-10-25
#2  ¥450    transport  train             2026-10-25
> delete 2
deleted #2
> get 2
error: expense #2 not found
> quit
```

Commands: `add <amount> <category> <note...>`, `list`, `get <id>`,
`update <id> <amount> <category> <note...>`, `delete <id>`, `help`, `quit`.

**Hints:** `bufio.NewScanner(os.Stdin)` reads lines, `strings.Fields` splits
words, `strconv.Atoi` converts text to numbers, and `switch` picks the command.

✅ Done when: every command works, bad input (`add abc food`, `delete 99`,
`add` with nothing after it) prints a clear error instead of crashing, and
IDs keep increasing even after a delete.

---

## Stage B: Save to a JSON file (encoding/json, files, error handling)

Expenses now survive restarts. Load `expenses.json` at startup and save it
after every change.

**Hints:** add JSON tags (`` `json:"amount_yen"` ``), and use `json.MarshalIndent`,
`os.WriteFile` and `os.ReadFile`. If the file doesn't exist yet, start
empty instead of failing (`errors.Is(err, os.ErrNotExist)`).

✅ Done when: you can add, quit, restart, and still `list` your expenses.
Then delete the file and confirm it still starts. Then put garbage in the
file and confirm it shows a clear error.

---

## Stage C: Real CLI commands (os.Args, flag package)

Replace the interactive prompt with commands, like real tools (`git`, `docker`):

```bash
go run ./02-expense-cli add -amount 1200 -category food -note "lunch"
go run ./02-expense-cli list -category food
go run ./02-expense-cli update 3 -amount 1500
go run ./02-expense-cli delete 3
go run ./02-expense-cli summary
```

`summary` shows the total per category and the overall total, sorted by
amount (biggest first). `update` changes only the flags you pass.

**Hints:** `os.Args[1]` is the command, and `flag.NewFlagSet("add", flag.ExitOnError)`
gives each command its own flags. Exit with `os.Exit(1)` on errors.

✅ Done when: everything above works, `go build -o expense ./02-expense-cli`
creates a binary, and `./expense list` works.

---

## Stage D: Packages, interfaces and tests (code organization)

Split the code into files and packages:

```
02-expense-cli/
  main.go                 only reads commands and prints results
  expense/expense.go      the Expense type and validation
  expense/store.go        a Store interface: Add, Get, List, Update, Delete
  expense/jsonstore.go    the JSON file implementation of Store
  expense/memstore.go     an in-memory implementation (for tests)
  expense/expense_test.go
```

✅ Done when: `main.go` only talks to the `Store` interface, and
`go test ./02-expense-cli/...` has table-driven tests for validation and
for every Store method (run against `memstore`).

---

## Stage E: Use a real database (SQL, PostgreSQL)

Write a `postgresstore.go` that implements the same `Store` interface using
PostgreSQL (`github.com/jackc/pgx/v5`). Thanks to the interface, `main.go`
barely changes. That's the point of Stage D.

Use the Postgres from FraudGuard (`cd ~/fraudguard && make db-up`), with
connection string `postgres://fraudguard:fraudguard@localhost:55432/fraudguard`.
Create your own table:

```sql
CREATE TABLE expenses (
    id         bigserial PRIMARY KEY,
    amount_yen bigint NOT NULL CHECK (amount_yen > 0),
    category   text   NOT NULL,
    note       text   NOT NULL DEFAULT '',
    spent_on   date   NOT NULL DEFAULT current_date
);
```

✅ Done when: a `-store=json|postgres` flag switches storage, and `summary`
is a single SQL query (`GROUP BY category ORDER BY sum DESC`) instead of a Go loop.

---

## Stage F: Turn it into an HTTP API (net/http, no framework)

Add `go run ./02-expense-cli serve` to start an HTTP server using the same Store:

```
POST   /expenses          create (JSON body)
GET    /expenses          list (?category=food)
GET    /expenses/{id}     get one
PUT    /expenses/{id}     update
DELETE /expenses/{id}     delete
GET    /summary           totals per category
```

✅ Done when: every endpoint works with `curl`, returns the right status
codes (201, 200, 204, 400, 404), and errors come back as JSON.

**After Stage F you're ready for FraudGuard**, which is this same pattern,
production-grade.
