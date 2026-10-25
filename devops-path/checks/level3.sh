#!/usr/bin/env bash
# Level 3: Docker Compose. Checks compose.yaml (app + Postgres + nginx).
# Runs under its own project name, so it never touches your own stack.
source "$(dirname "$0")/lib.sh"

need docker "install Docker: https://docs.docker.com/engine/install/"

PROJECT="visits-check"
URL="http://127.0.0.1:8090"
compose() { docker compose -p "$PROJECT" --project-directory "$ROOT" -f "$ROOT/compose.yaml" "$@"; }
# published SERVICE PORT: is the container port mapped to a real host port?
published() { compose port "$1" "$2" 2>/dev/null | grep -qE ':[1-9][0-9]*$'; }

section "Files"
require_file compose.yaml "define the app, postgres and nginx services"
require_file nginx/nginx.conf "nginx reverse proxy config that forwards to the app"
if [ -f "$ROOT/.env.example" ]; then
	pass ".env.example exists"
else
	fail ".env.example is missing"
	hint "commit an example env file with placeholder values; keep the real .env out of git"
fi
if grep -qxE '\.env|/\.env' "$ROOT/.gitignore" 2>/dev/null; then
	pass ".env is gitignored"
else
	fail ".env is not in .gitignore"
fi
if grep -qE 'POSTGRES_PASSWORD:[[:space:]]*[^$[:space:]]' "$ROOT/compose.yaml"; then
	fail "compose.yaml has a hardcoded POSTGRES_PASSWORD"
	hint "use \${POSTGRES_PASSWORD} or env_file: .env"
else
	pass "no hardcoded database password"
fi
if [ ! -f "$ROOT/.env" ] && [ -f "$ROOT/.env.example" ]; then
	cp "$ROOT/.env.example" "$ROOT/.env"
	info "(created .env from .env.example)"
fi
if compose config -q 2>/tmp/visits-compose.log; then
	pass "compose.yaml is valid"
else
	fail "compose.yaml is invalid"
	sed 's/^/        /' /tmp/visits-compose.log | head -10
	summary
fi

cleanup() { compose down -v --remove-orphans >/dev/null 2>&1; }
trap cleanup EXIT
cleanup

section "Start"
if compose up -d --build --wait --wait-timeout 120 >/tmp/visits-compose.log 2>&1; then
	pass "stack starts and every service with a healthcheck becomes healthy"
else
	fail "docker compose up --wait failed"
	tail -15 /tmp/visits-compose.log | sed 's/^/        /'
	compose logs --tail 10 2>&1 | sed 's/^/        /'
	summary
fi

for svc in app postgres nginx; do
	if [ -n "$(compose ps -q "$svc" 2>/dev/null)" ]; then
		pass "service '$svc' is running"
	else
		fail "service '$svc' is missing"
		hint "the checks expect services named app, postgres and nginx"
	fi
done
for svc in app postgres; do
	hc=$(docker inspect "$(compose ps -q "$svc" | head -1)" --format '{{if .Config.Healthcheck}}yes{{end}}' 2>/dev/null)
	[ "$hc" == "yes" ] && pass "'$svc' has a healthcheck" || fail "'$svc' has no healthcheck"
done

section "Networking"
if wait_http "$URL/" 30; then
	pass "nginx serves the app on localhost:8090"
else
	fail "nothing answers on $URL"
	hint "publish nginx as \"8090:80\" and proxy_pass to http://app:8080"
	summary
fi
if published app 8080; then
	fail "the app's port is published to the host"
	hint "only nginx should be reachable from outside; remove 'ports' from app"
else
	pass "the app is only reachable through nginx"
fi
if published postgres 5432; then
	fail "Postgres is published to the host"
	hint "the database should only be on the internal network"
else
	pass "Postgres is not exposed to the host"
fi
if curl -fsS "$URL/readyz" | grep -q '"storage":"postgres"'; then
	pass "the app is connected to Postgres"
else
	fail "/readyz doesn't report postgres storage"
	hint "set DATABASE_URL=postgres://USER:PASS@postgres:5432/DB?sslmode=disable for the app"
fi

section "Persistence"
before=$(curl -fsS -X POST "$URL/visits" | sed -E 's/.*"count":([0-9]+).*/\1/')
curl -fsS -X POST "$URL/visits" >/dev/null
compose down >/dev/null 2>&1 # keeps volumes
compose up -d --wait --wait-timeout 120 >/dev/null 2>&1
wait_http "$URL/readyz" 30
after=$(curl -fsS "$URL/visits" | sed -E 's/.*"count":([0-9]+).*/\1/')
if [[ "$before" =~ ^[0-9]+$ ]] && [ "$after" == "$((before + 1))" ]; then
	pass "data survives 'docker compose down' and 'up'"
else
	fail "visit count was '$before' then '$after' after a restart, want $((before + 1))"
	hint "store Postgres data in a named volume mounted at /var/lib/postgresql/data"
fi

section "Scaling"
compose up -d --scale app=3 --wait --wait-timeout 120 >/dev/null 2>&1
compose restart nginx >/dev/null 2>&1
wait_http "$URL/" 30
hosts=$(for _ in $(seq 1 30); do curl -fsS "$URL/" 2>/dev/null; echo; done | grep -o '"hostname":"[^"]*"' | sort -u | wc -l)
if ((hosts >= 2)); then
	pass "with 3 app replicas, nginx spreads requests across $hosts of them"
else
	fail "all requests went to $hosts replica(s)"
	hint "don't set container_name on app; nginx resolves 'app' to every replica's IP"
fi

summary
