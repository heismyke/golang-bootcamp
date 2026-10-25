package dsa

// ═════════════════════════════════════════════════════════════
// 5. RECURSION AND BACKTRACKING
// Every recursive function needs: (1) a base case that stops, and (2) a
// recursive call on a smaller problem.
// ═════════════════════════════════════════════════════════════

// Power returns base^exp for exp >= 0.
// Target: O(log exp), "fast exponentiation": x^10 = (x^5)², x^5 = x·(x^2)².
// The test calls Power(1, 1_000_000_000_000), so a simple loop never finishes.
func Power(base, exp int) int {
	return 0 // TODO
}

// Permutations returns every ordering of nums (all distinct).
// [1,2,3] → [[1,2,3],[1,3,2],[2,1,3],[2,3,1],[3,1,2],[3,2,1]] in any order.
// Backtracking: choose an unused number, recurse, un-choose.
// Watch out: append the COPY of your current path, not the path itself. Why?
func Permutations(nums []int) [][]int {
	return nil // TODO
}

// Subsets returns every subset of nums (all distinct), including the empty
// one. [1,2] → [[],[1],[2],[1,2]] in any order. There are 2^n of them.
// Hint: for each element, recurse twice: once with it, once without.
func Subsets(nums []int) [][]int {
	return nil // TODO
}

// Hanoi solves the Tower of Hanoi: move n disks from peg `from` to peg `to`
// using `via`, never putting a bigger disk on a smaller one. Return the
// moves as strings like "A->C". It takes exactly 2^n - 1 moves.
// Hint: move n-1 disks out of the way, move the biggest, move n-1 back on top.
func Hanoi(n int, from, to, via string) []string {
	return nil // TODO
}
