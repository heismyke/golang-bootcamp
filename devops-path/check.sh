#!/usr/bin/env bash
# Usage: ./check.sh LEVEL    e.g. ./check.sh 0
set -uo pipefail
cd "$(dirname "$0")"

level=${1:-}
script="checks/level${level}.sh"
if [ -z "$level" ] || [ ! -f "$script" ]; then
	echo "usage: ./check.sh LEVEL"
	echo "levels with automatic checks: $(ls checks/level*.sh | sed -E 's/.*level([0-9]+)\.sh/\1/' | sort -n | tr '\n' ' ')"
	echo "(levels 1, 9 and 10 use the checklists in README.md)"
	exit 2
fi
exec bash "$script"
