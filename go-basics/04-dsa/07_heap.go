package dsa

// ═════════════════════════════════════════════════════════════
// 7. HEAPS AND PRIORITY QUEUES
// A min-heap always gives you the smallest element in O(log n). It's a
// binary tree stored in a slice: for index i, the children are 2i+1 and
// 2i+2, and the parent is (i-1)/2.
// Build MinHeap yourself (no container/heap). After that you may use
// container/heap or your MinHeap in the other functions.
// ═════════════════════════════════════════════════════════════

type MinHeap struct {
	data []int
}

// Push adds v. Target: O(log n). Append at the end, then "sift up": swap
// with the parent while smaller than it.
func (h *MinHeap) Push(v int) {
	// TODO
}

// Pop removes and returns the smallest element. Target: O(log n). Move
// the last element to the root, then "sift down": swap with the smaller
// child while bigger than it.
func (h *MinHeap) Pop() (int, bool) {
	return 0, false // TODO
}

// Peek returns the smallest element without removing it.
func (h *MinHeap) Peek() (int, bool) {
	return 0, false // TODO
}

func (h *MinHeap) Len() int {
	return 0 // TODO
}

// KthLargest returns the k-th largest element: [3,2,1,5,6,4] with k=2 → 5.
// Target: O(n log k). Keep a min-heap of the k biggest seen so far.
func KthLargest(nums []int, k int) int {
	return 0 // TODO
}

// TopKFrequent returns the k most frequent words, most frequent first; ties
// in alphabetical order.
// ["go","rust","go","java","rust","go"], k=2 → ["go","rust"]
func TopKFrequent(words []string, k int) []string {
	return nil // TODO
}

// MergeKSorted merges k sorted slices into one sorted slice.
// Target: O(N log k). Put the first element of each list in a heap, and
// keep popping the smallest and pushing the next from the same list.
func MergeKSorted(lists [][]int) []int {
	return nil // TODO
}
