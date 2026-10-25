#!/usr/bin/env bash
# Level 2: Docker. Checks app/Dockerfile.
source "$(dirname "$0")/lib.sh"

need docker "install Docker: https://docs.docker.com/engine/install/"

IMAGE="visits:check"
VERSION="1.2.3"

section "Files"
require_file app/Dockerfile "write a multi-stage Dockerfile in app/"
if [ -f "$ROOT/app/.dockerignore" ]; then
	pass "app/.dockerignore exists"
else
	fail "app/.dockerignore is missing"
	hint "keep test files, .git and local junk out of the build context"
fi

section "Build"
if docker build -q -t "$IMAGE" --build-arg VERSION="$VERSION" "$ROOT/app" >/tmp/visits-build.log 2>&1; then
	pass "image builds"
else
	fail "docker build failed"
	tail -15 /tmp/visits-build.log | sed 's/^/        /'
	summary
fi

size=$(docker image inspect "$IMAGE" --format '{{.Size}}')
size_mb=$((size / 1024 / 1024))
if ((size_mb <= 30)); then
	pass "image is small (${size_mb} MB)"
else
	fail "image is ${size_mb} MB, want 30 MB or less"
	hint "multi-stage build: compile in golang:*, copy only the binary into distroless/static or scratch"
fi

user=$(docker image inspect "$IMAGE" --format '{{.Config.User}}')
if [ -n "$user" ] && [ "$user" != "root" ] && [ "$user" != "0" ] && [ "$user" != "0:0" ]; then
	pass "runs as a non-root user ($user)"
else
	fail "image runs as root"
	hint "add a USER instruction (distroless images have a 'nonroot' user)"
fi

if [ -n "$(docker image inspect "$IMAGE" --format '{{if .Config.Healthcheck}}{{.Config.Healthcheck.Test}}{{end}}')" ]; then
	pass "has a HEALTHCHECK"
else
	fail "no HEALTHCHECK instruction"
	hint "the app has a built-in -healthcheck flag, so no curl is needed"
fi

label=$(docker image inspect "$IMAGE" --format '{{index .Config.Labels "org.opencontainers.image.version"}}')
if [ "$label" == "$VERSION" ]; then
	pass "org.opencontainers.image.version label matches the VERSION build arg"
else
	fail "label org.opencontainers.image.version is '$label', want '$VERSION'"
	hint "ARG VERSION, then LABEL org.opencontainers.image.version=\$VERSION"
fi

section "Run"
port=$(free_port)
cid=$(docker run -d -p "127.0.0.1:$port:8080" "$IMAGE" 2>&1) || {
	fail "container failed to start: $cid"
	summary
}
cleanup() { docker rm -f "$cid" >/dev/null 2>&1; }
trap cleanup EXIT

if wait_http "http://127.0.0.1:$port/healthz" 20; then
	pass "/healthz responds on port 8080"
else
	fail "/healthz did not respond within 20s"
	docker logs "$cid" 2>&1 | tail -10 | sed 's/^/        /'
	summary
fi

if curl -fsS "http://127.0.0.1:$port/" | grep -q "\"version\":\"$VERSION\""; then
	pass "app reports version $VERSION"
else
	fail "GET / doesn't report version $VERSION"
	hint "pass it to the compiler: go build -ldflags \"-X main.version=\$VERSION\""
fi

status=""
for _ in $(seq 1 30); do
	status=$(docker inspect "$cid" --format '{{.State.Health.Status}}' 2>/dev/null)
	[ "$status" == "healthy" ] && break
	sleep 1
done
if [ "$status" == "healthy" ]; then
	pass "Docker reports the container healthy"
else
	fail "health status is '$status' after 30s"
	hint "use a short interval, e.g. HEALTHCHECK --interval=5s"
fi

start=$(date +%s)
docker stop "$cid" >/dev/null
elapsed=$(($(date +%s) - start))
code=$(docker inspect "$cid" --format '{{.State.ExitCode}}')
if ((elapsed < 8)) && [ "$code" == "0" ]; then
	pass "stops gracefully on SIGTERM (${elapsed}s, exit code 0)"
else
	fail "docker stop took ${elapsed}s with exit code $code"
	hint "use the exec form ENTRYPOINT [\"/app\"], not the shell form, so SIGTERM reaches the app"
fi

summary
