package dsa

// ═════════════════════════════════════════════════════════════
// 8. GRAPHS
// A graph is nodes plus edges. Store it as an "adjacency list": a map from
// each node to its neighbours. BFS uses a queue (finds shortest paths in
// unweighted graphs); DFS uses recursion or a stack. Always track visited
// nodes, or cycles will loop forever.
// ═════════════════════════════════════════════════════════════

// Graph is an undirected graph: if "a" lists "b", then "b" lists "a".
type Graph map[string][]string

// ShortestPath returns the path with the fewest edges from start to end,
// including both ends, e.g. ["A","C","E"]. Returns nil if end is
// unreachable, and [start] if start == end.
// Hint: BFS, and remember each node's parent so you can rebuild the path.
func ShortestPath(g Graph, start, end string) []string {
	return nil // TODO
}

// NumIslands counts groups of connected '1' cells (up, down, left, right)
// in a grid of '1' (land) and '0' (water). You may modify the grid.
// Hint: for every '1' you find, count it, then DFS to sink the whole island.
func NumIslands(grid [][]byte) int {
	return 0 // TODO
}

// TopoSort returns an order in which every item comes AFTER everything it
// depends on. deps["app"] = ["db", "cache"] means app needs db and cache
// first. Include items that only appear as dependencies. Return an error if
// there's a cycle (a needs b, b needs a).
// This is how build tools, package managers and migration runners order work.
// Hint: DFS with three states (unvisited, visiting, done), or Kahn's
// algorithm with in-degree counts.
func TopoSort(deps map[string][]string) ([]string, error) {
	return nil, nil // TODO
}

// Dijkstra returns the shortest distance from src to every node it can
// reach in a directed graph with non-negative weights.
// g["A"]["B"] = 4 means an edge A→B of cost 4. Unreachable nodes are left
// out of the result. Target: O((V+E) log V) with a min-heap of (dist, node).
func Dijkstra(g map[string]map[string]int, src string) map[string]int {
	return nil // TODO
}
