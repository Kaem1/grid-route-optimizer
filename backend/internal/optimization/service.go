package optimization

import (
	"fmt"

	"grid-route-optimizer/internal/graph"
	"grid-route-optimizer/internal/grid"
)

// Service is the entry point used by the API layer: it turns a validated
// grid definition, a set of start points and an algorithm name into a
// Solution, without knowing anything about HTTP or JSON (see project
// instructions, section 45 — "HTTP Handler -> Request DTO -> Service ->
// Domain Model -> Algorithm -> Solution -> Response DTO").
type Service struct {
	registry *Registry
}

// NewService builds a Service backed by the given algorithm registry.
func NewService(registry *Registry) *Service {
	return &Service{registry: registry}
}

// AlgorithmNames lists the algorithms available for simulation requests.
func (s *Service) AlgorithmNames() []string {
	return s.registry.Names()
}

// Simulate validates the grid/start points, builds the domain graph, and
// runs the requested algorithm against it. The backend never trusts data
// coming from the frontend at face value (section 47).
func (s *Service) Simulate(spec grid.GridSpec, startPointIDs []string, algorithmName string) (Solution, error) {
	algorithm, ok := s.registry.Get(algorithmName)
	if !ok {
		return Solution{}, fmt.Errorf("unknown algorithm %q", algorithmName)
	}

	g, err := grid.Build(spec)
	if err != nil {
		return Solution{}, fmt.Errorf("invalid grid: %w", err)
	}

	if len(startPointIDs) == 0 {
		return Solution{}, fmt.Errorf("at least one start point is required")
	}

	startPoints := make([]graph.VertexID, 0, len(startPointIDs))
	for _, id := range startPointIDs {
		vid := graph.VertexID(id)
		if !g.HasVertex(vid) {
			return Solution{}, fmt.Errorf("start point %q is not an active cell", id)
		}
		startPoints = append(startPoints, vid)
	}

	return algorithm.Solve(Problem{Graph: g, StartPoints: startPoints})
}
