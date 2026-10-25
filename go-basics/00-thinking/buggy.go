package thinking

import (
	"slices"
	"strings"
	"unicode"
)

// ═════════════════════════════════════════════════════════════
// EXERCISE 5: DEBUGGING
// Each function below has at least one bug. The doc comment says what it
// SHOULD do. Fix them using the debugging method from GUIDE.md:
//
//  1. run the test and read the failure carefully
//  2. form a hypothesis ("I think X happens because Y")
//  3. check it (add a print, or a smaller test)
//  4. fix the cause, not the symptom
//
// For each bug, write a comment above your fix: what was wrong and how
// you found it. Don't rewrite whole functions; find the actual mistakes.
//
//	go test ./00-thinking/ -v -run Buggy
// ═════════════════════════════════════════════════════════════

// Average returns the mean of nums, or 0 for an empty slice.
// Average([]int{1, 2}) == 1.5
func Average(nums []int) float64 {
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return float64(sum / len(nums))
}

// Median returns the middle value of nums (the mean of the two middle
// values for an even count), or 0 for an empty slice. It must NOT change
// the caller's slice.
// Median([]int{5, 1, 3}) == 3, Median([]int{4, 1, 3, 2}) == 2.5
func Median(nums []int) float64 {
	if len(nums) == 0 {
		return 0
	}
	slices.Sort(nums)
	mid := len(nums) / 2
	if len(nums)%2 == 0 {
		return float64(nums[mid]+nums[mid+1]) / 2
	}
	return float64(nums[mid])
}

// Paginate returns page number `page` (1-based) of items, with `size`
// items per page. A page past the end returns an empty slice. The last
// page may be shorter.
// Paginate([a b c d e], 2, 2) == [c d], Paginate([a b c d e], 3, 2) == [e]
func Paginate(items []string, page, size int) []string {
	start := page * size
	end := start + size
	return items[start:end]
}

// IsLeapYear reports whether year is a leap year in the Gregorian
// calendar: divisible by 4, except years divisible by 100, except years
// divisible by 400. 2024 → true, 1900 → false, 2000 → true, 2023 → false.
func IsLeapYear(year int) bool {
	return year%4 == 0
}

// CamelToSnake converts camelCase or PascalCase to snake_case, keeping
// acronyms together:
//
//	"userName"   → "user_name"
//	"UserID"     → "user_id"
//	"HTTPServer" → "http_server"
//	"getHTTPResponseCode" → "get_http_response_code"
//	"id"         → "id"
//
// Rule: put an underscore before an uppercase letter if the previous letter
// is lowercase, OR if the next letter is lowercase and the previous one
// is uppercase (that's where an acronym ends).
func CamelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteRune('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
