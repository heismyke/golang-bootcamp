DATABASE_URL ?= postgres://fraudguard:fraudguard@localhost:55432/fraudguard?sslmode=disable
HTTP_ADDR    ?= :8088
export DATABASE_URL HTTP_ADDR

PSQL = docker compose exec -T postgres psql -U fraudguard -d fraudguard

.PHONY: run test test-race test-integration vet fmt db-up db-down db-reset seed psql up

run:              ## Run the API locally against the compose database
	go run ./cmd/api

test:             ## Unit tests (no database needed)
	go test ./...

test-race:        ## Unit tests with the race detector (needs gcc: sudo apt install gcc)
	CGO_ENABLED=1 go test -race ./...

test-integration: ## Unit + Postgres integration tests (run `make db-up` first)
	go test -count=1 -tags=integration ./...

vet:
	go vet ./...
	go vet -tags=integration ./...

fmt:
	gofmt -l -w .

db-up:            ## Start Postgres in the background
	docker compose up -d --wait postgres

db-down:
	docker compose down

db-reset:         ## Destroy the database volume and start fresh
	docker compose down -v

seed:             ## Load ~200k realistic transactions for SQL practice
	$(PSQL) < scripts/seed.sql

psql:             ## Open an interactive SQL shell
	docker compose exec postgres psql -U fraudguard -d fraudguard

up:               ## Build and run the full stack (API on :8088)
	docker compose up --build
