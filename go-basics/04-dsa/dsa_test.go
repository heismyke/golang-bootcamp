// DON'T EDIT THIS FILE. These tests check your answers.
package dsa

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"time"
)

// ─── helpers ──────────────────────────────────────────────────

// within fails the test if f is too slow, which usually means the
// algorithm has the wrong time complexity (or loops forever).
func within(t *testing.T, d time.Duration, hint string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() { defer close(done); f() }()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("too slow (over %v). %s", d, hint)
	}
}

func list(vals ...int) *ListNode {
	dummy := &ListNode{}
	cur := dummy
	for _, v := range vals {
		cur.Next = &ListNode{Val: v}
		cur = cur.Next
	}
	return dummy.Next
}

func listVals(head *ListNode) []int {
	out := []int{}
	for n := 0; head != nil; head = head.Next {
		if n++; n > 10_000 {
			return append(out, -999) // cycle guard
		}
		out = append(out, head.Val)
	}
	return out
}

func node(v int, l, r *TreeNode) *TreeNode { return &TreeNode{Val: v, Left: l, Right: r} }

func leaf(v int) *TreeNode { return &TreeNode{Val: v} }

// normalize sorts inner and outer slices so answers in any order compare equal.
func normalize[T interface{ ~int | ~string }](groups [][]T) [][]T {
	out := make([][]T, len(groups))
	for i, g := range groups {
		out[i] = slices.Sorted(slices.Values(g))
	}
	slices.SortFunc(out, slices.Compare)
	return out
}

func equal2D[T comparable](a, b [][]T) bool {
	return slices.EqualFunc(a, b, func(x, y []T) bool { return slices.Equal(x, y) })
}

func randomInts(n int) []int {
	r := rand.New(rand.NewPCG(1, 2))
	out := make([]int, n)
	for i := range out {
		out[i] = r.IntN(1_000_000) - 500_000
	}
	return out
}

// ─── 1. Arrays ────────────────────────────────────────────────

func TestArrays_TwoSum(t *testing.T) {
	i, j, ok := TwoSum([]int{2, 7, 11, 15}, 9)
	if !ok || i != 0 || j != 1 {
		t.Errorf("TwoSum([2 7 11 15], 9) = (%d, %d, %v), want (0, 1, true)", i, j, ok)
	}
	i, j, ok = TwoSum([]int{3, 3}, 6)
	if !ok || i != 0 || j != 1 {
		t.Errorf("TwoSum([3 3], 6) = (%d, %d, %v), want (0, 1, true)", i, j, ok)
	}
	if _, _, ok := TwoSum([]int{1, 2, 3}, 100); ok {
		t.Error("TwoSum with no answer returned ok = true")
	}
	if _, _, ok := TwoSum([]int{5}, 10); ok {
		t.Error("TwoSum must use two DIFFERENT indexes: [5] with target 10 has no answer")
	}

	big := make([]int, 200_000)
	for k := range big {
		big[k] = k * 2
	}
	within(t, 2*time.Second, "TwoSum must be O(n): use a map, not nested loops.", func() {
		i, j, ok := TwoSum(big, big[len(big)-2]+big[len(big)-1])
		if !ok || i != len(big)-2 || j != len(big)-1 {
			t.Errorf("big input: got (%d, %d, %v)", i, j, ok)
		}
	})
}

func TestArrays_IsPalindrome(t *testing.T) {
	cases := map[string]bool{
		"A man, a plan, a canal: Panama": true,
		"race a car":                     false,
		"":                               true,
		" ":                              true,
		"No 'x' in Nixon":                true,
		"0P":                             false,
	}
	for s, want := range cases {
		if got := IsPalindrome(s); got != want {
			t.Errorf("IsPalindrome(%q) = %v, want %v", s, got, want)
		}
	}
}

