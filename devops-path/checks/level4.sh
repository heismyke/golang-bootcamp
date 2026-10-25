#!/usr/bin/env bash
# Level 4: CI/CD with GitHub Actions. Checks .github/workflows/ci.yml and,
# if the repo is on GitHub, that the latest run on main succeeded.
source "$(dirname "$0")/lib.sh"

WF="$ROOT/.github/workflows/ci.yml"

section "Workflow file"
require_file .github/workflows/ci.yml "create the GitHub Actions workflow"

has() { grep -qE "$1" "$WF"; }
has '^on:|^"on":' && has 'push' && has 'pull_request' &&
	pass "runs on push and pull_request" || fail "workflow should trigger on push and pull_request"
has 'actions/setup-go' && pass "sets up Go" || fail "doesn't use actions/setup-go"
has 'go vet' && pass "runs go vet" || fail "doesn't run go vet"
has 'gofmt|golangci-lint' && pass "checks formatting or lints" || fail "doesn't check formatting (gofmt -l) or run golangci-lint"
has 'go test' && pass "runs go test" || fail "doesn't run go test"
has 'docker/build-push-action|docker build|docker buildx build' && pass "builds the Docker image" || fail "doesn't build the Docker image"
has 'ghcr.io' && pass "pushes to GitHub Container Registry" || fail "doesn't push to ghcr.io"
has 'needs:' && pass "jobs depend on each other (needs:)" || fail "no 'needs:'; the image should only build after tests pass"
has "github.ref == 'refs/heads/main'|refs/tags|branches:" && pass "only publishes from main or tags" ||
	fail "publishing isn't limited to main or tags"
has 'permissions:' && pass "declares token permissions" || fail "no 'permissions:' block (least privilege for GITHUB_TOKEN)"
if has 'cache'; then pass "uses caching"; else fail "no caching (setup-go caches modules; build-push-action supports cache-from/cache-to)"; fi

section "GitHub"
if ! command -v gh >/dev/null || ! gh auth status >/dev/null 2>&1; then
	fail "gh is not logged in, can't check workflow runs"
	hint "run: gh auth login"
	summary
fi
if ! (cd "$ROOT" && git remote get-url origin >/dev/null 2>&1); then
	fail "this repo has no GitHub remote"
	hint "gh repo create devops-path --private --source . --push"
	summary
fi
run=$(cd "$ROOT" && gh run list --workflow ci.yml --branch main --limit 1 --json status,conclusion,url 2>/dev/null)
conclusion=$(jq -r '.[0].conclusion // empty' <<<"$run")
url=$(jq -r '.[0].url // empty' <<<"$run")
if [ "$conclusion" == "success" ]; then
	pass "latest CI run on main succeeded ($url)"
elif [ -z "$run" ] || [ "$run" == "[]" ]; then
	fail "no CI runs on main yet"
	hint "push to main and watch: gh run watch"
else
	fail "latest CI run on main: ${conclusion:-in progress} ($url)"
	hint "see why: gh run view --log-failed"
fi

summary
