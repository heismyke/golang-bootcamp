package thinking

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ═════════════════════════════════════════════════════════════
// EXERCISE 6: REFACTORING
// Summarize WORKS: its tests already pass. But it's hard to read, hard to
// change, and hard to test. Your job: make it clean WITHOUT changing what
// it does. Run the tests after every small change; they must stay green.
//
//	go test ./00-thinking/ -v -run Messy
//
// Things to fix:
//   - one giant function → small functions with clear names
//     (e.g. parseLine, totalsByCategory, formatReport)
//   - bad names (d, t, x, s2, flag)
//   - deep nesting → early returns / continue
//   - magic numbers and strings → named constants
//   - duplicated code
//   - a struct for a parsed line instead of juggling three variables
//
// What it does: each line is "name,category,amountYen". Lines that are
// blank, start with '#', don't have exactly 3 fields, have an empty
// category, or have an amount that isn't a positive integer are skipped.
// Category names are case-insensitive and trimmed. The report lists each
// category with its total and count, biggest total first (ties
// alphabetical), then the grand total and the number of invalid lines
// (blank lines and '#' comments are ignored silently, not counted).
// ═════════════════════════════════════════════════════════════

func Summarize(lines []string) string {
	d := map[string]int{}
	c := map[string]int{}
	s2 := 0
	t := 0
	for _, x := range lines {
		x = strings.TrimSpace(x)
		if x != "" {
			if !strings.HasPrefix(x, "#") {
				p := strings.Split(x, ",")
				if len(p) == 3 {
					cat := strings.ToLower(strings.TrimSpace(p[1]))
					if cat != "" {
						a, err := strconv.Atoi(strings.TrimSpace(p[2]))
						if err == nil {
							if a > 0 {
								d[cat] = d[cat] + a
								c[cat] = c[cat] + 1
								t = t + a
							} else {
								s2++
							}
						} else {
							s2++
						}
					} else {
						s2++
					}
				} else {
					s2++
				}
			}
		}
	}
	var keys []string
	for k := range d {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		flag := false
		if d[keys[i]] > d[keys[j]] {
			flag = true
		} else if d[keys[i]] == d[keys[j]] {
			if keys[i] < keys[j] {
				flag = true
			}
		}
		return flag
	})
	out := ""
	for _, k := range keys {
		out = out + fmt.Sprintf("%-12s %8d (%d)\n", k, d[k], c[k])
	}
	out = out + fmt.Sprintf("%-12s %8d\n", "TOTAL", t)
	out = out + fmt.Sprintf("skipped: %d", s2)
	return out
}