func TestArrays_MaxSubarraySum(t *testing.T) {
	cases := []struct {
		in   []int
		want int
	}{
		{[]int{-2, 1, -3, 4, -1, 2, 1, -5, 4}, 6},
		{[]int{1}, 1},
		{[]int{5, 4, -1, 7, 8}, 23},
		{[]int{-3, -1, -2}, -1},
	}
	for _, c := range cases {
		if got := MaxSubarraySum(c.in); got != c.want {
			t.Errorf("MaxSubarraySum(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestArrays_LongestUniqueSubstring(t *testing.T) {
	cases := map[string]int{"abcabcbb": 3, "bbbbb": 1, "pwwkew": 3, "": 0, "abcdef": 6, "abba": 2, "dvdf": 3}
	for s, want := range cases {
		if got := LongestUniqueSubstring(s); got != want {
			t.Errorf("LongestUniqueSubstring(%q) = %d, want %d", s, got, want)
		}
	}
}

func TestArrays_MergeSorted(t *testing.T) {
	cases := []struct{ a, b, want []int }{
		{[]int{1, 3, 5}, []int{2, 4, 6}, []int{1, 2, 3, 4, 5, 6}},
		{[]int{}, []int{1, 2}, []int{1, 2}},
		{[]int{1, 1, 9}, []int{1, 10}, []int{1, 1, 1, 9, 10}},
	}
	for _, c := range cases {
		if got := MergeSorted(c.a, c.b); !slices.Equal(got, c.want) {
			t.Errorf("MergeSorted(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestArrays_GroupAnagrams(t *testing.T) {
	got := normalize(GroupAnagrams([]string{"eat", "tea", "tan", "ate", "nat", "bat"}))
	want := normalize([][]string{{"eat", "tea", "ate"}, {"tan", "nat"}, {"bat"}})
	if !equal2D(got, want) {
		t.Errorf("GroupAnagrams = %v, want %v", got, want)
	}
}

// ─── 2. Stacks and queues ─────────────────────────────────────

func TestStacks_ValidBrackets(t *testing.T) {
	cases := map[string]bool{"()": true, "()[]{}": true, "([]{})": true, "(]": false, "([)]": false, "((": false, "))": false, "": true}
	for s, want := range cases {
		if got := ValidBrackets(s); got != want {
			t.Errorf("ValidBrackets(%q) = %v, want %v", s, got, want)
		}
	}
}

func TestStacks_EvalRPN(t *testing.T) {
	ok := []struct {
		in   string
		want int
	}{
		{"2 1 + 3 *", 9},
		{"4 13 5 / +", 6},
		{"10 6 9 3 + -11 * / * 17 + 5 +", 22},
		{"7", 7},
	}
	for _, c := range ok {
		got, err := EvalRPN(strings.Fields(c.in))
		if err != nil || got != c.want {
			t.Errorf("EvalRPN(%q) = (%d, %v), want (%d, nil)", c.in, got, err, c.want)
		}
	}
	for _, bad := range []string{"1 +", "1 2", "1 0 /", "1 2 x", ""} {
		if _, err := EvalRPN(strings.Fields(bad)); err == nil {
			t.Errorf("EvalRPN(%q) returned no error, want an error", bad)
		}
	}
}

func TestStacks_Queue(t *testing.T) {
	var q Queue[string]
	if _, ok := q.Dequeue(); ok {
		t.Error("Dequeue on empty queue returned ok = true")
	}
	q.Enqueue("a")
	q.Enqueue("b")
	if v, _ := q.Dequeue(); v != "a" {
		t.Errorf("Dequeue = %q, want a (first in, first out)", v)
	}
	q.Enqueue("c")
	if q.Len() != 2 {
		t.Errorf("Len = %d, want 2", q.Len())
	}
	for _, want := range []string{"b", "c"} {
		if v, ok := q.Dequeue(); v != want || !ok {
			t.Errorf("Dequeue = (%q, %v), want (%q, true)", v, ok, want)
		}
	}
	if q.Len() != 0 {
		t.Errorf("Len = %d, want 0", q.Len())
	}
}

func TestStacks_DailyTemperatures(t *testing.T) {
	got := DailyTemperatures([]int{73, 74, 75, 71, 69, 72, 76, 73})
	if want := []int{1, 1, 4, 2, 1, 1, 0, 0}; !slices.Equal(got, want) {
		t.Errorf("DailyTemperatures = %v, want %v", got, want)
	}
	big := make([]int, 200_000)
	for i := range big {
		big[i] = len(big) - i // always getting colder: worst case for O(n²)
	}
	within(t, 2*time.Second, "use a monotonic stack for O(n).", func() {
		got := DailyTemperatures(big)
		if len(got) != len(big) || slices.Max(got) != 0 {
			t.Error("big input: every answer should be 0")
		}
	})
}

// ─── 3. Linked lists ──────────────────────────────────────────

func TestLinkedList_Reverse(t *testing.T) {
	cases := []struct{ in, want []int }{
		{[]int{1, 2, 3, 4, 5}, []int{5, 4, 3, 2, 1}},
		{[]int{1}, []int{1}},
		{[]int{}, []int{}},
	}
	for _, c := range cases {
		if got := listVals(ReverseList(list(c.in...))); !slices.Equal(got, c.want) {
			t.Errorf("ReverseList(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestLinkedList_Middle(t *testing.T) {
	if m := MiddleNode(list(1, 2, 3, 4, 5)); m == nil || m.Val != 3 {
		t.Errorf("MiddleNode(1..5) = %v, want node 3", m)
	}
	if m := MiddleNode(list(1, 2, 3, 4)); m == nil || m.Val != 3 {
		t.Errorf("MiddleNode(1..4) = %v, want node 3 (second middle)", m)
	}
}

func TestLinkedList_HasCycle(t *testing.T) {
	if HasCycle(list(1, 2, 3)) || HasCycle(nil) {
		t.Error("HasCycle reported a cycle in a normal list")
	}
	l := list(1, 2, 3, 4)
	l.Next.Next.Next.Next = l.Next // 4 → 2
	within(t, time.Second, "your loop never ends on a cyclic list.", func() {
		if !HasCycle(l) {
			t.Error("HasCycle missed a cycle 1→2→3→4→2")
		}
	})
	self := list(1)
	self.Next = self
	within(t, time.Second, "your loop never ends on a node pointing to itself.", func() {
		if !HasCycle(self) {
			t.Error("HasCycle missed a node pointing to itself")
		}
	})
}

func TestLinkedList_Merge(t *testing.T) {
	got := listVals(MergeTwoLists(list(1, 2, 4), list(1, 3, 4)))
	if want := []int{1, 1, 2, 3, 4, 4}; !slices.Equal(got, want) {
		t.Errorf("MergeTwoLists = %v, want %v", got, want)
	}
	if got := listVals(MergeTwoLists(nil, list(0))); !slices.Equal(got, []int{0}) {
		t.Errorf("MergeTwoLists(nil, [0]) = %v, want [0]", got)
	}
}

func TestLinkedList_RemoveNthFromEnd(t *testing.T) {
	cases := []struct {
		in   []int
		n    int
		want []int
	}{
		{[]int{1, 2, 3, 4, 5}, 2, []int{1, 2, 3, 5}},
		{[]int{1}, 1, []int{}},
		{[]int{1, 2}, 2, []int{2}},
	}
	for _, c := range cases {
		if got := listVals(RemoveNthFromEnd(list(c.in...), c.n)); !slices.Equal(got, c.want) {
			t.Errorf("RemoveNthFromEnd(%v, %d) = %v, want %v", c.in, c.n, got, c.want)
		}
	}
}

// ─── 4. Searching and sorting ─────────────────────────────────

func TestSearchSort_BinarySearch(t *testing.T) {
	nums := []int{-5, 0, 3, 7, 9, 12, 20}
	for i, v := range nums {
		if got := BinarySearch(nums, v); got != i {
			t.Errorf("BinarySearch(%v, %d) = %d, want %d", nums, v, got, i)
		}
	}
	for _, v := range []int{-10, 1, 8, 100} {
		if got := BinarySearch(nums, v); got != -1 {
			t.Errorf("BinarySearch(%v, %d) = %d, want -1", nums, v, got)
		}
	}
	if got := BinarySearch(nil, 1); got != -1 {
		t.Errorf("BinarySearch(empty) = %d, want -1", got)
	}
}

func TestSearchSort_FirstTrue(t *testing.T) {
	const n = 1_000_000_000_000
	for _, target := range []int{0, 1, 777_777_777_777, n - 1, n} {
		calls := 0
		var got int
		within(t, time.Second, "FirstTrue must be binary search, not a loop.", func() {
			got = FirstTrue(n, func(i int) bool { calls++; return i >= target })
		})
		if got != target {
			t.Errorf("FirstTrue: got %d, want %d", got, target)
		}
		if calls > 45 {
			t.Errorf("FirstTrue called pred %d times, binary search needs about 40", calls)
		}
	}
}

func TestSearchSort_InsertionSort(t *testing.T) {
	nums := randomInts(2_000)
	want := slices.Sorted(slices.Values(nums))
	InsertionSort(nums)
	if !slices.Equal(nums, want) {
		t.Error("InsertionSort did not sort the slice")
	}
}

func TestSearchSort_MergeSort(t *testing.T) {
	if got := MergeSort([]int{5, 2, 9, 1, 5, 6}); !slices.Equal(got, []int{1, 2, 5, 5, 6, 9}) {
		t.Errorf("MergeSort small = %v", got)
	}
	if got := MergeSort([]int{}); len(got) != 0 {
		t.Errorf("MergeSort(empty) = %v", got)
	}
	nums := randomInts(300_000)
	want := slices.Sorted(slices.Values(nums))
	within(t, 2*time.Second, "MergeSort must be O(n log n).", func() {
		if !slices.Equal(MergeSort(nums), want) {
			t.Error("MergeSort big input not sorted correctly")
		}
	})
}

func TestSearchSort_QuickSort(t *testing.T) {
	small := []int{5, 2, 9, 1, 5, 6}
	QuickSort(small)
	if !slices.Equal(small, []int{1, 2, 5, 5, 6, 9}) {
		t.Errorf("QuickSort small = %v", small)
	}
	nums := randomInts(300_000)
	want := slices.Sorted(slices.Values(nums))
	within(t, 2*time.Second, "QuickSort is too slow on random input.", func() {
		QuickSort(nums)
		if !slices.Equal(nums, want) {
			t.Error("QuickSort random input not sorted correctly")
		}
	})
	sorted := make([]int, 300_000)
	for i := range sorted {
		sorted[i] = i
	}
	within(t, 2*time.Second, "QuickSort is O(n²) on already-sorted input. Choose a better pivot.", func() {
		QuickSort(sorted)
		if !slices.IsSorted(sorted) {
			t.Error("QuickSort broke an already sorted slice")
		}
	})
}

// ─── 5. Recursion ─────────────────────────────────────────────

func TestRecursion_Power(t *testing.T) {
	cases := [][3]int{{2, 10, 1024}, {3, 0, 1}, {5, 3, 125}, {-2, 3, -8}, {7, 1, 7}}
	for _, c := range cases {
		if got := Power(c[0], c[1]); got != c[2] {
			t.Errorf("Power(%d, %d) = %d, want %d", c[0], c[1], got, c[2])
		}
	}
	within(t, time.Second, "Power must be O(log exp): square and halve.", func() {
		if got := Power(1, 1_000_000_000_000); got != 1 {
			t.Errorf("Power(1, 10^12) = %d, want 1", got)
		}
	})
}

func TestRecursion_Permutations(t *testing.T) {
	got := normalizeOrder(Permutations([]int{1, 2, 3}))
	want := [][]int{{1, 2, 3}, {1, 3, 2}, {2, 1, 3}, {2, 3, 1}, {3, 1, 2}, {3, 2, 1}}
	if !equal2D(got, want) {
		t.Errorf("Permutations([1 2 3]) = %v, want %v", got, want)
	}
	if n := len(Permutations([]int{1, 2, 3, 4, 5})); n != 120 {
		t.Errorf("Permutations of 5 items: got %d, want 120", n)
	}
}

// normalizeOrder sorts the list of permutations but keeps each one's order.
func normalizeOrder(p [][]int) [][]int {
	out := slices.Clone(p)
	slices.SortFunc(out, slices.Compare)
	return out
}

func TestRecursion_Subsets(t *testing.T) {
	got := normalize(Subsets([]int{1, 2, 3}))
	want := normalize([][]int{{}, {1}, {2}, {3}, {1, 2}, {1, 3}, {2, 3}, {1, 2, 3}})
	if !equal2D(got, want) {
		t.Errorf("Subsets([1 2 3]) = %v, want %v", got, want)
	}
}

func TestRecursion_Hanoi(t *testing.T) {
	for _, n := range []int{1, 3, 6} {
		moves := Hanoi(n, "A", "C", "B")
		if len(moves) != 1<<n-1 {
			t.Errorf("Hanoi(%d): %d moves, want %d", n, len(moves), 1<<n-1)
			continue
		}
		pegs := map[string][]int{"A": {}, "B": {}, "C": {}}
		for d := n; d >= 1; d-- {
			pegs["A"] = append(pegs["A"], d)
		}
		for i, m := range moves {
			from, to, ok := strings.Cut(m, "->")
			src, dst := pegs[from], pegs[to]
			if !ok || len(src) == 0 || dst == nil {
				t.Fatalf("Hanoi(%d) move %d %q is invalid", n, i, m)
			}
			disk := src[len(src)-1]
			if len(dst) > 0 && dst[len(dst)-1] < disk {
				t.Fatalf("Hanoi(%d) move %d %q puts disk %d on smaller disk %d", n, i, m, disk, dst[len(dst)-1])
			}
			pegs[from], pegs[to] = src[:len(src)-1], append(dst, disk)
		}
		if len(pegs["C"]) != n {
			t.Errorf("Hanoi(%d): peg C has %d disks at the end, want %d", n, len(pegs["C"]), n)
		}
	}
}

// ─── 6. Trees ─────────────────────────────────────────────────

func buildBST(vals ...int) *TreeNode {
	var root *TreeNode
	for _, v := range vals {
		root = Insert(root, v)
	}
	return root
}

func TestTrees_InsertAndInOrder(t *testing.T) {
	root := buildBST(50, 30, 70, 20, 40, 60, 80, 30)
	if got := InOrder(root); !slices.Equal(got, []int{20, 30, 40, 50, 60, 70, 80}) {
		t.Errorf("InOrder after inserts = %v, want sorted without duplicates", got)
	}
	if root == nil || root.Val != 50 || root.Left == nil || root.Left.Val != 30 {
		t.Error("Insert didn't build the BST shape correctly (root should be 50, its left child 30)")
	}
	if got := InOrder(nil); len(got) != 0 {
		t.Errorf("InOrder(nil) = %v, want empty", got)
	}
}

func TestTrees_Contains(t *testing.T) {
	root := buildBST(50, 30, 70, 20, 40, 60, 80)
	for _, v := range []int{50, 20, 80, 40} {
		if !Contains(root, v) {
			t.Errorf("Contains(%d) = false, want true", v)
		}
	}
	for _, v := range []int{0, 45, 100} {
		if Contains(root, v) {
			t.Errorf("Contains(%d) = true, want false", v)
		}
	}
}

func TestTrees_Height(t *testing.T) {
	tree := node(3, leaf(9), node(20, leaf(15), node(7, nil, leaf(1))))
	cases := []struct {
		tree *TreeNode
		want int
	}{{nil, 0}, {leaf(1), 1}, {tree, 4}}
	for _, c := range cases {
		if got := Height(c.tree); got != c.want {
			t.Errorf("Height = %d, want %d", got, c.want)
		}
	}
}

func TestTrees_LevelOrder(t *testing.T) {
	tree := node(3, leaf(9), node(20, leaf(15), leaf(7)))
	if got := LevelOrder(tree); !equal2D(got, [][]int{{3}, {9, 20}, {15, 7}}) {
		t.Errorf("LevelOrder = %v, want [[3] [9 20] [15 7]]", got)
	}
	if got := LevelOrder(nil); len(got) != 0 {
		t.Errorf("LevelOrder(nil) = %v, want empty", got)
	}
}

func TestTrees_IsValidBST(t *testing.T) {
	cases := []struct {
		name string
		tree *TreeNode
		want bool
	}{
		{"empty", nil, true},
		{"valid", node(5, node(3, leaf(1), leaf(4)), node(8, nil, leaf(9))), true},
		{"bad child", node(5, leaf(6), leaf(8)), false},
		// 4 is a valid child of 6, but it's in 5's RIGHT subtree, so it must be > 5.
		{"bad grandchild", node(5, leaf(1), node(6, leaf(4), leaf(7))), false},
		{"duplicate", node(5, leaf(5), nil), false},
	}
	for _, c := range cases {
		if got := IsValidBST(c.tree); got != c.want {
			t.Errorf("IsValidBST(%s) = %v, want %v", c.name, got, c.want)
		}
	}
}

// ─── 7. Heaps ─────────────────────────────────────────────────

func TestHeap_MinHeap(t *testing.T) {
	var h MinHeap
	if _, ok := h.Pop(); ok {
		t.Error("Pop on empty heap returned ok = true")
	}
	if _, ok := h.Peek(); ok {
		t.Error("Peek on empty heap returned ok = true")
	}
	for _, v := range []int{5, 3, 8, 1, 9, 1} {
		h.Push(v)
	}
	if v, _ := h.Peek(); v != 1 || h.Len() != 6 {
		t.Errorf("Peek = %d, Len = %d; want 1, 6", v, h.Len())
	}
	var got []int
	for h.Len() > 0 {
		v, _ := h.Pop()
		got = append(got, v)
	}
	if !slices.Equal(got, []int{1, 1, 3, 5, 8, 9}) {
		t.Errorf("popping everything gave %v, want [1 1 3 5 8 9]", got)
	}

	nums := randomInts(300_000)
	within(t, 2*time.Second, "Push and Pop must be O(log n): sift up and sift down.", func() {
		var big MinHeap
		for _, v := range nums {
			big.Push(v)
		}
		prev := -1 << 62
		for big.Len() > 0 {
			v, _ := big.Pop()
			if v < prev {
				t.Error("big heap popped values out of order")
				return
			}
			prev = v
		}
	})
}

func TestHeap_KthLargest(t *testing.T) {
	cases := []struct {
		nums    []int
		k, want int
	}{
		{[]int{3, 2, 1, 5, 6, 4}, 2, 5},
		{[]int{3, 2, 3, 1, 2, 4, 5, 5, 6}, 4, 4},
		{[]int{7}, 1, 7},
	}
	for _, c := range cases {
		if got := KthLargest(c.nums, c.k); got != c.want {
			t.Errorf("KthLargest(%v, %d) = %d, want %d", c.nums, c.k, got, c.want)
		}
	}
}

func TestHeap_TopKFrequent(t *testing.T) {
	words := []string{"go", "rust", "go", "java", "rust", "go", "c", "c"}
	if got := TopKFrequent(words, 2); !slices.Equal(got, []string{"go", "c"}) {
		t.Errorf("TopKFrequent(k=2) = %v, want [go c] (c and rust tie at 2; c comes first alphabetically)", got)
	}
	if got := TopKFrequent(words, 1); !slices.Equal(got, []string{"go"}) {
		t.Errorf("TopKFrequent(k=1) = %v, want [go]", got)
	}
}

func TestHeap_MergeKSorted(t *testing.T) {
	got := MergeKSorted([][]int{{1, 4, 5}, {1, 3, 4}, {2, 6}, {}})
	if want := []int{1, 1, 2, 3, 4, 4, 5, 6}; !slices.Equal(got, want) {
		t.Errorf("MergeKSorted = %v, want %v", got, want)
	}
	if got := MergeKSorted(nil); len(got) != 0 {
		t.Errorf("MergeKSorted(nil) = %v, want empty", got)
	}
}

// ─── 8. Graphs ────────────────────────────────────────────────

func undirected(edges ...string) Graph {
	g := Graph{}
	for _, e := range edges {
		a, b, _ := strings.Cut(e, "-")
		g[a] = append(g[a], b)
		g[b] = append(g[b], a)
	}
	return g
}

func TestGraphs_ShortestPath(t *testing.T) {
	// A-B-D-F is 3 edges; A-C-E-F is also 3; A-C-F is the unique shortest (2).
	g := undirected("A-B", "B-D", "D-F", "A-C", "C-E", "E-F", "C-F", "X-Y")
	if got := ShortestPath(g, "A", "F"); !slices.Equal(got, []string{"A", "C", "F"}) {
		t.Errorf("ShortestPath(A, F) = %v, want [A C F]", got)
	}
	if got := ShortestPath(g, "A", "A"); !slices.Equal(got, []string{"A"}) {
		t.Errorf("ShortestPath(A, A) = %v, want [A]", got)
	}
	within(t, time.Second, "does your BFS mark nodes as visited?", func() {
		if got := ShortestPath(g, "A", "X"); got != nil {
			t.Errorf("ShortestPath(A, X) = %v, want nil (unreachable)", got)
		}
	})
}

func grid(rows ...string) [][]byte {
	out := make([][]byte, len(rows))
	for i, r := range rows {
		out[i] = []byte(r)
	}
	return out
}

func TestGraphs_NumIslands(t *testing.T) {
	cases := []struct {
		grid [][]byte
		want int
	}{
		{grid("11110", "11010", "11000", "00000"), 1},
		{grid("11000", "11000", "00100", "00011"), 3},
		{grid("101", "010", "101"), 5}, // diagonals don't connect
		{grid("000"), 0},
	}
	for i, c := range cases {
		if got := NumIslands(c.grid); got != c.want {
			t.Errorf("case %d: NumIslands = %d, want %d", i, got, c.want)
		}
	}
}

func TestGraphs_TopoSort(t *testing.T) {
	deps := map[string][]string{
		"app":    {"db", "cache", "config"},
		"db":     {"config"},
		"cache":  {"config"},
		"worker": {"db"},
	}
	order, err := TopoSort(deps)
	if err != nil {
		t.Fatalf("TopoSort returned error: %v", err)
	}
	pos := map[string]int{}
	for i, n := range order {
		pos[n] = i
	}
	if len(order) != 5 || len(pos) != 5 {
		t.Fatalf("TopoSort = %v, want each of the 5 items exactly once (including config)", order)
	}
	for item, needs := range deps {
		for _, d := range needs {
			if pos[d] > pos[item] {
				t.Errorf("TopoSort = %v: %s comes before its dependency %s", order, item, d)
			}
		}
	}

	within(t, time.Second, "cycle detection is looping forever.", func() {
		if _, err := TopoSort(map[string][]string{"a": {"b"}, "b": {"c"}, "c": {"a"}}); err == nil {
			t.Error("TopoSort with a cycle a→b→c→a returned no error")
		}
	})
}

func TestGraphs_Dijkstra(t *testing.T) {
	g := map[string]map[string]int{
		"A": {"B": 4, "C": 1},
		"C": {"B": 2, "D": 5},
		"B": {"D": 1},
		"E": {"A": 1}, // E can reach A, but A can't reach E
	}
	got := Dijkstra(g, "A")
	want := map[string]int{"A": 0, "C": 1, "B": 3, "D": 4}
	if !maps.Equal(got, want) {
		t.Errorf("Dijkstra(A) = %v, want %v", got, want)
	}
}

// ─── 9. Dynamic programming ───────────────────────────────────

func TestDP_Fib(t *testing.T) {
	cases := map[int]int{0: 0, 1: 1, 2: 1, 10: 55, 50: 12586269025}
	within(t, time.Second, "plain recursion is O(2^n). Memoize or use a loop.", func() {
		for n, want := range cases {
			if got := Fib(n); got != want {
				t.Errorf("Fib(%d) = %d, want %d", n, got, want)
			}
		}
		if got := Fib(90); got != 2880067194370816120 {
			t.Errorf("Fib(90) = %d, want 2880067194370816120", got)
		}
	})
}

func TestDP_ClimbStairs(t *testing.T) {
	cases := map[int]int{1: 1, 2: 2, 3: 3, 5: 8, 45: 1836311903}
	within(t, time.Second, "ClimbStairs needs DP, not plain recursion.", func() {
		for n, want := range cases {
			if got := ClimbStairs(n); got != want {
				t.Errorf("ClimbStairs(%d) = %d, want %d", n, got, want)
			}
		}
	})
}

func TestDP_CoinChange(t *testing.T) {
	cases := []struct {
		coins        []int
		amount, want int
	}{
		{[]int{1, 5, 10}, 27, 5},
		{[]int{1, 3, 4}, 6, 2}, // greedy would say 3
		{[]int{2}, 3, -1},
		{[]int{1}, 0, 0},
		{[]int{186, 419, 83, 408}, 6249, 20},
	}
	within(t, 2*time.Second, "CoinChange needs DP over amounts 0..amount.", func() {
		for _, c := range cases {
			if got := CoinChange(c.coins, c.amount); got != c.want {
				t.Errorf("CoinChange(%v, %d) = %d, want %d", c.coins, c.amount, got, c.want)
			}
		}
	})
}

func TestDP_LCS(t *testing.T) {
	cases := [][2]string{{"abcde", "ace"}, {"abc", "abc"}, {"abc", "def"}, {"", "abc"}, {"AGGTAB", "GXTXAYB"}}
	wants := []int{3, 3, 0, 0, 4}
	for i, c := range cases {
		if got := LCS(c[0], c[1]); got != wants[i] {
			t.Errorf("LCS(%q, %q) = %d, want %d", c[0], c[1], got, wants[i])
		}
	}
	a, b := strings.Repeat("ab", 1000), strings.Repeat("ba", 1000)
	within(t, 2*time.Second, "LCS needs a DP table: O(len(a)×len(b)).", func() {
		if got := LCS(a, b); got != 1999 {
			t.Errorf("LCS of long strings = %d, want 1999", got)
		}
	})
}

func TestDP_EditDistance(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"kitten", "sitting", 3},
		{"horse", "ros", 3},
		{"", "abc", 3},
		{"same", "same", 0},
		{"intention", "execution", 5},
	}
	for _, c := range cases {
		if got := EditDistance(c.a, c.b); got != c.want {
			t.Errorf("EditDistance(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestDP_Knapsack(t *testing.T) {
	cases := []struct {
		w, v      []int
		cap, want int
	}{
		{[]int{1, 3, 4, 5}, []int{1, 4, 5, 7}, 7, 9},
		{[]int{10}, []int{100}, 5, 0},
		{[]int{1, 2, 3}, []int{6, 10, 12}, 5, 22},
	}
	for _, c := range cases {
		if got := Knapsack(c.w, c.v, c.cap); got != c.want {
			t.Errorf("Knapsack(%v, %v, %d) = %d, want %d", c.w, c.v, c.cap, got, c.want)
		}
	}
}

// ─── 10. Design ───────────────────────────────────────────────

func TestDesign_LRUCache(t *testing.T) {
	c := NewLRUCache(2)
	c.Put(1, 1)
	c.Put(2, 2)
	if v, ok := c.Get(1); v != 1 || !ok {
		t.Errorf("Get(1) = (%d, %v), want (1, true)", v, ok)
	}
	c.Put(3, 3) // evicts 2, because 1 was used more recently
	if _, ok := c.Get(2); ok {
		t.Error("key 2 should have been evicted (least recently used)")
	}
	c.Put(1, 10) // update; also marks 1 as recently used
	c.Put(4, 4)  // evicts 3
	if _, ok := c.Get(3); ok {
		t.Error("key 3 should have been evicted")
	}
	if v, ok := c.Get(1); v != 10 || !ok {
		t.Errorf("Get(1) = (%d, %v), want (10, true) after update", v, ok)
	}
	if v, ok := c.Get(4); v != 4 || !ok {
		t.Errorf("Get(4) = (%d, %v), want (4, true)", v, ok)
	}

	within(t, 2*time.Second, "Get and Put must be O(1): map + doubly linked list.", func() {
		big := NewLRUCache(50_000)
		for i := range 150_000 {
			big.Put(i, i)
			big.Get(i / 2)
		}
		if _, ok := big.Get(149_999); !ok {
			t.Error("most recent key missing from big cache")
		}
	})
}

func TestDesign_Trie(t *testing.T) {
	tr := NewTrie()
	for _, w := range []string{"go", "golang", "gopher", "good", "rust", "ruby"} {
		tr.Insert(w)
	}
	for w, want := range map[string]bool{"go": true, "gopher": true, "gop": false, "rus": false, "ruby": true, "": false} {
		if got := tr.Search(w); got != want {
			t.Errorf("Search(%q) = %v, want %v", w, got, want)
		}
	}
	for p, want := range map[string]bool{"go": true, "gop": true, "r": true, "py": false, "golangs": false} {
		if got := tr.StartsWith(p); got != want {
			t.Errorf("StartsWith(%q) = %v, want %v", p, got, want)
		}
	}
	if got := tr.WithPrefix("go"); !slices.Equal(got, []string{"go", "golang", "good", "gopher"}) {
		t.Errorf("WithPrefix(go) = %v, want [go golang good gopher]", got)
	}
	if got := tr.WithPrefix("x"); len(got) != 0 {
		t.Errorf("WithPrefix(x) = %v, want empty", got)
	}
}

// Example of using fmt in tests (keeps the import used if you add debug prints).
var _ = fmt.Sprintf
