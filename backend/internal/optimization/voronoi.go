package optimization

import (
	"fmt"

	"grid-route-optimizer/internal/graph"
)

// VoronoiTree implements method M1 from the algorithm spec: partition the
// graph into one region per start point via multi-source BFS (nearest
// start wins ties by discovery order, section 4), then cover each region
// with a 2-opt/Or-opt route (algorithms_patch.md section 6). It does no
// balancing between regions, so costs between robots can be uneven —
// method M2 (weighted_voronoi) refines this by shifting boundary cells to
// even out region costs.
type VoronoiTree struct{}

// Name identifies this algorithm for the API/registry.
func (VoronoiTree) Name() string { return "voronoi-tree" }

// Solve implements Algorithm.
func (VoronoiTree) Solve(problem Problem) (Solution, error) {
	if err := validateProblem(problem); err != nil {
		return Solution{}, err
	}

	owner, _ := graph.MultiSourceBFS(problem.Graph, problem.StartPoints)
	if len(owner) != problem.Graph.VertexCount() {
		return Solution{}, fmt.Errorf("grid contains a cell unreachable from every start point")
	}

	regions := regionsFromOwner(owner, len(problem.StartPoints))
	walks, costs, err := routeAllRegions(problem.Graph, problem.StartPoints, regions, problem.ReturnToStart)
	if err != nil {
		return Solution{}, err
	}

	return buildSolution(problem, walks, costs), nil
}
