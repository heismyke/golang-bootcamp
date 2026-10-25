# Assignment 4: Data structures and algorithms

45 exercises in 10 topics, all checked by tests. Some tests use big inputs
and a time limit, so a brute-force answer that's too slow fails with a hint.
That mirrors real interviews: "it works" isn't enough, it has to be efficient.

## How to work

```bash
cd ~/go-basics
go test ./04-dsa/ -v -run Arrays            # one topic
go test ./04-dsa/ -v -run Arrays_TwoSum     # one exercise
```

For **every** exercise:
1. Solve 2–3 examples **on paper** first, by hand.
2. Write the brute-force idea in a comment and its Big-O.
3. Find the better approach (the hint names the pattern), then code it.
4. Once it passes, write its time and space complexity above the function:
   `// Time: O(n), Space: O(n)`.
5. Can't solve it after 30 minutes? Ask for a hint. After a hint and another
   30 minutes, look up the approach, understand it, delete your code, and
   write it again from memory the next day.

## Topics, in order

| # | File | Topic | Test name | Key patterns |
|---|---|---|---|---|
| 1 | `01_arrays.go` | Arrays, strings, hash maps | `Arrays` | hash map lookup, two pointers, sliding window, Kadane |
| 2 | `02_stacks_queues.go` | Stacks and queues | `Stacks` | stack for matching, monotonic stack, queue from 2 stacks |
| 3 | `03_linked_list.go` | Linked lists | `LinkedList` | pointer juggling, fast/slow pointers, dummy node |
| 4 | `04_search_sort.go` | Binary search, sorting | `SearchSort` | binary search on the answer, merge sort, quicksort |
| 5 | `05_recursion.go` | Recursion, backtracking | `Recursion` | divide and conquer, choose/explore/un-choose |
| 6 | `06_trees.go` | Binary trees, BSTs | `Trees` | recursive DFS, BFS with a queue, passing bounds down |
| 7 | `07_heap.go` | Heaps, priority queues | `Heap` | sift up/down, top-k, k-way merge |
| 8 | `08_graphs.go` | Graphs | `Graphs` | BFS shortest path, grid DFS, topological sort, Dijkstra |
| 9 | `09_dp.go` | Dynamic programming | `DP` | memoization, 1D and 2D tables, 0/1 knapsack |
| 10 | `10_design.go` | Data structure design | `Design` | LRU cache (map + linked list), trie |

## Big-O cheat sheet

| Complexity | Name | n = 1,000,000 means about | Example |
|---|---|---|---|
| O(1) | constant | 1 step | map lookup, slice index |
| O(log n) | logarithmic | 20 steps | binary search, heap push/pop |
| O(n) | linear | 1 million steps | one loop over the data |
| O(n log n) | linearithmic | 20 million steps | good sorting |
| O(n²) | quadratic | 1 trillion steps (too slow) | nested loops over the data |
| O(2ⁿ) | exponential | impossible | naive recursive Fibonacci, all subsets |

Rule of thumb: about 100 million simple operations per second.

## Go operation costs to know

| Operation | Cost |
|---|---|
| `m[k]`, `m[k] = v`, `delete(m, k)` | O(1) average |
| `s[i]`, `len(s)`, `s = s[:len(s)-1]` | O(1) |
| `append(s, v)` | O(1) amortized |
| `s = s[1:]` | O(1), but the old memory stays reachable until the whole slice is gone |
| inserting or deleting in the middle of a slice | O(n) |
| `slices.Contains`, `strings.Contains` | O(n) |
| `s1 + s2` on strings | O(len), so building strings in a loop is O(n²). Use `strings.Builder` |
| `slices.Sort` | O(n log n) |

## After this

Practise with timed problems on [NeetCode 150](https://neetcode.io/practice)
(same topics, same order, video explanations), solving them **in Go**. Aim
for 2–3 problems a day. Re-solve any problem you needed help with a week later.
