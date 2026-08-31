package optimization

import (
	"fmt"

	"grid-route-optimizer/internal/graph"
	"grid-route-optimizer/internal/routes"
)

// DFSCoverage is the simplest possible covering algorithm: from each start
// point it walks a depth-first spanning tree of the reachable component,
// stepping back to the parent vertex whenever a branch is exhausted. The
// resulting walk is a single continuous path that visits every vertex
// reachable from the start point at least once, which satisfies the
// covering goal from project instructions section 35 without attempting
// any actual optimization — later algorithms can replace this one without
// changing the Algorithm interface.
type DFSCoverage struct{}

// Name identifies this algorithm for the API/registry.
func (DFSCoverage) Name() string { return "dfs-coverage" }

// Solve implements Algorithm.
func (DFSCoverage) Solve(problem Problem) (Solution, error) {
	if problem.Graph == nil {
		return Solution{}, fmt.Errorf("problem has no graph")
	}
	if len(problem.StartPoints) == 0 {
		return Solution{}, fmt.Errorf("problem has no start points")
	}

	covered := make(map[graph.VertexID]bool)
	result := make([]routes.Route, 0, len(problem.StartPoints))

	for _, start := range problem.StartPoints {
		if !problem.Graph.HasVertex(start) {
			return Solution{}, fmt.Errorf("start point %q is not part of the grid", start)
		}

		walk := dfsWalk(problem.Graph, start)
		for _, v := range walk {
			covered[v] = true
		}

		result = append(result, routes.Route{
			StartPoint: start,
			Vertices:   walk,
			Length:     len(walk) - 1,
		})
	}

	totalLength := 0
	for _, route := range result {
		totalLength += route.Length
	}

	ratio := 0.0
	if total := problem.Graph.VertexCount(); total > 0 {
		ratio = float64(len(covered)) / float64(total)
	}

	return Solution{
		Routes: result,
		Metrics: Metrics{
			TotalLength:     totalLength,
			NumberOfRoutes:  len(result),
			CoveredVertices: len(covered),
			CoverageRatio:   ratio,
		},
	}, nil
}

// dfsWalk performs an iterative depth-first traversal of the connected
// component containing start, recording every step — including backtracking
// steps — so the result is one continuous walk rather than just a visited
// set.
func dfsWalk(g *graph.Graph, start graph.VertexID) []graph.VertexID {
	visited := map[graph.VertexID]bool{start: true}
	walk := []graph.VertexID{start}
	stack := []graph.VertexID{start}

	for len(stack) > 0 {
		current := stack[len(stack)-1]

		next := firstUnvisitedNeighbor(g, current, visited)
		if next == "" {
			stack = stack[:len(stack)-1]
			if len(stack) > 0 {
				walk = append(walk, stack[len(stack)-1])
			}
			continue
		}

		visited[next] = true
		walk = append(walk, next)
		stack = append(stack, next)
	}

	return walk
}

func firstUnvisitedNeighbor(g *graph.Graph, id graph.VertexID, visited map[graph.VertexID]bool) graph.VertexID {
	for _, n := range g.Neighbors(id) {
		if !visited[n] {
			return n
		}
	}
	return ""
}
