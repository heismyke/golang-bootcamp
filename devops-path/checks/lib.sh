# Shared helpers for the check scripts. Sourced, not run directly.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PASS=0
FAIL=0

section() { printf '\n\033[1m%s\033[0m\n' "$*"; }
pass()    { printf '  \033[32m✔\033[0m %s\n' "$*"; PASS=$((PASS + 1)); }
fail()    { printf '  \033[31m✘\033[0m %s\n' "$*"; FAIL=$((FAIL + 1)); }
hint()    { printf '      \033[33mhint:\033[0m %s\n' "$*"; }
info()    { printf '      %s\n' "$*"; }

summary() {
	echo
	if [ "$FAIL" -eq 0 ]; then
		printf '\033[32mAll %d checks passed.\033[0m\n' "$PASS"
		exit 0
	fi
	printf '\033[31m%d passed, %d failed.\033[0m\n' "$PASS" "$FAIL"
	exit 1
}

# need CMD HINT: stop early if a required tool is missing.
need() {
	if ! command -v "$1" >/dev/null 2>&1; then
		fail "$1 is not installed"
		hint "$2"
		summary
	fi
}

# require_file PATH HINT: stop early if a deliverable doesn't exist yet.
require_file() {
	if [ ! -e "$ROOT/$1" ]; then
		fail "$1 does not exist yet"
		hint "$2"
		summary
	fi
	pass "$1 exists"
}

# wait_http URL [SECONDS]: wait until URL returns 2xx.
wait_http() {
	local url=$1 timeout=${2:-30} i
	for ((i = 0; i < timeout; i++)); do
		curl -fsS -o /dev/null --max-time 2 "$url" 2>/dev/null && return 0
		sleep 1
	done
	return 1
}

# free_port: print a free TCP port.
free_port() {
	python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1])'
}
