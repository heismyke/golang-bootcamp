package advanced

// ═════════════════════════════════════════════════════════════
// PART 2: GENERICS
// ═════════════════════════════════════════════════════════════
// Type parameters ([T any]) let one function work for many types while
// keeping full type safety.

// ─── 7. Map ───────────────────────────────────────────────────
// Map applies f to every element: Map([]int{1,2}, strconv.Itoa) → []string{"1","2"}
func Map[T, U any](s []T, f func(T) U) []U {
	return nil // TODO
}

// ─── 8. Filter ────────────────────────────────────────────────
// Filter keeps the elements for which keep returns true.
func Filter[T any](s []T, keep func(T) bool) []T {
	return nil // TODO
}

// ─── 9. Reduce ────────────────────────────────────────────────
// Reduce combines all elements into one value, starting from init.
// Reduce([]int{1,2,3}, 0, func(acc, n int) int { return acc + n }) → 6
func Reduce[T, A any](s []T, init A, f func(A, T) A) A {
	var zero A
	return zero // TODO
}

// ─── 10. A generic type: Stack ────────────────────────────────

type Stack[T any] struct {
	items []T
}

// Push adds v on top.
func (s *Stack[T]) Push(v T) {
	// TODO
}

// Pop removes and returns the top item. On an empty stack it returns the
// zero value and false.
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	return zero, false // TODO
}

// Len returns how many items are on the stack.
func (s *Stack[T]) Len() int {
	return 0 // TODO
}

// ─── 11. Constraints ──────────────────────────────────────────
// A constraint limits which types T can be. ~int means "int, or any type
// whose underlying type is int" (like Yen, whose underlying type is int64).

type Number interface {
	~int | ~int64 | ~float64
}

// Sum adds up any slice of numbers: Sum([]Yen{100, 200}) → Yen(300).
func Sum[T Number](s []T) T {
	var zero T
	return zero // TODO
}
