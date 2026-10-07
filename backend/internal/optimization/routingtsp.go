package optimization

import (
	"fmt"
	"sort"

	"grid-route-optimizer/internal/graph"
)

// maxTSPRegionSize caps how large a region RouteRegion will run
// nearest-neighbor + 2-opt/Or-opt on (O(n^2) distance matrix). Larger
// regions fall back to the cheap spanning-tree walk so the app stays
// responsive on very large boards (algorithms_patch.md "Skalowalność" notes
// a KD-tree/lazy-BFS alternative for that case; not implemented here since
// this interactive app's boards rarely need it).
const maxTSPRegionSize = 2000

// tspNeighborListCount is the "k" in build_neighbor_lists.
const tspNeighborListCount = 16

// tspPerturbationRestarts is how many double-bridge perturbation rounds
// RouteRegion tries after the first 2-opt/Or-opt convergence.
const tspPerturbationRestarts = 6

// RouteRegion computes the final, reported route for a region
// (algorithms_patch.md "Zamiana sekcji 6", level 2): a nearest-neighbor
// tour over the region's shortest-path distance closure, refined by
// alternating 2-opt and Or-opt (with a few double-bridge perturbation
// restarts to escape local optima), then expanded back into an actual
// graph walk. This replaces WalkRegion's spanning-tree doubling (still
// used as the cheap "level 1" surrogate cost inside search loops) as the
// route actually reported to the user. The metric-closure TSP tour treats
// every cell as a waypoint visited exactly once, so on "thin ring with
// pinch points" topologies it can occasionally revisit cells twice where a
// tree-doubling walk wouldn't need to — RouteRegion falls back to
// WalkRegion's result whenever it happens to be cheaper, so callers can
// always rely on RouteRegion being at least as good as WalkRegion.
func RouteRegion(g *graph.Graph, region map[graph.VertexID]bool, root graph.VertexID, returnToStart bool) ([]graph.VertexID, error) {
	fallback, err := WalkRegion(g, region, root, returnToStart)
	if err != nil {
		return nil, err
	}
	if len(region) <= 2 || len(region) > maxTSPRegionSize {
		return fallback, nil
	}

	d, cells, err := regionDistanceMatrix(g, region, root)
	if err != nil {
		return nil, err
	}
	neighborLists := buildNeighborLists(d, tspNeighborListCount)

	best := optimizeTour(constructNearestNeighborTour(d), d, returnToStart, neighborLists)
	bestCost := tourCost(best, d, returnToStart)

	// Plain 2-opt/Or-opt can get stuck in a local optimum a simple
	// "double bridge" 4-opt move would easily escape (neither 2-opt nor
	// chain relocation can undo it in one step). A few deterministic,
	// differently-shaped perturbations followed by re-optimization make
	// RouteRegion far more reliably at least as good as WalkRegion.
	for restart := 1; restart <= tspPerturbationRestarts; restart++ {
		perturbed := doubleBridge(best, returnToStart, restart)
		candidate := optimizeTour(perturbed, d, returnToStart, neighborLists)
		if cost := tourCost(candidate, d, returnToStart); cost < bestCost {
			best, bestCost = candidate, cost
		}
	}

	if returnToStart {
		best = rotateToFront(best, 0)
	}

	routed := expandTour(g, region, cells, best, returnToStart)
	if len(routed)-1 >= len(fallback)-1 {
		return fallback, nil
	}
	return routed, nil
}

// optimizeTour alternates 2-opt and Or-opt until neither improves the tour
// further.
func optimizeTour(tour []int, d [][]int, closed bool, neighborLists [][]int) []int {
	for {
		before := tourCost(tour, d, closed)
		tour = twoOpt(tour, d, closed, neighborLists)
		tour = orOpt(tour, d, closed, neighborLists)
		if tourCost(tour, d, closed) >= before {
			return tour
		}
	}
}

