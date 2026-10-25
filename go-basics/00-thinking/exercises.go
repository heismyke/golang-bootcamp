// Package thinking: problem-solving exercises. YOUR CODE GOES IN
// exercises.go, buggy.go and messy.go. Read GUIDE.md and
// worked_example.go first.
//
// For each exercise in this file, write STEP 1 (understand: examples and
// edge cases) and STEP 2 (plan: pseudocode) as comments BEFORE coding.
//
//	go test ./00-thinking/ -v -run Receipt
package thinking

// ═════════════════════════════════════════════════════════════
// EXERCISE 1: DECOMPOSITION
// Skill: break a problem into small functions, each easy to get right.
// ═════════════════════════════════════════════════════════════

type Item struct {
	Name     string
	PriceYen int // price of one unit
	Qty      int
}

// FormatReceipt returns a printable receipt. Every line is exactly:
//
//	fmt.Sprintf("%-14s%10s", label, amount)
//
// i.e. the label left-aligned in 14 characters and the amount
// right-aligned in 10. Amounts look like ¥1,234 (yen sign, commas).
//
// For items []Item{{"Coffee", 400, 2}, {"Sandwich", 650, 1}} the result is:
//
//	Coffee x2           ¥800
//	Sandwich x1         ¥650
//	------------------------          (24 dashes)
//	Subtotal          ¥1,450
//	Tax 10%             ¥145
//	Total             ¥1,595
//
// Rules:
//   - an item line's label is "<name> x<qty>"; its amount is price × qty
//   - names longer than 10 characters are cut to 10 ("Strawberry Cake" → "Strawberry")
//   - items with Qty 0 or less are left out
//   - tax is 10% of the subtotal, rounded DOWN to the yen
//   - lines are joined with "\n", and there's no newline at the end
//   - if there are no items to print, return exactly "no items"
//
// Before coding, decide which helper functions you'll write. Hint: you
// probably want at least: formatYen, line, itemLabel. Test each one.
func FormatReceipt(items []Item) string {
	return "" // TODO
}

// ═════════════════════════════════════════════════════════════
// EXERCISE 2: EDGE CASES
// Skill: list every way the input can be weird before you code.
// ═════════════════════════════════════════════════════════════

// ShiftMinutes takes a work shift like "09:00-17:30" and returns its
// length in minutes (510).
//
//   - Shifts can cross midnight: "22:00-02:00" is 240 minutes.
//   - Times are HH:MM with exactly two digits each, 00:00 to 23:59.
//   - Spaces around the whole string are allowed: " 09:00-10:00 " is fine.
//   - A zero-length shift ("09:00-09:00") is an error.
//   - Anything malformed is an error: "9:00-17:00", "25:00-01:00",
//     "09:60-10:00", "09:00", "09:00-", "", "aa:bb-cc:dd", "09:00 - 10:00".
//
// Before coding: write down at least 10 test inputs and their expected
// result in a comment. The tests here check more than the examples above.
func ShiftMinutes(shift string) (int, error) {
	return 0, nil // TODO
}

// ═════════════════════════════════════════════════════════════
// EXERCISE 3: MODELLING STATE
// Skill: decide what data you need to remember, then follow a spec exactly.
// ═════════════════════════════════════════════════════════════

// RunBank runs a list of commands against a tiny bank and returns one
// output line per command. All accounts start at 0. Amounts are whole yen.
//
//	Command                  Output on success
//	open NAME                ok
//	deposit NAME AMOUNT      ok
//	withdraw NAME AMOUNT     ok
//	transfer FROM TO AMOUNT  ok
//	balance NAME             NAME: BALANCE        e.g. "alice: 70"
//
// Errors (checked in this order, first match wins):
//
//	unknown command word               error: unknown command
//	wrong number of words              error: wrong number of arguments
//	open an existing account           error: account exists
//	any account that wasn't opened     error: no such account
//	transfer to the same account       error: same account
//	amount not a whole number > 0      error: invalid amount
//	withdraw/transfer more than held   error: insufficient funds
//
// A failed command changes nothing. Words are separated by single spaces;
// "" (an empty command) is an unknown command.
//
// Hint: what state must you keep between commands? What's the simplest
// type that holds it?
func RunBank(commands []string) []string {
	return nil // TODO
}

// ═════════════════════════════════════════════════════════════
// EXERCISE 4: FINDING THE PATTERN
// Skill: work examples by hand until you see the rule, then code the rule.
// ═════════════════════════════════════════════════════════════

// ToRoman converts 1–3999 to Roman numerals: 4 → "IV", 9 → "IX",
// 14 → "XIV", 1994 → "MCMXCIV", 3999 → "MMMCMXCIX".
// Returns an error outside 1–3999.
//
// Symbols: I=1 V=5 X=10 L=50 C=100 D=500 M=1000. Write out 1 to 20 by hand
// first. What are the only "subtractive" pairs that exist?
func ToRoman(n int) (string, error) {
	return "", nil // TODO
}

// FromRoman converts a Roman numeral back to a number, and returns an error
// for anything that isn't the one correct (canonical) way of writing a
// number: "IIII", "VV", "IC", "MCMC", "iv", "" and "ABC" are all errors.
//
// Writing all the rules for invalid numerals is surprisingly hard.
// Think about it for a while before reading the hint at the bottom of this file.
func FromRoman(s string) (int, error) {
	return 0, nil // TODO
}

// Hint for FromRoman (read only after 30 minutes of thinking):
// You already have a function that produces the ONE correct spelling of
// every number. Parse loosely, then use it to check your result.
