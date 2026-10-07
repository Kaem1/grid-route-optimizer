package optimization_test

import (
	"reflect"
	"testing"

	"grid-route-optimizer/internal/graph"
	"grid-route-optimizer/internal/optimization"
)

func TestWeightedVoronoiSingleVertex(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})

	solution, err := optimization.WeightedVoronoi{}.Solve(optimization.Problem{
		Graph:       g,
		StartPoints: []graph.VertexID{"0,0"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(solution.Routes) != 1 || len(solution.Routes[0].Vertices) != 1 {
		t.Fatalf("expected a single-vertex walk, got %+v", solution.Routes)
	}
	if solution.Metrics.CoverageRatio != 1.0 {
		t.Fatalf("expected full coverage, got %v", solution.Metrics.CoverageRatio)
	}
}

func TestWeightedVoronoiRejectsUnknownStartPoint(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})

	_, err := optimization.WeightedVoronoi{}.Solve(optimization.Problem{
		Graph:       g,
		StartPoints: []graph.VertexID{"9,9"},
	})
	if err == nil {
		t.Fatal("expected an error for a start point outside the grid")
	}
}

func TestWeightedVoronoiRejectsUnreachableCell(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})
	g.AddVertex(graph.Vertex{ID: "5,5", X: 5, Y: 5})

	_, err := optimization.WeightedVoronoi{}.Solve(optimization.Problem{
		Graph:       g,
		StartPoints: []graph.VertexID{"0,0"},
	})
	if err == nil {
		t.Fatal("expected an error for a cell unreachable from every start")
	}
}

func TestWeightedVoronoiCoversMultiRobotGrid(t *testing.T) {
	for _, returnToStart := range []bool{false, true} {
		g := buildGrid(6, 5)
		starts := []graph.VertexID{"0,0", "5,4", "5,0"}

		solution, err := optimization.WeightedVoronoi{}.Solve(optimization.Problem{
			Graph:         g,
			StartPoints:   starts,
			ReturnToStart: returnToStart,
		})
		if err != nil {
			t.Fatalf("returnToStart=%v: unexpected error: %v", returnToStart, err)
		}
		if solution.Metrics.CoverageRatio != 1.0 {
			t.Fatalf("returnToStart=%v: expected full coverage, got %v", returnToStart, solution.Metrics.CoverageRatio)
		}

		covered := map[graph.VertexID]bool{}
		maxLen := 0
		for i, route := range solution.Routes {
			validateWalk(t, g, starts[i], route.Vertices, returnToStart)
			for _, v := range route.Vertices {
				covered[v] = true
			}
			if route.Length > maxLen {
				maxLen = route.Length
			}
		}
		if len(covered) != g.VertexCount() {
			t.Fatalf("returnToStart=%v: expected every cell covered, got %d/%d", returnToStart, len(covered), g.VertexCount())
		}
		if solution.Metrics.Makespan != maxLen {
			t.Fatalf("returnToStart=%v: expected makespan %d, got %d", returnToStart, maxLen, solution.Metrics.Makespan)
		}
		if solution.Metrics.Makespan < solution.Metrics.LowerBound {
			t.Fatalf("returnToStart=%v: makespan %d below lower bound %d", returnToStart, solution.Metrics.Makespan, solution.Metrics.LowerBound)
		}
	}
}

func TestWeightedVoronoiIsDeterministic(t *testing.T) {
	problem := optimization.Problem{
		Graph:       buildGrid(5, 5),
		StartPoints: []graph.VertexID{"0,0", "4,4"},
	}

	first, err := optimization.WeightedVoronoi{}.Solve(problem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := optimization.WeightedVoronoi{}.Solve(problem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(first.Routes, second.Routes) {
		t.Fatalf("expected identical routes across runs, got %v vs %v", first.Routes, second.Routes)
	}
}

// TestWeightedVoronoiBalancesBetterThanPlainVoronoi checks the whole point
// of M2: on a grid where M1's plain nearest-start partition is lopsided
// (one start crammed in a corner near another), M2's balancing should not
// produce a worse (makespan, total) than M1's.
func TestWeightedVoronoiBalancesBetterThanPlainVoronoi(t *testing.T) {
	g := buildGrid(10, 4)
	starts := []graph.VertexID{"0,0", "1,0"}
	problem := optimization.Problem{Graph: g, StartPoints: starts}

	m1, err := optimization.VoronoiTree{}.Solve(problem)
	if err != nil {
		t.Fatalf("M1 unexpected error: %v", err)
	}
	m2, err := optimization.WeightedVoronoi{}.Solve(problem)
	if err != nil {
		t.Fatalf("M2 unexpected error: %v", err)
	}

	if m2.Metrics.Makespan > m1.Metrics.Makespan {
		t.Fatalf("expected M2 makespan (%d) <= M1 makespan (%d)", m2.Metrics.Makespan, m1.Metrics.Makespan)
	}
}
