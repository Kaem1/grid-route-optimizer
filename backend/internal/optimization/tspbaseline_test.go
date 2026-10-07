package optimization_test

import (
	"reflect"
	"testing"

	"grid-route-optimizer/internal/graph"
	"grid-route-optimizer/internal/optimization"
)

func TestTSPBaselineSingleVertex(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})

	solution, err := optimization.TSPBaseline{}.Solve(optimization.Problem{
		Graph:       g,
		StartPoints: []graph.VertexID{"0,0"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(solution.Routes) != 1 || len(solution.Routes[0].Vertices) != 1 {
		t.Fatalf("expected a single-vertex walk, got %+v", solution.Routes)
	}
}

func TestTSPBaselineRejectsWrongStartCount(t *testing.T) {
	g := buildGrid(3, 3)
	baseline := optimization.TSPBaseline{}

	if _, err := baseline.Solve(optimization.Problem{Graph: g}); err == nil {
		t.Fatal("expected an error for zero start points")
	}
	if _, err := baseline.Solve(optimization.Problem{
		Graph:       g,
		StartPoints: []graph.VertexID{"0,0", "1,0"},
	}); err == nil {
		t.Fatal("expected an error for more than one start point")
	}
}

func TestTSPBaselineCoversWholeBoard(t *testing.T) {
	for _, returnToStart := range []bool{false, true} {
		g := buildGrid(6, 5)
		solution, err := optimization.TSPBaseline{}.Solve(optimization.Problem{
			Graph:         g,
			StartPoints:   []graph.VertexID{"0,0"},
			ReturnToStart: returnToStart,
		})
		if err != nil {
			t.Fatalf("returnToStart=%v: unexpected error: %v", returnToStart, err)
		}
		if solution.Metrics.CoverageRatio != 1.0 {
			t.Fatalf("returnToStart=%v: expected full coverage, got %v", returnToStart, solution.Metrics.CoverageRatio)
		}
		validateWalk(t, g, "0,0", solution.Routes[0].Vertices, returnToStart)
		if solution.Metrics.Makespan < solution.Metrics.LowerBound {
			t.Fatalf("returnToStart=%v: makespan %d below lower bound %d", returnToStart, solution.Metrics.Makespan, solution.Metrics.LowerBound)
		}
	}
}

func TestTSPBaselineIsDeterministic(t *testing.T) {
	problem := optimization.Problem{Graph: buildGrid(5, 5), StartPoints: []graph.VertexID{"0,0"}}

	first, err := optimization.TSPBaseline{}.Solve(problem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := optimization.TSPBaseline{}.Solve(problem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(first.Routes, second.Routes) {
		t.Fatalf("expected identical routes across runs, got %v vs %v", first.Routes, second.Routes)
	}
}
