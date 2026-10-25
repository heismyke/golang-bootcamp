// Package dsa: data structures and algorithms. YOUR CODE GOES IN THE
// NUMBERED FILES (01_arrays.go ... 10_design.go). Replace each "TODO".
//
// Run one topic at a time:
//
//	go test ./04-dsa/ -v -run Arrays
//
// Topics: Arrays, Stacks, LinkedList, SearchSort, Recursion, Trees, Heap,
// Graphs, DP, Design. Read 04-dsa/README.md first.
//
// Some tests use big inputs: a slow brute-force answer will time out. Each
// function says the target time complexity. Don't edit dsa_test.go.
package dsa

// ═════════════════════════════════════════════════════════════
// 1. ARRAYS, STRINGS AND HASH MAPS
// Patterns: hash map lookup, two pointers, sliding window
// ═════════════════════════════════════════════════════════════

// TwoSum finds two different indexes i < j with nums[i] + nums[j] == target.
// Returns ok=false if there are none.
// Target: O(n). Hint: a map from value to index. For each number, is
// (target - number) already in the map?
func TwoSum(nums []int, target int) (i, j int, ok bool) {
	return 0, 0, false // TODO
}

// IsPalindrome reports whether s reads the same backwards, looking only at
// letters and digits and ignoring case.
// "A man, a plan, a canal: Panama" → true.
// Target: O(n), two pointers moving inwards. Hint: unicode.IsLetter, unicode.IsDigit
func IsPalindrome(s string) bool {
	return false // TODO
}

// MaxSubarraySum returns the largest sum of any non-empty run of consecutive
// elements. [-2,1,-3,4,-1,2,1,-5,4] → 6 (from [4,-1,2,1]).
// Target: O(n). Look up "Kadane's algorithm". nums is never empty.
func MaxSubarraySum(nums []int) int {
	return 0 // TODO
}

// LongestUniqueSubstring returns the length of the longest substring with no
// repeated characters. "abcabcbb" → 3 ("abc"), "pwwkew" → 3 ("wke").
// Target: O(n), sliding window: grow the right side, shrink the left side
// when you see a repeat.
func LongestUniqueSubstring(s string) int {
	return 0 // TODO
}

// MergeSorted merges two sorted slices into one sorted slice.
// Target: O(len(a)+len(b)), two pointers. Don't just append and sort.
func MergeSorted(a, b []int) []int {
	return nil // TODO
}

// GroupAnagrams groups words that use the same letters:
// ["eat","tea","tan","ate","nat","bat"] → [["eat","tea","ate"],["tan","nat"],["bat"]]
// Order of groups and of words inside a group doesn't matter.
// Hint: what key do all anagrams share? (sort the letters, or count them)
func GroupAnagrams(words []string) [][]string {
	return nil // TODO
}
