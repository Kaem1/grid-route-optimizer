package optimization

import (
	"fmt"
	"sort"

	"grid-route-optimizer/internal/graph"
)

// TreeKind selects how BuildTree spans a region (algorithm spec section 6):
// TreeBFS produces a shallow tree (used when routes return to start, where
// the walk cost doesn't depend on tree shape); TreeDFS produces a deep tree
// via Warnsdorff's rule, which is cheaper to walk without returning.
type TreeKind int

const (
	TreeBFS TreeKind = iota
	TreeDFS
)

// BuildTree computes a spanning tree of the subgraph induced by region,
// rooted at root, represented as a node -> children adjacency map. It
// returns an error if root cannot reach every vertex of region.
func BuildTree(g *graph.Graph, region map[graph.VertexID]bool, root graph.VertexID, kind TreeKind) (map[graph.VertexID][]graph.VertexID, error) {
	children := make(map[graph.VertexID][]graph.VertexID)
	visited := map[graph.VertexID]bool{root: true}

	switch kind {
	case TreeBFS:
		queue := []graph.VertexID{root}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			for _, n := range g.Neighbors(current) {
				if !region[n] || visited[n] {
					continue
				}
				visited[n] = true
				children[current] = append(children[current], n)
				queue = append(queue, n)
			}
		}
	case TreeDFS:
		stack := []graph.VertexID{root}
		for len(stack) > 0 {
			current := stack[len(stack)-1]
			next, ok := warnsdorffNext(g, region, visited, current)
			if !ok {
				stack = stack[:len(stack)-1]
				continue
			}
			visited[next] = true
			children[current] = append(children[current], next)
			stack = append(stack, next)
		}
	}

	if len(visited) != len(region) {
		return nil, fmt.Errorf("region rooted at %q is not connected: reached %d of %d cells", root, len(visited), len(region))
	}

	return children, nil
}

// warnsdorffNext picks the next unvisited in-region neighbor of current
// with the fewest unvisited in-region neighbors of its own, tie-broken by
// the ascending order Neighbors already returns (spec section 6).
func warnsdorffNext(g *graph.Graph, region, visited map[graph.VertexID]bool, current graph.VertexID) (graph.VertexID, bool) {
	type candidate struct {
		id        graph.VertexID
		unvisited int
	}

	var candidates []candidate
	for _, n := range g.Neighbors(current) {
		if !region[n] || visited[n] {
			continue
		}
		count := 0
		for _, nn := range g.Neighbors(n) {
			if region[nn] && !visited[nn] {
				count++
			}
		}
		candidates = append(candidates, candidate{id: n, unvisited: count})
	}
	if len(candidates) == 0 {
		return "", false
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].unvisited < candidates[j].unvisited
	})
	return candidates[0].id, true
}

// WalkTree turns a spanning tree into a single continuous walk starting and
// ending at root (algorithm spec section 6). With returnToStart it performs
// a full Euler tour (cost 2(n-1)); otherwise it descends the deepest child
// last at every node and skips the final return step along that chain
// (cost 2(n-1) - height(root)), per the "on_last_chain" rule. Implemented
// iteratively (explicit stack) since tree depth may be large.
func WalkTree(children map[graph.VertexID][]graph.VertexID, root graph.VertexID, returnToStart bool) []graph.VertexID {
	if !returnToStart {
		heights := treeHeights(children, root)
		sortChildrenByHeight(children, heights)
	}

	type frame struct {
		node        graph.VertexID
		idx         int
		onLastChain bool
		isLastChild bool
	}

	walk := []graph.VertexID{root}
	stack := []frame{{node: root, onLastChain: true}}

	for len(stack) > 0 {
		top := &stack[len(stack)-1]
		kids := children[top.node]

		if top.idx < len(kids) {
			childIdx := top.idx
			top.idx++
			isLast := childIdx == len(kids)-1
			child := kids[childIdx]

			walk = append(walk, child)
			stack = append(stack, frame{
				node:        child,
				onLastChain: isLast && top.onLastChain,
				isLastChild: isLast,
			})
			continue
		}

		finished := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if len(stack) == 0 {
			break
		}
		if returnToStart || !(finished.isLastChild && finished.onLastChain) {
			walk = append(walk, stack[len(stack)-1].node)
		}
	}

	return walk
}

// treeHeights computes, for every node in the tree, 0 for a leaf or
// 1 + max(child heights) otherwise. Implemented iteratively via an
// explicit post-order stack.
func treeHeights(children map[graph.VertexID][]graph.VertexID, root graph.VertexID) map[graph.VertexID]int {
	heights := make(map[graph.VertexID]int)

	type frame struct {
		node graph.VertexID
		idx  int
	}

	stack := []frame{{node: root}}
	for len(stack) > 0 {
		top := &stack[len(stack)-1]
		kids := children[top.node]

		if top.idx < len(kids) {
			child := kids[top.idx]
			top.idx++
			stack = append(stack, frame{node: child})
			continue
		}

		height := 0
		for _, c := range kids {
			if h := heights[c] + 1; h > height {
				height = h
			}
		}
		heights[top.node] = height
		stack = stack[:len(stack)-1]
	}

	return heights
}

// sortChildrenByHeight orders each node's children ascending by height (tie
// broken by the ascending order they already have) so the deepest child is
// visited last, per spec section 6.3.
func sortChildrenByHeight(children map[graph.VertexID][]graph.VertexID, heights map[graph.VertexID]int) {
	for node, kids := range children {
		sort.SliceStable(kids, func(i, j int) bool {
			return heights[kids[i]] < heights[kids[j]]
		})
		children[node] = kids
	}
}

// WalkRegion computes a covering walk for region starting (and, if
// returnToStart, ending) at root: a DFS spanning tree (Warnsdorff-ordered)
// when routes don't return to start, a BFS spanning tree otherwise (tree
// shape doesn't affect cost in that mode), walked via WalkTree.
func WalkRegion(g *graph.Graph, region map[graph.VertexID]bool, root graph.VertexID, returnToStart bool) ([]graph.VertexID, error) {
	kind := TreeBFS
	if !returnToStart {
		kind = TreeDFS
	}

	tree, err := BuildTree(g, region, root, kind)
	if err != nil {
		return nil, err
	}

	return WalkTree(tree, root, returnToStart), nil
}
