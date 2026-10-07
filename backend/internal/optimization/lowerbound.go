package optimization

import "grid-route-optimizer/internal/graph"

// LowerBoundGlobal is the section-5 (patched) lower bound on makespan,
// independent of any particular partition: the max of a counting bound, a
// graph-distance bound, and (for return_to_start) a checkerboard-parity
// bound (algorithms_patch.md section "Zmiana sekcji 5").
func LowerBoundGlobal(problem Problem) int {
	g := problem.Graph
	k := len(problem.StartPoints)
	if g == nil || k == 0 {
		return 0
	}
	n := g.VertexCount()

	_, dist := graph.MultiSourceBFS(g, problem.StartPoints)

	bound := countingLowerBound(n, k, problem.ReturnToStart)
	if db := distanceLowerBound(dist, problem.ReturnToStart); db > bound {
		bound = db
	}
	if pb := globalParityLowerBound(g, k, problem.ReturnToStart); pb > bound {
		bound = pb
	}
	return bound
}

// countingLowerBound: with return, ceil(N/k) rounded up to even (closed
// walks on a bipartite grid have even length); without return,
// max(ceil(N/k)-1, 0).
func countingLowerBound(n, k int, returnToStart bool) int {
	base := ceilDiv(n, k)
	if returnToStart {
		if base%2 != 0 {
			base++
		}
		return base
	}
	if base == 0 {
		return 0
	}
	return base - 1
}

// distanceLowerBound: D = max_v min_i d(s_i,v); without return D, with
// return 2D.
func distanceLowerBound(dist map[graph.VertexID]int, returnToStart bool) int {
	maxDist := 0
	for _, d := range dist {
		if d > maxDist {
			maxDist = d
		}
	}
	if returnToStart {
		return 2 * maxDist
	}
	return maxDist
}

// globalParityLowerBound: for return_to_start only, makespan >=
// ceil((N + Δ_total) / k), derived by chaining all k closed walks into one
// even-length multi-walk covering the whole board (patch section 5). Not
// derived for the open-walk case, so it's skipped there.
func globalParityLowerBound(g *graph.Graph, k int, returnToStart bool) int {
	if !returnToStart {
		return 0
	}
	black, white := 0, 0
	for _, v := range g.Vertices() {
		if cellColor(v) == 0 {
			black++
		} else {
			white++
		}
	}
	n := black + white
	delta := black - white
	if delta < 0 {
		delta = -delta
	}
	return ceilDiv(n+delta, k)
}

func cellColor(v graph.Vertex) int {
	c := (v.X + v.Y) % 2
	if c < 0 {
		c += 2
	}
	return c
}

func ceilDiv(a, b int) int {
	if b <= 0 {
		return 0
	}
	return (a + b - 1) / b
}
