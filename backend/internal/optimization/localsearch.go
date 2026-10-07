package optimization

import (
	"math"
	"math/rand"

	"grid-route-optimizer/internal/graph"
)

// LocalSearch implements method M4 from the algorithm spec: a simplified,
// author's variant inspired by LS-MCPP (NOT a faithful reproduction) that
// repeatedly moves one boundary cell between two disjoint regions under a
// simulated-annealing schedule, minimizing Σ cost_i². Regions stay
// disjoint (no "transit cells" shared between robots, unlike LS-MCPP) —
// a known, accepted limitation of this variant.
type LocalSearch struct{}

const localSearchTEnd = 0.05

// Name identifies this algorithm for the API/registry.
func (LocalSearch) Name() string { return "local-search" }

// Solve implements Algorithm. It initializes from M2 (weighted_voronoi,
// the spec's default init="m2") and anneals from there using WalkRegion's
// cheap surrogate cost, then routes the final regions with RouteRegion
// (algorithms_patch.md: "koszt użyty w pętli SA zostaje poziomem 1 ...
// metryki końcowe licz zawsze na trasach z route_region").
func (LocalSearch) Solve(problem Problem) (Solution, error) {
	initRegions, initWalks, initCosts, err := weightedVoronoiSearch(problem)
	if err != nil {
		return Solution{}, err
	}

	regions := localSearchAnneal(problem, initRegions, initWalks, initCosts)
	walks, costs, err := routeAllRegions(problem.Graph, problem.StartPoints, regions, problem.ReturnToStart)
	if err != nil {
		return Solution{}, err
	}
	return buildSolution(problem, walks, costs), nil
}

// localSearchAnneal runs the M4 simulated-annealing loop (spec section 7)
// and returns the best regions found, evaluated via the cheap WalkRegion
// surrogate cost throughout.
func localSearchAnneal(problem Problem, initRegions []map[graph.VertexID]bool, initWalks [][]graph.VertexID, initCosts []int) []map[graph.VertexID]bool {
	g := problem.Graph
	starts := problem.StartPoints
	k := len(starts)
	n := g.VertexCount()

	owner := make(map[graph.VertexID]int, n)
	regions := make([]map[graph.VertexID]bool, k)
	for i, region := range initRegions {
		regions[i] = make(map[graph.VertexID]bool, len(region))
		for v := range region {
			regions[i][v] = true
			owner[v] = i
		}
	}

	walks := append([][]graph.VertexID(nil), initWalks...)
	costs := append([]int(nil), initCosts...)

	bestRegions := append([]map[graph.VertexID]bool(nil), regions...)
	bestMakespan, bestTotal := summarize(costs)

	meanCost := 0.0
	for _, c := range costs {
		meanCost += float64(c)
	}
	meanCost /= float64(k)

	t0 := 2 * meanCost
	if t0 <= localSearchTEnd {
		t0 = localSearchTEnd
	}

	rng := rand.New(rand.NewSource(problem.Seed))
	iters := 200 * n

	for it := 0; it < iters; it++ {
		temperature := t0
		if iters > 1 {
			temperature = t0 * math.Pow(localSearchTEnd/t0, float64(it)/float64(iters-1))
		}

		h := selectRegion(rng, costs)
		v, j, ok := selectCandidate(rng, g, regions, owner, starts[h], h)
		if !ok {
			continue
		}

		trimmed := cloneRegionWithout(regions[h], v)
		if _, err := BuildTree(g, trimmed, starts[h], TreeBFS); err != nil {
			continue // removing v would disconnect region h
		}
		grown := cloneRegionWith(regions[j], v)

		newWalkH, err := WalkRegion(g, trimmed, starts[h], problem.ReturnToStart)
		if err != nil {
			continue
		}
		newWalkJ, err := WalkRegion(g, grown, starts[j], problem.ReturnToStart)
		if err != nil {
			continue
		}
		newCostH, newCostJ := len(newWalkH)-1, len(newWalkJ)-1

		delta := float64(newCostH*newCostH+newCostJ*newCostJ) - float64(costs[h]*costs[h]+costs[j]*costs[j])
		accept := delta <= 0
		if !accept {
			accept = rng.Float64() < math.Exp(-delta/temperature)
		}
		if !accept {
			continue
		}

		regions[h], regions[j] = trimmed, grown
		owner[v] = j
		walks[h], walks[j] = newWalkH, newWalkJ
		costs[h], costs[j] = newCostH, newCostJ

		makespan, total := summarize(costs)
		if makespan < bestMakespan || (makespan == bestMakespan && total < bestTotal) {
			bestMakespan, bestTotal = makespan, total
			bestRegions = append([]map[graph.VertexID]bool(nil), regions...)
		}
	}

	return bestRegions
}

// selectRegion picks the region index to modify: 50% of the time the
// current max-cost region (ties broken randomly), otherwise a uniformly
// random region (spec section 7, M4 step 1).
func selectRegion(rng *rand.Rand, costs []int) int {
	if rng.Float64() < 0.5 {
		maxCost := costs[0]
		for _, c := range costs[1:] {
			if c > maxCost {
				maxCost = c
			}
		}
		var tied []int
		for i, c := range costs {
			if c == maxCost {
				tied = append(tied, i)
			}
		}
		return tied[rng.Intn(len(tied))]
	}
	return rng.Intn(len(costs))
}

// selectCandidate gathers every boundary cell of region h (excluding its
// start) that touches a different region j, in deterministic order, then
// picks one uniformly at random (spec section 7, M4 steps 2-3).
func selectCandidate(rng *rand.Rand, g *graph.Graph, regions []map[graph.VertexID]bool, owner map[graph.VertexID]int, start graph.VertexID, h int) (graph.VertexID, int, bool) {
	type candidate struct {
		v graph.VertexID
		j int
	}

	var candidates []candidate
	seen := make(map[candidate]bool)
	for _, v := range sortedCells(g, regions[h]) {
		if v == start {
			continue
		}
		for _, nb := range g.Neighbors(v) {
			j := owner[nb]
			if j == h {
				continue
			}
			c := candidate{v: v, j: j}
			if seen[c] {
				continue
			}
			seen[c] = true
			candidates = append(candidates, c)
		}
	}

	if len(candidates) == 0 {
		return "", 0, false
	}
	pick := candidates[rng.Intn(len(candidates))]
	return pick.v, pick.j, true
}

func cloneRegionWithout(region map[graph.VertexID]bool, remove graph.VertexID) map[graph.VertexID]bool {
	clone := make(map[graph.VertexID]bool, len(region))
	for v := range region {
		if v != remove {
			clone[v] = true
		}
	}
	return clone
}

func cloneRegionWith(region map[graph.VertexID]bool, add graph.VertexID) map[graph.VertexID]bool {
	clone := make(map[graph.VertexID]bool, len(region)+1)
	for v := range region {
		clone[v] = true
	}
	clone[add] = true
	return clone
}