// doubleBridge performs a deterministic 4-opt "double bridge" perturbation
// (segments A B C D -> A C B D): the classic move used to escape 2-opt/
// Or-opt local optima, since it can't be undone by either in a single
// step. variant shifts the split points so different restarts explore
// different perturbations while staying fully deterministic. For open
// tours, index 0 (the root) is never moved.
func doubleBridge(tour []int, closed bool, variant int) []int {
	n := len(tour)
	base := 0
	if !closed {
		base = 1
	}
	span := n - base
	if span < 8 {
		return append([]int(nil), tour...)
	}

	shift := (variant * span) / (tspPerturbationRestarts + 1)
	positions := []int{
		base + (span/4+shift)%span,
		base + (span/2+shift)%span,
		base + (3*span/4+shift)%span,
	}
	sort.Ints(positions)
	p1, p2, p3 := positions[0], positions[1], positions[2]
	if p1 <= base || p2 <= p1 || p3 <= p2 || p3 >= n {
		return append([]int(nil), tour...)
	}

	result := make([]int, 0, n)
	result = append(result, tour[:base]...)
	result = append(result, tour[base:p1]...)
	result = append(result, tour[p2:p3]...)
	result = append(result, tour[p1:p2]...)
	result = append(result, tour[p3:]...)
	return result
}

// regionDistanceMatrix computes the all-pairs shortest-path distance
// matrix for a region, restricted to edges inside the region. cells[0] is
// always root, so local index 0 always represents it.
func regionDistanceMatrix(g *graph.Graph, region map[graph.VertexID]bool, root graph.VertexID) ([][]int, []graph.VertexID, error) {
	cells := make([]graph.VertexID, 0, len(region))
	cells = append(cells, root)
	for _, v := range sortedCells(g, region) {
		if v != root {
			cells = append(cells, v)
		}
	}

	n := len(cells)
	d := make([][]int, n)
	for i, from := range cells {
		dist := regionBFSDistances(g, region, from)
		row := make([]int, n)
		for j, to := range cells {
			dj, ok := dist[to]
			if !ok {
				return nil, nil, fmt.Errorf("region rooted at %q is not connected: %q unreachable from %q", root, to, from)
			}
			row[j] = dj
		}
		d[i] = row
	}
	return d, cells, nil
}

// regionBFSDistances is a BFS restricted to region's induced subgraph.
func regionBFSDistances(g *graph.Graph, region map[graph.VertexID]bool, start graph.VertexID) map[graph.VertexID]int {
	dist := map[graph.VertexID]int{start: 0}
	queue := []graph.VertexID{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, n := range g.Neighbors(cur) {
			if !region[n] {
				continue
			}
			if _, seen := dist[n]; seen {
				continue
			}
			dist[n] = dist[cur] + 1
			queue = append(queue, n)
		}
	}
	return dist
}

// constructNearestNeighborTour builds an initial tour (local indices,
// starting at 0) by always stepping to the closest unvisited city.
func constructNearestNeighborTour(d [][]int) []int {
	n := len(d)
	tour := make([]int, 0, n)
	visited := make([]bool, n)
	tour = append(tour, 0)
	visited[0] = true
	cur := 0
	for len(tour) < n {
		best, bestDist := -1, 0
		for v := 0; v < n; v++ {
			if visited[v] {
				continue
			}
			if best == -1 || d[cur][v] < bestDist {
				best, bestDist = v, d[cur][v]
			}
		}
		tour = append(tour, best)
		visited[best] = true
		cur = best
	}
	return tour
}

// tourCost sums the tour's edge distances, closing the loop back to the
// start when closed is true.
func tourCost(tour []int, d [][]int, closed bool) int {
	cost := 0
	for i := 0; i+1 < len(tour); i++ {
		cost += d[tour[i]][tour[i+1]]
	}
	if closed && len(tour) > 1 {
		cost += d[tour[len(tour)-1]][tour[0]]
	}
	return cost
}

// buildNeighborLists returns, for every city, its k nearest other cities
// ascending by distance (tie broken by id) — the 2-opt/Or-opt candidate
// lists.
func buildNeighborLists(d [][]int, k int) [][]int {
	n := len(d)
	if k > n-1 {
		k = n - 1
	}
	lists := make([][]int, n)
	for i := 0; i < n; i++ {
		type candidate struct{ id, dist int }
		candidates := make([]candidate, 0, n-1)
		for j := 0; j < n; j++ {
			if j != i {
				candidates = append(candidates, candidate{j, d[i][j]})
			}
		}
		sort.Slice(candidates, func(a, b int) bool {
			if candidates[a].dist != candidates[b].dist {
				return candidates[a].dist < candidates[b].dist
			}
			return candidates[a].id < candidates[b].id
		})
		if k < len(candidates) {
			candidates = candidates[:k]
		}
		list := make([]int, len(candidates))
		for idx, c := range candidates {
			list[idx] = c.id
		}
		lists[i] = list
	}
	return lists
}

