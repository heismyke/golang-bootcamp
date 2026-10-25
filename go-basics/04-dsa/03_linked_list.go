package dsa

// ═════════════════════════════════════════════════════════════
// 3. LINKED LISTS
// Each node points to the next. Draw boxes and arrows on paper before
// coding. Almost every linked-list bug is a lost pointer.
// ═════════════════════════════════════════════════════════════

type ListNode struct {
	Val  int
	Next *ListNode
}

// ReverseList reverses the list and returns the new head.
// 1→2→3 becomes 3→2→1. Target: O(n) time, O(1) extra memory.
// Hint: three pointers: prev, cur, next.
func ReverseList(head *ListNode) *ListNode {
	return nil // TODO
}

// MiddleNode returns the middle node. For an even length, return the second
// of the two middles: 1→2→3→4 returns the node 3.
// Target: one pass. Hint: a slow pointer moves 1 step, a fast one moves 2.
func MiddleNode(head *ListNode) *ListNode {
	return nil // TODO
}

// HasCycle reports whether the list loops back on itself.
// Target: O(1) extra memory (no map). Look up "Floyd's tortoise and hare".
func HasCycle(head *ListNode) bool {
	return false // TODO
}

// MergeTwoLists merges two sorted lists into one sorted list by relinking
// the existing nodes. Hint: start from a dummy node.
func MergeTwoLists(a, b *ListNode) *ListNode {
	return nil // TODO
}

// RemoveNthFromEnd removes the n-th node from the end and returns the head.
// 1→2→3→4→5 with n=2 becomes 1→2→3→5. n is always valid.
// Hint: move one pointer n steps ahead, then move both until it hits the end.
func RemoveNthFromEnd(head *ListNode, n int) *ListNode {
	return nil // TODO
}
