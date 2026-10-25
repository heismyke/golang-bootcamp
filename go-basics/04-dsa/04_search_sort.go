package dsa

// ═════════════════════════════════════════════════════════════
// 4. SEARCHING AND SORTING
// Don't use sort.Ints or slices.Sort here. The point is to build them.
// ═════════════════════════════════════════════════════════════

// BinarySearch returns the index of target in the sorted slice, or -1.
// Target: O(log n). Watch out for off-by-one errors in lo/hi.
func BinarySearch(nums []int, target int) int {
	return 0 // TODO
}

// FirstTrue returns the smallest i in [0, n) where pred(i) is true, or n if
// none. pred is "monotonic": false, false, ..., false, true, true, ...
// This is the general form of binary search ("first bad version", "minimum
// capacity that works", ...). The test uses n = 1,000,000,000,000 and
// counts your calls, so you can't loop.
func FirstTrue(n int, pred func(int) bool) int {
	return 0 // TODO
}

// InsertionSort sorts nums in place. O(n²), but fast for small or nearly
// sorted input. It's what real sort libraries use for tiny slices.
func InsertionSort(nums []int) {
	// TODO
}

// MergeSort returns a new sorted slice: split in half, sort each half
// recursively, merge (reuse your MergeSorted!). Target: O(n log n), stable.
func MergeSort(nums []int) []int {
	return nil // TODO
}

// QuickSort sorts nums in place: pick a pivot, put smaller elements left and
// bigger ones right, recurse on each side. Target: O(n log n) on average.
// The test sorts an ALREADY SORTED slice of 300,000. If you always pick
// the first element as the pivot, that's O(n²) and times out. Why? Fix it
// with a random pivot (math/rand/v2) or the middle element.
func QuickSort(nums []int) {
	// TODO
}
