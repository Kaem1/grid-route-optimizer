package optimization

import (
	"container/heap"
	"fmt"
	"math"

	"grid-route-optimizer/internal/graph"
)

// WeightedVoronoi implements method M2 from the algorithm spec: an
// author's-variant balancing scheme (inspired by DARP's region-balancing
// idea, NOT a faithful DARP reproduction). It repeatedly re-partitions the
// graph via a weighted multi-source Dijkstra, nudging each start's weight
// up whenever its region is more expensive than average so it loses
// boundary cells to its neighbors, until region costs are balanced or
// progress stalls.
type WeightedVoronoi struct{}

const (
	weightedVoronoiMaxIter  = 200
	weightedVoronoiEta0     = 0.5
	weightedVoronoiDecay    = 0.98
	weightedVoronoiPatience = 30
)

// Name identifies this algorithm for the API/registry.
func (WeightedVoronoi) Name() string { return "weighted-voronoi" }

// Solve implements Algorithm.
func (WeightedVoronoi) Solve(problem Problem) (Solution, error) {
	regions, _, _, err := weightedVoronoiSearch(problem)
	if err != nil {
		return Solution{}, err
	}
	walks, costs, err := routeAllRegions(problem.Graph, problem.StartPoints, regions, problem.ReturnToStart)
	if err != nil {
		return Solution{}, err
	}
	return buildSolution(problem, walks, costs), nil
}

// weightedVoronoiSearch runs the M2 balancing loop (spec section 7) using
// WalkRegion's cheap surrogate cost, and returns the best regions found
// across all iterations (plus their surrogate walks/costs, which M4 reuses
// directly as its initial state). Callers that need the final, reported
// route must pass the returned regions through routeAllRegions themselves
// (algorithms_patch.md: "M1 i M2 ... ich krok 'trasa' na końcu ... ma
// wywołać route_region zamiast walk_region", while the balancing loop
// itself keeps using the cheap level-1 surrogate).
func weightedVoronoiSearch(problem Problem) ([]map[graph.VertexID]bool, [][]graph.VertexID, []int, error) {
	if err := validateProblem(problem); err != nil {
		return nil, nil, nil, err
	}

	k := len(problem.StartPoints)
	w := make([]float64, k)
	eta := weightedVoronoiEta0

	var bestRegions []map[graph.VertexID]bool
	var bestWalks [][]graph.VertexID
	var bestCosts []int
	bestMakespan, bestTotal := math.MaxInt64, math.MaxInt64
	noImprovement := 0

	for it := 0; it < weightedVoronoiMaxIter; it++ {
		owner := weightedVoronoiOwner(problem.Graph, problem.StartPoints, w)
		if len(owner) != problem.Graph.VertexCount() {
			return nil, nil, nil, fmt.Errorf("grid contains a cell unreachable from every start point")
		}

		regions := regionsFromOwner(owner, k)
		walks, costs, err := walkAllRegions(problem.Graph, problem.StartPoints, regions, problem.ReturnToStart)
		if err != nil {
			return nil, nil, nil, err
		}

		makespan, total := summarize(costs)
		if makespan < bestMakespan || (makespan == bestMakespan && total < bestTotal) {
			bestMakespan, bestTotal = makespan, total
			bestRegions, bestWalks, bestCosts = regions, walks, costs
			noImprovement = 0
		} else {
			noImprovement++
		}

		threshold := 1
		if problem.ReturnToStart {
			threshold = 2
		}
		maxCost, minCost := extremes(costs)
		if maxCost-minCost <= threshold || noImprovement >= weightedVoronoiPatience {
			break
		}

		mean := 0.0
		for _, c := range costs {
			mean += float64(c)
		}
		mean /= float64(k)
		for i, c := range costs {
			w[i] += eta * (float64(c) - mean)
		}
		eta *= weightedVoronoiDecay
	}

	return bestRegions, bestWalks, bestCosts, nil
}

// weightedVoronoiOwner runs one weighted multi-source Dijkstra partition
// step (spec section 7, M2 2a): every start point is locked to its own
// robot forever (a rival's expansion can never claim it), and equal-cost
// ties among the remaining cells are broken by the smaller robot index via
// the priority queue's (dist, robot, cell) ordering.
func weightedVoronoiOwner(g *graph.Graph, starts []graph.VertexID, w []float64) map[graph.VertexID]int {
	owner := make(map[graph.VertexID]int, len(starts))
	locked := make(map[graph.VertexID]int, len(starts))
	for i, s := range starts {
		locked[s] = i
	}

	pq := make(voronoiQueue, 0, len(starts))
	for i, s := range starts {
		owner[s] = i
		pq = append(pq, voronoiItem{dist: w[i], robot: i, cell: s})
	}
	heap.Init(&pq)

	settled := make(map[graph.VertexID]bool, len(starts))
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(voronoiItem)
		if settled[item.cell] {
			continue
		}
		settled[item.cell] = true
		owner[item.cell] = item.robot

		for _, n := range g.Neighbors(item.cell) {
			if settled[n] {
				continue
			}
			if lockedOwner, isLocked := locked[n]; isLocked && lockedOwner != item.robot {
				continue
			}
			heap.Push(&pq, voronoiItem{dist: item.dist + 1, robot: item.robot, cell: n})
		}
	}

	return owner
}

// voronoiItem is one weightedVoronoiOwner priority-queue entry; its
// ordering key is (dist, robot, cell) per spec section 7, so equal-distance
// ties always favor the smaller robot index deterministically.
type voronoiItem struct {
	dist  float64
	robot int
	cell  graph.VertexID
}

type voronoiQueue []voronoiItem

func (q voronoiQueue) Len() int { return len(q) }

func (q voronoiQueue) Less(i, j int) bool {
	if q[i].dist != q[j].dist {
		return q[i].dist < q[j].dist
	}
	if q[i].robot != q[j].robot {
		return q[i].robot < q[j].robot
	}
	return q[i].cell < q[j].cell
}

func (q voronoiQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }

func (q *voronoiQueue) Push(x any) { *q = append(*q, x.(voronoiItem)) }

func (q *voronoiQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[:n-1]
	return item
}
