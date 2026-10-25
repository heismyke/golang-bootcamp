#!/usr/bin/env bash
# Level 0: Linux and shell scripting.
source "$(dirname "$0")/lib.sh"

LOG="$ROOT/data/access.log"
DIR="$ROOT/level0"

# expected TASK [ARGS...]: compute the correct answer with Python, so you can
# compare your shell pipeline against an independent implementation.
expected() {
	python3 - "$LOG" "$@" <<'EOF'
import re, sys, collections
log, task, args = sys.argv[1], sys.argv[2], sys.argv[3:]
pat = re.compile(r'^(\S+) \S+ \S+ \[\d+/\w+/\d+:(\d\d):\d\d:\d\d [^\]]+\] "(\S+) (\S+) [^"]*" (\d{3}) \d+ "[^"]*" "[^"]*" ([\d.]+)$')
rows = [pat.match(l.rstrip("\n")).groups() for l in open(log)]
if task == "top_ips":
    n = int(args[0]) if args else 5
    c = collections.Counter(r[0] for r in rows)
    for ip, k in sorted(c.items(), key=lambda x: (-x[1], x[0]))[:n]:
        print(k, ip)
elif task == "status_counts":
    c = collections.Counter(r[4] for r in rows)
    for code in sorted(c):
        print(code, c[code])
elif task == "errors_by_hour":
    c = collections.Counter(r[1] for r in rows if r[4].startswith("5"))
    for h in sorted(c):
        print(h, c[h])
elif task == "slow_requests":
    th = float(args[0]) if args else 1.0
    slow = [(float(r[5]), r[5], r[3]) for r in rows if float(r[5]) > th]
    for _, rt, path in sorted(slow, key=lambda x: (-x[0], x[2]))[:10]:
        print(rt, path)
elif task == "suspicious":
    c = collections.Counter(r[0] for r in rows if r[3] in ("/admin", "/wp-login.php"))
    for ip, k in sorted(c.items(), key=lambda x: (-x[1], x[0])):
        if k >= 3:
            print(k, ip)
EOF
}

# check_script NAME: the file exists, is executable, has a shebang, parses.
check_script() {
	local f="$DIR/$1"
	if [ ! -f "$f" ]; then
		fail "level0/$1 does not exist yet"
		return 1
	fi
	if [ ! -x "$f" ]; then
		fail "level0/$1 is not executable"
		hint "chmod +x level0/$1"
		return 1
	fi
	if ! head -1 "$f" | grep -q '^#!'; then
		fail "level0/$1 has no shebang line"
		hint "the first line should be #!/usr/bin/env bash"
		return 1
	fi
	if ! bash -n "$f" 2>/dev/null; then
		fail "level0/$1 has a syntax error"
		hint "run: bash -n level0/$1"
		return 1
	fi
	if command -v shellcheck >/dev/null && ! shellcheck -S warning "$f" >/dev/null 2>&1; then
		fail "level0/$1 has shellcheck warnings"
		hint "run: shellcheck level0/$1"
		return 1
	fi
	return 0
}

# compare NAME TASK [ARGS...]: run the script and compare its output.
compare() {
	local name=$1 task=$2
	shift 2
	check_script "$name" || return
	local got want
	got=$(cd /tmp && "$DIR/$name" "$LOG" "$@" 2>&1)
	want=$(expected "$task" "$@")
	if [ "$got" == "$want" ]; then
		pass "$name${*:+ $*} gives the right output"
	else
		fail "$name${*:+ $*} output is wrong"
		info "expected (first lines):"
		head -5 <<<"$want" | sed 's/^/        /'
		info "got (first lines):"
		head -5 <<<"$got" | sed 's/^/        /'
	fi
}

section "Log analysis (data/access.log)"
compare top_ips.sh top_ips
compare top_ips.sh top_ips 3
compare status_counts.sh status_counts
compare errors_by_hour.sh errors_by_hour
compare slow_requests.sh slow_requests
compare slow_requests.sh slow_requests 2.5
compare suspicious.sh suspicious

section "System information"
if check_script sysinfo.sh; then
	out=$("$DIR/sysinfo.sh" 2>&1)
	field() { grep -E "^$1: " <<<"$out" | head -1 | cut -d' ' -f2-; }
	[ "$(field hostname)" == "$(hostname)" ] && pass "hostname is correct" || fail "hostname: got '$(field hostname)', want '$(hostname)'"
	[ "$(field kernel)" == "$(uname -r)" ] && pass "kernel is correct" || fail "kernel: got '$(field kernel)', want '$(uname -r)'"
	[ "$(field cpus)" == "$(nproc)" ] && pass "cpus is correct" || fail "cpus: got '$(field cpus)', want '$(nproc)'"
	want_mem=$(awk '/MemTotal/ {print int($2/1024)}' /proc/meminfo)
	got_mem=$(field memory_total_mb)
	if [[ "$got_mem" =~ ^[0-9]+$ ]] && ((got_mem >= want_mem - 50 && got_mem <= want_mem + 50)); then
		pass "memory_total_mb is correct"
	else
		fail "memory_total_mb: got '$got_mem', want about $want_mem"
		hint "read MemTotal from /proc/meminfo (it's in kB)"
	fi
	want_disk=$(df -BG --output=avail / | tail -1 | tr -dc '0-9')
	got_disk=$(field disk_free_gb)
	if [[ "$got_disk" =~ ^[0-9]+$ ]] && ((got_disk >= want_disk - 2 && got_disk <= want_disk + 2)); then
		pass "disk_free_gb is correct"
	else
		fail "disk_free_gb: got '$got_disk', want about $want_disk"
	fi
fi

section "Backup script"
if check_script backup.sh; then
	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT
	mkdir -p "$tmp/src/sub" "$tmp/dest"
	echo hello >"$tmp/src/a.txt"
	echo world >"$tmp/src/sub/b.txt"

	if "$DIR/backup.sh" >/dev/null 2>&1; then
		fail "backup.sh with no arguments should exit with an error"
		hint "print a usage message to stderr and exit 1"
	else
		pass "backup.sh with no arguments exits with an error"
	fi
	if "$DIR/backup.sh" "$tmp/missing" "$tmp/dest" >/dev/null 2>&1; then
		fail "backup.sh should fail when the source doesn't exist"
	else
		pass "backup.sh fails when the source doesn't exist"
	fi

	for _ in 1 2 3 4 5; do
		"$DIR/backup.sh" "$tmp/src" "$tmp/dest" >/dev/null 2>&1
		sleep 1.1 # names include seconds, so they differ
	done
	mapfile -t backups < <(find "$tmp/dest" -maxdepth 1 -name 'backup-*.tar.gz' | sort)
	if [ "${#backups[@]}" -eq 3 ]; then
		pass "only the 3 newest backups are kept"
	else
		fail "after 5 runs there are ${#backups[@]} backups, want 3"
		hint "name them backup-YYYYmmdd-HHMMSS.tar.gz, then delete all but the newest 3"
	fi
	if [ "${#backups[@]}" -gt 0 ] && tar -tzf "${backups[-1]}" 2>/dev/null | grep -q 'sub/b.txt'; then
		pass "the backup contains the source files"
	else
		fail "the newest backup doesn't contain sub/b.txt"
	fi
fi

summary
