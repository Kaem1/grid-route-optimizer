package optimization_test

import (
	"testing"

	"grid-route-optimizer/internal/graph"
	"grid-route-optimizer/internal/grid"
	"grid-route-optimizer/internal/optimization"
)

func TestLowerBoundGlobalRectangleClosed(t *testing.T) {
	g := buildGrid(4, 4)
	problem := optimization.Problem{
		Graph:         g,
		StartPoints:   []graph.VertexID{"0,0"},
		ReturnToStart: true,
	}
	// 4x4 is perfectly balanced (Δ=0), so the parity bound N+Δ=16 should
	// dominate the counting (ceil(16/1)=16) and distance (2*6=12) bounds.
	if lb := optimization.LowerBoundGlobal(problem); lb != 16 {
		t.Fatalf("expected lower bound 16, got %d", lb)
	}
}

func TestLowerBoundGlobalRectangleOpen(t *testing.T) {
	g := buildGrid(4, 4)
	problem := optimization.Problem{
		Graph:       g,
		StartPoints: []graph.VertexID{"0,0"},
	}
	// counting: ceil(16/1)-1=15; distance: max dist to (3,3) = 6; global
	// parity isn't derived for the open case, so 15 should win.
	if lb := optimization.LowerBoundGlobal(problem); lb != 15 {
		t.Fatalf("expected lower bound 15, got %d", lb)
	}
}

// TestLowerBoundGlobalParityCorner mirrors the patch's regression test: a
// board with a parity imbalance of exactly 2 (two same-color cells cut from
// an otherwise balanced rectangle) must push the global bound to N+Δ, and a
// real solver should get at or very close to it.
func TestLowerBoundGlobalParityCorner(t *testing.T) {
	var active []string
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if (x == 0 && y == 0) || (x == 1 && y == 1) {
				continue // remove two same-color (black) cells near a corner
			}
			active = append(active, grid.VertexID(x, y))
		}
	}
	g, err := grid.Build(grid.GridSpec{Cols: 4, Rows: 4, ActiveVertices: active})
	if err != nil {
		t.Fatalf("unexpected error building grid: %v", err)
	}

	black, white := 0, 0
	for _, v := range g.Vertices() {
		if (v.X+v.Y)%2 == 0 {
			black++
		} else {
			white++
		}
	}
	delta := black - white
	if delta < 0 {
		delta = -delta
	}
	if delta != 2 {
		t.Fatalf("test construction error: expected delta=2, got %d (black=%d white=%d)", delta, black, white)
	}
	n := black + white

	problem := optimization.Problem{
		Graph:         g,
		StartPoints:   []graph.VertexID{"2,0"},
		ReturnToStart: true,
	}
	lb := optimization.LowerBoundGlobal(problem)
	if lb != n+delta {
		t.Fatalf("expected lower bound %d, got %d", n+delta, lb)
	}

	solution, err := optimization.VoronoiTree{}.Solve(problem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if solution.Metrics.Makespan < lb {
		t.Fatalf("makespan %d must not be below the lower bound %d", solution.Metrics.Makespan, lb)
	}
	if solution.Metrics.Makespan > lb+4 {
		t.Fatalf("expected the router to land close to the lower bound %d, got makespan %d", lb, solution.Metrics.Makespan)
	}
}

func TestLowerBoundGlobalRejectsNothingForTrivialProblem(t *testing.T) {
	// A single-cell board/start is its own trivial solution; the bound
	// must not panic or go negative.
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})
	problem := optimization.Problem{Graph: g, StartPoints: []graph.VertexID{"0,0"}}
	if lb := optimization.LowerBoundGlobal(problem); lb < 0 {
		t.Fatalf("expected a non-negative lower bound, got %d", lb)
	}
}
