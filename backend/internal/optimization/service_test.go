package optimization_test

import (
	"testing"

	"grid-route-optimizer/internal/grid"
	"grid-route-optimizer/internal/optimization"
)

func newService() *optimization.Service {
	registry := optimization.NewRegistry(optimization.DFSCoverage{})
	return optimization.NewService(registry)
}

func TestServiceSimulateHappyPath(t *testing.T) {
	spec := grid.GridSpec{
		Cols:           2,
		Rows:           1,
		ActiveVertices: []string{"0,0", "1,0"},
	}

	solution, err := newService().Simulate(spec, []string{"0,0"}, "dfs-coverage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(solution.Routes) != 1 {
		t.Fatalf("expected one route, got %d", len(solution.Routes))
	}
}

func TestServiceSimulateRejectsUnknownAlgorithm(t *testing.T) {
	spec := grid.GridSpec{Cols: 1, Rows: 1, ActiveVertices: []string{"0,0"}}
	_, err := newService().Simulate(spec, []string{"0,0"}, "does-not-exist")
	if err == nil {
		t.Fatal("expected an error for an unknown algorithm")
	}
}

func TestServiceSimulateRejectsStartPointOutsideBoard(t *testing.T) {
	spec := grid.GridSpec{Cols: 2, Rows: 1, ActiveVertices: []string{"0,0"}}
	_, err := newService().Simulate(spec, []string{"1,0"}, "dfs-coverage")
	if err == nil {
		t.Fatal("expected an error for a start point that is not an active cell")
	}
}

func TestServiceSimulateRejectsEmptyStartPoints(t *testing.T) {
	spec := grid.GridSpec{Cols: 1, Rows: 1, ActiveVertices: []string{"0,0"}}
	_, err := newService().Simulate(spec, nil, "dfs-coverage")
	if err == nil {
		t.Fatal("expected an error when no start points are provided")
	}
}
