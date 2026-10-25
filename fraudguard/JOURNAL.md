# Journal

## Level 0.1 — Codebase walkthrough

### Request flow
<!-- In your own words, list the steps a POST /v1/transactions request goes through,
     from main.go to the database and back. One line per step. -->
1.
2.
3.

### Q1: Why doesn't the `fraud` package import the `transaction` package? What would go wrong if it did?
<!-- Look at the imports at the top of internal/fraud/engine.go and internal/transaction/service.go -->


### Q2: What does `defer` do in migrate.go? Why does the unlock use `context.Background()` instead of `ctx`?
<!-- Look at internal/platform/postgres/migrate.go. Hint: what if ctx is already cancelled? -->


### Q3: `Repository` is an interface. Name two places in the code that benefit from that.
<!-- Look at internal/transaction/repository.go, then search for who implements it -->


### Q4: What happens if two requests with the same Idempotency-Key arrive at the same moment? Which lines keep it safe?
<!-- Look at Ingest() in internal/transaction/service.go and the transactions table in migrations/0001_init.sql -->


### Bonus: What attack would succeed without the account lock?
<!-- Read the comment above LockAccount in service.go -->


### Experiments (curl results)
<!-- Paste what you sent and what came back: normal request, replay, missing header, bad amount, bad JSON -->


### What I learned / what confused me

