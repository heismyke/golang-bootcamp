// Package advanced: YOUR CODE GOES IN interfaces.go, generics.go and
// concurrency.go. Replace each "TODO" with your own code, then run:
//
//	go test ./03-advanced/ -v
//
// Run just one part with -run, for example:
//
//	go test ./03-advanced/ -v -run Interfaces
//
// Don't edit advanced_test.go.
package advanced

import (
	"io"
	"time"
)

// ═════════════════════════════════════════════════════════════
// PART 1: INTERFACES IN DEPTH
// ═════════════════════════════════════════════════════════════

// ─── 1. fmt.Stringer ──────────────────────────────────────────
// Any type with a `String() string` method prints that way with fmt.Println.
// That's the fmt.Stringer interface, one of the most used in Go.

type Yen int64

// String formats with a ¥ sign and thousands separators:
// Yen(1234567) prints as "¥1,234,567", Yen(0) as "¥0".
// Hint: strconv.FormatInt, then insert commas working from the right.
func (y Yen) String() string {
	return "" // TODO
}

// ─── 2. Custom error types ────────────────────────────────────
// The error interface is just: type error interface { Error() string }
// A struct error can carry data that callers inspect with errors.As.

type NotFoundError struct {
	Resource string
	ID       int
}

// Error returns a message like "expense 7 not found".
func (e *NotFoundError) Error() string {
	return "" // TODO
}

// FindExpense pretends to look up an expense. Only ID 1 exists. For any
// other ID, return a *NotFoundError with Resource "expense".
func FindExpense(id int) error {
	return nil // TODO
}

// ─── 3. Type switches ─────────────────────────────────────────
// `any` (= interface{}) can hold a value of any type. A type switch finds
// out which type it actually is: switch x := v.(type) { case int: ... }

// TypeName returns:
//
//	int      → "int: 5"
//	string   → "string: hi"
//	[]int    → "slice of 3 ints"
//	error    → "error: <the message>"
//	nil      → "nil"
//	anything else → "unknown"
func TypeName(v any) string {
	return "" // TODO
}

// ─── 4. Implementing io.Writer ────────────────────────────────
// io.Writer is: Write(p []byte) (n int, err error)
// Implement it, and your type works with fmt.Fprintf, io.Copy, loggers,
// HTTP responses... everything that writes.

// UpperWriter upper-cases everything written to it, then passes it on to W.
type UpperWriter struct {
	W io.Writer
}

// Write must return len(p) on success, even though it writes a modified copy.
// Hint: bytes.ToUpper
func (u UpperWriter) Write(p []byte) (int, error) {
	return 0, nil // TODO
}

// ─── 5. Implementing io.Reader ────────────────────────────────
// io.Reader is: Read(p []byte) (n int, err error)
// Fill p with up to len(p) bytes and return how many you wrote. When there's
// nothing left, return 0, io.EOF. Callers call Read again and again.

// RepeatReader produces the byte B exactly N times, then io.EOF.
// Notice the pointer receiver: Read has to remember how much is left.
type RepeatReader struct {
	B byte
	N int
}

func (r *RepeatReader) Read(p []byte) (int, error) {
	return 0, io.EOF // TODO
}

// ─── 6. Interfaces for testability (dependency injection) ─────
// Code that calls time.Now() directly is hard to test. If it asks for a
// Clock instead, tests can pass a fake clock.

type Clock interface {
	Now() time.Time
}

// Greeting returns "Good morning" before 12:00, "Good afternoon" from 12:00
// to 17:59, and "Good evening" from 18:00.
func Greeting(c Clock) string {
	return "" // TODO
}
