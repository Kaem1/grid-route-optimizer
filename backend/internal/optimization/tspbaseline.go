package optimization

import (
	"fmt"

	"grid-route-optimizer/internal/graph"
)

// TSPBaseline implements method M7 from algorithms_patch.md: a single
// robot covering the entire board via RouteRegion. It exists as a
// reference point ("how much would one robot alone cost") and as an
// isolated test bed for the 2-opt/Or-opt routing logic, independent of any
// region-partitioning method.
type TSPBaseline struct{}

// Name identifies this algorithm for the API/registry.
func (TSPBaseline) Name() string { return "tsp-baseline" }

// Solve implements Algorithm. Requires exactly one start point.
func (TSPBaseline) Solve(problem Problem) (Solution, error) {
	if err := validateProblem(problem); err != nil {
		return Solution{}, err
	}
	if len(problem.StartPoints) != 1 {
		return Solution{}, fmt.Errorf("tsp-baseline requires exactly one start point, got %d", len(problem.StartPoints))
	}

	region := make(map[graph.VertexID]bool, problem.Graph.VertexCount())
	for _, v := range problem.Graph.Vertices() {
		region[v.ID] = true
	}

	walk, err := RouteRegion(problem.Graph, region, problem.StartPoints[0], problem.ReturnToStart)
	if err != nil {
		return Solution{}, err
	}

	return buildSolution(problem, [][]graph.VertexID{walk}, []int{len(walk) - 1}), nil
}
