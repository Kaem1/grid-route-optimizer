package optimization_test

import (
	"reflect"
	"testing"

	"grid-route-optimizer/internal/graph"
	"grid-route-optimizer/internal/optimization"
)

func TestLocalSearchSingleVertex(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})

	solution, err := optimization.LocalSearch{}.Solve(optimization.Problem{
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

func TestLocalSearchRejectsUnreachableCell(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})
	g.AddVertex(graph.Vertex{ID: "5,5", X: 5, Y: 5})

	_, err := optimization.LocalSearch{}.Solve(optimization.Problem{
		Graph:       g,
		StartPoints: []graph.VertexID{"0,0"},
	})
	if err == nil {
		t.Fatal("expected an error for a cell unreachable from every start")
	}
}

func TestLocalSearchSingleRobotIsNoOp(t *testing.T) {
	// With one robot there are never two regions to trade cells between,
	// so the search must leave the (only possible) full-coverage state
	// untouched.
	g := buildGrid(4, 4)
	solution, err := optimization.LocalSearch{}.Solve(optimization.Problem{
		Graph:       g,
		StartPoints: []graph.VertexID{"0,0"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if solution.Metrics.CoverageRatio != 1.0 {
		t.Fatalf("expected full coverage, got %v", solution.Metrics.CoverageRatio)
	}
}

func TestLocalSearchCoversMultiRobotGrid(t *testing.T) {
	for _, returnToStart := range []bool{false, true} {
		g := buildGrid(6, 5)
		starts := []graph.VertexID{"0,0", "5,4", "5,0"}

		solution, err := optimization.LocalSearch{}.Solve(optimization.Problem{
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

func TestLocalSearchIsDeterministic(t *testing.T) {
	problem := optimization.Problem{
		Graph:       buildGrid(5, 5),
		StartPoints: []graph.VertexID{"0,0", "4,4"},
		Seed:        42,
	}

	first, err := optimization.LocalSearch{}.Solve(problem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := optimization.LocalSearch{}.Solve(problem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(first.Routes, second.Routes) {
		t.Fatalf("expected identical routes across runs, got %v vs %v", first.Routes, second.Routes)
	}
}

// TestLocalSearchNotWorseThanInit is the acceptance criterion from spec
// section 11: M4 must not be worse (higher makespan, or equal makespan but
// higher total) than its own initialization (M2).
func TestLocalSearchNotWorseThanInit(t *testing.T) {
	g := buildGrid(8, 6)
	starts := []graph.VertexID{"0,0", "7,5", "7,0", "0,5"}
	problem := optimization.Problem{Graph: g, StartPoints: starts}

	init, err := optimization.WeightedVoronoi{}.Solve(problem)
	if err != nil {
		t.Fatalf("M2 unexpected error: %v", err)
	}
	result, err := optimization.LocalSearch{}.Solve(problem)
	if err != nil {
		t.Fatalf("M4 unexpected error: %v", err)
	}

	if result.Metrics.Makespan > init.Metrics.Makespan {
		t.Fatalf("expected M4 makespan (%d) <= M2 makespan (%d)", result.Metrics.Makespan, init.Metrics.Makespan)
	}
	if result.Metrics.Makespan == init.Metrics.Makespan && result.Metrics.TotalLength > init.Metrics.TotalLength {
		t.Fatalf("expected M4 total (%d) <= M2 total (%d) at equal makespan", result.Metrics.TotalLength, init.Metrics.TotalLength)
	}
}