// twoOpt runs neighbor-list 2-opt with don't-look bits (Johnson & McGeoch
// style). For open tours, local index 0 (the root) never moves — every
// reversal this function performs excludes position 0 by construction. For
// closed tours the cycle is processed freely and re-anchored by the caller
// (RouteRegion rotates index 0 back to the front afterward).
func twoOpt(tour []int, d [][]int, closed bool, neighborLists [][]int) []int {
	n := len(tour)
	if n < 4 {
		return tour
	}
	tour = append([]int(nil), tour...)
	pos := make([]int, n)
	for i, c := range tour {
		pos[c] = i
	}
	dontLook := make([]bool, n)
	inQueue := make([]bool, n)
	queue := make([]int, n)
	copy(queue, tour)
	for i := range inQueue {
		inQueue[i] = true
	}

	stepPos := func(i, dir int) (int, bool) {
		j := i + dir
		if j >= 0 && j < n {
			return j, true
		}
		if closed {
			return (j + n) % n, true
		}
		return 0, false
	}

	reverse := func(from, to int) {
		length := to - from
		if length < 0 {
			length += n
		}
		length++
		for s := 0; s < length/2; s++ {
			a := (from + s) % n
			b := (to - s + n) % n
			tour[a], tour[b] = tour[b], tour[a]
			pos[tour[a]] = a
			pos[tour[b]] = b
		}
	}

	activate := func(cities ...int) {
		for _, c := range cities {
			dontLook[c] = false
			if !inQueue[c] {
				queue = append(queue, c)
				inQueue[c] = true
			}
		}
	}

	for len(queue) > 0 {
		c1 := queue[0]
		queue = queue[1:]
		inQueue[c1] = false
		if dontLook[c1] {
			continue
		}

		improved := false
		for _, dir := range [2]int{1, -1} {
			i := pos[c1]
			i2, ok := stepPos(i, dir)
			if !ok {
				continue
			}
			c2 := tour[i2]
			d12 := d[c1][c2]

			for _, c3 := range neighborLists[c1] {
				if d[c1][c3] >= d12 {
					break // neighbor list sorted ascending: no candidate beyond this can help
				}
				k := pos[c3]
				k2, ok := stepPos(k, dir)
				if !ok {
					continue
				}
				c4 := tour[k2]
				if c4 == c1 || c3 == c2 {
					continue
				}

				gain := d12 + d[c3][c4] - d[c1][c3] - d[c2][c4]
				if gain <= 0 {
					continue
				}

				e1, e2 := i, k
				if dir == -1 {
					e1, e2 = i2, k2
				}
				if !closed && e1 > e2 {
					e1, e2 = e2, e1
				}
				reverse((e1+1)%n, e2)

				activate(c1, c2, c3, c4)
				improved = true
				break
			}
			if improved {
				break
			}
		}

		if !improved {
			dontLook[c1] = true
		}
	}

	return tour
}

