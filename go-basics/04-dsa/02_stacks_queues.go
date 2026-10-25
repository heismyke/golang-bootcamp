package dsa

// ═════════════════════════════════════════════════════════════
// 2. STACKS AND QUEUES
// Stack = last in, first out (a slice: append + remove last).
// Queue = first in, first out.
// ═════════════════════════════════════════════════════════════

// ValidBrackets reports whether every (, [, { is closed by the matching
// bracket in the right order. "([]{})" → true, "([)]" → false, "((" → false.
// Target: O(n) with a stack.
func ValidBrackets(s string) bool {
	return false // TODO
}

// EvalRPN evaluates Reverse Polish Notation: ["2","1","+","3","*"] means
// (2 + 1) * 3 = 9. Supports + - * / (integer division). Returns an error for
// invalid input (unknown token, not enough operands, leftover operands) or
// division by zero.
// Hint: numbers go on a stack; an operator pops two and pushes the result.
func EvalRPN(tokens []string) (int, error) {
	return 0, nil // TODO
}

// Queue is a FIFO queue built from TWO stacks: push onto `in`; when popping,
// if `out` is empty, move everything from `in` to `out` (this reverses the
// order). Every element moves at most once, so it's O(1) amortized.
type Queue[T any] struct {
	in, out []T
}

func (q *Queue[T]) Enqueue(v T) {
	// TODO
}

// Dequeue removes and returns the oldest element, or false if empty.
func (q *Queue[T]) Dequeue() (T, bool) {
	var zero T
	return zero, false // TODO
}

func (q *Queue[T]) Len() int {
	return 0 // TODO
}

// DailyTemperatures returns, for each day, how many days you'd wait for a
// warmer temperature (0 if never). [73,74,75,71,69,72,76,73] → [1,1,4,2,1,1,0,0]
// Target: O(n) with a "monotonic stack" of indexes still waiting for a
// warmer day. The test uses 200,000 days, so O(n²) will time out.
func DailyTemperatures(temps []int) []int {
	return nil // TODO
}
