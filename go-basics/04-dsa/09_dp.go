package dsa

// ═════════════════════════════════════════════════════════════
// 9. DYNAMIC PROGRAMMING
// DP = recursion + remembering answers to subproblems so you never compute
// the same thing twice. Steps: (1) define what dp[i] means, (2) write the
// recurrence, (3) set base cases, (4) fill the table in the right order.
// ═════════════════════════════════════════════════════════════

// Fib returns the n-th Fibonacci number: Fib(0)=0, Fib(1)=1, Fib(n)=Fib(n-1)+Fib(n-2).
// The test asks for Fib(90). Plain recursion makes about 2^90 calls and never
// finishes. Use memoization (a map or slice) or a loop.
func Fib(n int) int {
	return 0 // TODO
}

// ClimbStairs returns how many distinct ways there are to climb n steps
// taking 1 or 2 steps at a time. ClimbStairs(3) = 3: (1,1,1), (1,2), (2,1).
// Hint: how do you get to step n? From n-1 or from n-2.
func ClimbStairs(n int) int {
	return 0 // TODO
}

// CoinChange returns the fewest coins that add up to amount, or -1 if it's
// impossible. coins=[1,5,10], amount=27 → 5 (10+10+5+1+1).
// Careful: greedy (biggest coin first) fails for coins=[1,3,4], amount=6:
// greedy gives 4+1+1 = 3 coins, but 3+3 = 2 coins.
// dp[a] = fewest coins for amount a.
func CoinChange(coins []int, amount int) int {
	return 0 // TODO
}

// LCS returns the length of the longest common subsequence (characters in
// order, not necessarily adjacent). LCS("abcde", "ace") = 3.
// Classic 2D DP: dp[i][j] = LCS of a[:i] and b[:j].
func LCS(a, b string) int {
	return 0 // TODO
}

// EditDistance returns the minimum number of single-character inserts,
// deletes and replacements to turn a into b. "kitten" → "sitting" = 3.
// Used in spell checkers, diff tools, and fuzzy matching of names in fraud
// checks.
func EditDistance(a, b string) int {
	return 0 // TODO
}

// Knapsack returns the maximum total value of items that fit in capacity.
// Each item can be used at most once.
// weights=[1,3,4,5], values=[1,4,5,7], capacity=7 → 9 (items of weight 3 and 4).
func Knapsack(weights, values []int, capacity int) int {
	return 0 // TODO
}
