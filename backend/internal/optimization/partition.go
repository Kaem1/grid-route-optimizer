package optimization

import (
	"fmt"
	"sort"

	"grid-route-optimizer/internal/graph"
	"grid-route-optimizer/internal/routes"
)

// validateProblem checks the preconditions every method shares: a graph, at
// least one start point, and start points that are actually part of it.
func validateProblem(problem Problem) error {
	if problem.Graph == nil {
		return fmt.Errorf("problem has no graph")
	}
	if len(problem.StartPoints) == 0 {
		return fmt.Errorf("problem has no start points")
	}
	for _, start := range problem.StartPoints {
		if !problem.Graph.HasVertex(start) {
			return fmt.Errorf("start point %q is not part of the grid", start)
		}
	}
	return nil
}

// regionsFromOwner turns a cell -> robot-index map into one cell set per
// robot.
func regionsFromOwner(owner map[graph.VertexID]int, k int) []map[graph.VertexID]bool {
	regions := make([]map[graph.VertexID]bool, k)
	for i := range regions {
		regions[i] = make(map[graph.VertexID]bool)
	}
	for v, i := range owner {
		regions[i][v] = true
	}
	return regions
}

// walkAllRegions computes a cheap covering walk (and its cost) for every
// robot's region via WalkRegion (spanning-tree doubling, spec section 6).
// This is the "level 1" surrogate cost used inside search loops (M2/M4);
// final reported routes go through routeAllRegions instead.
func walkAllRegions(g *graph.Graph, starts []graph.VertexID, regions []map[graph.VertexID]bool, returnToStart bool) ([][]graph.VertexID, []int, error) {
	walks := make([][]graph.VertexID, len(starts))
	costs := make([]int, len(starts))
	for i, start := range starts {
		walk, err := WalkRegion(g, regions[i], start, returnToStart)
		if err != nil {
			return nil, nil, fmt.Errorf("robot %d: %w", i, err)
		}
		walks[i] = walk
		costs[i] = len(walk) - 1
	}
	return walks, costs, nil
}

// routeAllRegions computes the final, reported route (and its cost) for
// every robot's region via RouteRegion (nearest-neighbor + 2-opt/Or-opt,
// algorithms_patch.md "routing_tsp.py" level 2).
func routeAllRegions(g *graph.Graph, starts []graph.VertexID, regions []map[graph.VertexID]bool, returnToStart bool) ([][]graph.VertexID, []int, error) {
	walks := make([][]graph.VertexID, len(starts))
	costs := make([]int, len(starts))
	for i, start := range starts {
		walk, err := RouteRegion(g, regions[i], start, returnToStart)
		if err != nil {
			return nil, nil, fmt.Errorf("robot %d: %w", i, err)
		}
		walks[i] = walk
		costs[i] = len(walk) - 1
	}
	return walks, costs, nil
}

// summarize returns (makespan, total) for a list of route costs.
func summarize(costs []int) (makespan, total int) {
	for _, c := range costs {
		total += c
		if c > makespan {
			makespan = c
		}
	}
	return makespan, total
}

// extremes returns (max, min) of a non-empty list of costs.
func extremes(costs []int) (max, min int) {
	max, min = costs[0], costs[0]
	for _, c := range costs[1:] {
		if c > max {
			max = c
		}
		if c < min {
			min = c
		}
	}
	return max, min
}

// coverage returns how many distinct vertices are visited across every walk
// and that count as a ratio of the graph's total vertex count.
func coverage(g *graph.Graph, walks [][]graph.VertexID) (int, float64) {
	covered := make(map[graph.VertexID]bool)
	for _, walk := range walks {
		for _, v := range walk {
			covered[v] = true
		}
	}
	ratio := 0.0
	if total := g.VertexCount(); total > 0 {
		ratio = float64(len(covered)) / float64(total)
	}
	return len(covered), ratio
}

// sortedCells returns a region's vertex IDs in the same deterministic
// (Y,X,ID) order graph.Graph.Neighbors already uses.
func sortedCells(g *graph.Graph, region map[graph.VertexID]bool) []graph.VertexID {
	cells := make([]graph.VertexID, 0, len(region))
	for v := range region {
		cells = append(cells, v)
	}
	sort.Slice(cells, func(i, j int) bool {
		vi, _ := g.Vertex(cells[i])
		vj, _ := g.Vertex(cells[j])
		if vi.Y != vj.Y {
			return vi.Y < vj.Y
		}
		if vi.X != vj.X {
			return vi.X < vj.X
		}
		return cells[i] < cells[j]
	})
	return cells
}

// buildSolution assembles a Solution from per-robot walks/costs, shared by
// every method (spec section 3).
func buildSolution(problem Problem, walks [][]graph.VertexID, costs []int) Solution {
	result := make([]routes.Route, len(walks))
	for i, walk := range walks {
		result[i] = routes.Route{StartPoint: problem.StartPoints[i], Vertices: walk, Length: costs[i]}
	}
	makespan, total := summarize(costs)
	coveredCount, ratio := coverage(problem.Graph, walks)
	lowerBound := LowerBoundGlobal(problem)
	gap := 0.0
	if lowerBound > 0 {
		gap = float64(makespan-lowerBound) / float64(lowerBound)
	}

	return Solution{
		Routes: result,
		Metrics: Metrics{
			TotalLength:     total,
			NumberOfRoutes:  len(result),
			CoveredVertices: coveredCount,
			CoverageRatio:   ratio,
			Makespan:        makespan,
			LowerBound:      lowerBound,
			Gap:             gap,
		},
	}
}