// orOpt repeatedly relocates chains of 1-3 consecutive cities (in either
// orientation) next to one of the chain's nearest neighbors, whenever that
// lowers total cost. For open tours, a chain touching either tour endpoint
// is left in place so index 0 (the root) never moves.
func orOpt(tour []int, d [][]int, closed bool, neighborLists [][]int) []int {
	tour = append([]int(nil), tour...)
	n := len(tour)

	improvedAny := true
	for improvedAny {
		improvedAny = false
		pos := make([]int, n)
		for i, c := range tour {
			pos[c] = i
		}

		for chainLen := 1; chainLen <= 3 && chainLen+2 <= n; chainLen++ {
			for start := 0; start < n; start++ {
				end := start + chainLen - 1
				if end >= n {
					continue // keep chains contiguous without wraparound
				}
				if !closed && (start == 0 || end == n-1) {
					continue
				}

				prevPos, nextPos := start-1, end+1
				if prevPos < 0 {
					prevPos += n
				}
				if nextPos >= n {
					nextPos -= n
				}
				prev, next := tour[prevPos], tour[nextPos]
				first, last := tour[start], tour[end]
				removeGain := d[prev][first] + d[last][next] - d[prev][next]

				inChain := make(map[int]bool, chainLen)
				for _, c := range tour[start : end+1] {
					inChain[c] = true
				}

				bestDelta, bestAfter, bestReversed := 0, -1, false
				for _, c3 := range neighborLists[first] {
					if inChain[c3] || c3 == prev {
						continue
					}
					j := pos[c3]
					jNext := j + 1
					if jNext >= n {
						if !closed {
							continue
						}
						jNext = 0
					}
					c3next := tour[jNext]
					if inChain[c3next] {
						continue
					}

					addFwd := d[c3][first] + d[last][c3next] - d[c3][c3next]
					if delta := addFwd - removeGain; delta < bestDelta {
						bestDelta, bestAfter, bestReversed = delta, c3, false
					}
					addRev := d[c3][last] + d[first][c3next] - d[c3][c3next]
					if delta := addRev - removeGain; delta < bestDelta {
						bestDelta, bestAfter, bestReversed = delta, c3, true
					}
				}

				if bestAfter == -1 {
					continue
				}

				tour = applyOrOptMove(tour, start, end, bestAfter, bestReversed)
				improvedAny = true
				pos = make([]int, n)
				for i, c := range tour {
					pos[c] = i
				}
			}
		}
	}

	return tour
}

// applyOrOptMove removes tour[start:end+1] and reinserts it (reversed if
// requested) immediately after city `after`.
func applyOrOptMove(tour []int, start, end, after int, reversed bool) []int {
	chain := append([]int(nil), tour[start:end+1]...)
	if reversed {
		for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
			chain[i], chain[j] = chain[j], chain[i]
		}
	}

	rest := make([]int, 0, len(tour)-len(chain))
	rest = append(rest, tour[:start]...)
	rest = append(rest, tour[end+1:]...)

	result := make([]int, 0, len(tour))
	for _, c := range rest {
		result = append(result, c)
		if c == after {
			result = append(result, chain...)
		}
	}
	return result
}

// rotateToFront rotates a closed tour so that local index `target` sits at
// position 0, without changing the cycle's edges/cost.
func rotateToFront(tour []int, target int) []int {
	idx := 0
	for i, c := range tour {
		if c == target {
			idx = i
			break
		}
	}
	if idx == 0 {
		return tour
	}
	rotated := make([]int, len(tour))
	for i := range tour {
		rotated[i] = tour[(idx+i)%len(tour)]
	}
	return rotated
}

// expandTour converts a local-index tour back into an actual graph walk,
// expanding each consecutive pair into the shortest path inside the region.
func expandTour(g *graph.Graph, region map[graph.VertexID]bool, cells []graph.VertexID, tour []int, returnToStart bool) []graph.VertexID {
	walk := []graph.VertexID{cells[tour[0]]}
	for i := 0; i+1 < len(tour); i++ {
		segment := regionShortestPath(g, region, cells[tour[i]], cells[tour[i+1]])
		walk = append(walk, segment[1:]...)
	}
	if returnToStart {
		root := cells[tour[0]]
		if last := walk[len(walk)-1]; last != root {
			segment := regionShortestPath(g, region, last, root)
			walk = append(walk, segment[1:]...)
		}
	}
	return walk
}

// regionShortestPath returns a shortest path from -> to using only edges
// inside region, via BFS with parent pointers.
func regionShortestPath(g *graph.Graph, region map[graph.VertexID]bool, from, to graph.VertexID) []graph.VertexID {
	if from == to {
		return []graph.VertexID{from}
	}

	parent := map[graph.VertexID]graph.VertexID{from: from}
	queue := []graph.VertexID{from}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == to {
			break
		}
		for _, n := range g.Neighbors(cur) {
			if !region[n] {
				continue
			}
			if _, seen := parent[n]; seen {
				continue
			}
			parent[n] = cur
			queue = append(queue, n)
		}
	}

	path := []graph.VertexID{to}
	for cur := to; cur != from; {
		cur = parent[cur]
		path = append(path, cur)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}
