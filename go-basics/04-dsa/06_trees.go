package dsa

// ═════════════════════════════════════════════════════════════
// 6. BINARY TREES AND BINARY SEARCH TREES
// BST rule: everything in the left subtree < node < everything in the right.
// Most tree problems are: handle nil, recurse left, recurse right, combine.
// ═════════════════════════════════════════════════════════════

type TreeNode struct {
	Val         int
	Left, Right *TreeNode
}

// Insert adds v to the BST and returns the (possibly new) root. Ignore
// duplicates. Insert(nil, 5) returns a new single node.
func Insert(root *TreeNode, v int) *TreeNode {
	return nil // TODO
}

// Contains reports whether v is in the BST. Target: O(height). Only go
// left or right, never both.
func Contains(root *TreeNode, v int) bool {
	return false // TODO
}

// InOrder returns values in left, node, right order. For a BST that's
// sorted order.
func InOrder(root *TreeNode) []int {
	return nil // TODO
}

// Height returns the number of nodes on the longest root-to-leaf path.
// nil → 0, a single node → 1.
func Height(root *TreeNode) int {
	return 0 // TODO
}

// LevelOrder returns values level by level, top to bottom:
//
//	    3
//	   / \
//	  9   20       →  [[3], [9, 20], [15, 7]]
//	     /  \
//	    15   7
//
// This is breadth-first search (BFS) with a queue.
func LevelOrder(root *TreeNode) [][]int {
	return nil // TODO
}

// IsValidBST reports whether the tree follows the BST rule everywhere, not
// just between parent and child. Hint: pass down the allowed (min, max)
// range for each node.
func IsValidBST(root *TreeNode) bool {
	return false // TODO
}
