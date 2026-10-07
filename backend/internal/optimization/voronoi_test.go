package optimization_test

import (
	"reflect"
	"testing"

	"grid-route-optimizer/internal/graph"
	"grid-route-optimizer/internal/grid"
	"grid-route-optimizer/internal/optimization"
)

func TestVoronoiTreeSingleVertex(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})

	solution, err := optimization.VoronoiTree{}.Solve(optimization.Problem{
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

func TestVoronoiTreeRejectsUnknownStartPoint(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})

	_, err := optimization.VoronoiTree{}.Solve(optimization.Problem{
		Graph:       g,
		StartPoints: []graph.VertexID{"9,9"},
	})
	if err == nil {
		t.Fatal("expected an error for a start point outside the grid")
	}
}

func TestVoronoiTreeRejectsNoStartPoints(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})
	_, err := optimization.VoronoiTree{}.Solve(optimization.Problem{Graph: g})
	if err == nil {
		t.Fatal("expected an error when no start points are given")
	}
}

func TestVoronoiTreeRejectsNilGraph(t *testing.T) {
	_, err := optimization.VoronoiTree{}.Solve(optimization.Problem{
		StartPoints: []graph.VertexID{"0,0"},
	})
	if err == nil {
		t.Fatal("expected an error for a nil graph")
	}
}

func TestVoronoiTreeRejectsUnreachableCell(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})
	g.AddVertex(graph.Vertex{ID: "5,5", X: 5, Y: 5}) // disconnected, no start nearby

	_, err := optimization.VoronoiTree{}.Solve(optimization.Problem{
		Graph:       g,
		StartPoints: []graph.VertexID{"0,0"},
	})
	if err == nil {
		t.Fatal("expected an error for a cell unreachable from every start")
	}
}

// buildGrid constructs a cols x rows board (no walls) via internal/grid, so
// this test exercises the exact code path the frontend/API layer uses.
func buildGrid(cols, rows int) *graph.Graph {
	active := make([]string, 0, cols*rows)
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			active = append(active, grid.VertexID(x, y))
		}
	}
	g, err := grid.Build(grid.GridSpec{Cols: cols, Rows: rows, ActiveVertices: active})
	if err != nil {
		panic(err)
	}
	return g
}

func validateWalk(t *testing.T, g *graph.Graph, start graph.VertexID, walk []graph.VertexID, returnToStart bool) {
	t.Helper()
	if len(walk) == 0 || walk[0] != start {
		t.Fatalf("walk must start at %q, got %v", start, walk)
	}
	if returnToStart && walk[len(walk)-1] != start {
		t.Fatalf("returnToStart walk must end at %q, got %v", start, walk)
	}
	for i := 1; i < len(walk); i++ {
		if walk[i] == walk[i-1] {
			t.Fatalf("walk has a stationary step at index %d: %v", i, walk)
		}
		if !g.HasEdge(walk[i-1], walk[i]) {
			t.Fatalf("walk has a non-adjacent step %q -> %q", walk[i-1], walk[i])
		}
	}
}

func TestVoronoiTreeCoversMultiRobotGrid(t *testing.T) {
	for _, returnToStart := range []bool{false, true} {
		g := buildGrid(6, 5)
		starts := []graph.VertexID{"0,0", "5,4", "5,0"}

		solution, err := optimization.VoronoiTree{}.Solve(optimization.Problem{
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
		if len(solution.Routes) != len(starts) {
			t.Fatalf("returnToStart=%v: expected %d routes, got %d", returnToStart, len(starts), len(solution.Routes))
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

func TestVoronoiTreeIsDeterministic(t *testing.T) {
	problem := optimization.Problem{
		Graph:       buildGrid(5, 5),
		StartPoints: []graph.VertexID{"0,0", "4,4"},
	}

	first, err := optimization.VoronoiTree{}.Solve(problem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := optimization.VoronoiTree{}.Solve(problem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(first.Routes, second.Routes) {
		t.Fatalf("expected identical routes across runs, got %v vs %v", first.Routes, second.Routes)
	}
}
