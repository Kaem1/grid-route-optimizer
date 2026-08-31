package optimization_test

import (
	"testing"

	"grid-route-optimizer/internal/graph"
	"grid-route-optimizer/internal/optimization"
)

func lineGraph() *graph.Graph {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})
	g.AddVertex(graph.Vertex{ID: "1,0", X: 1, Y: 0})
	g.AddVertex(graph.Vertex{ID: "2,0", X: 2, Y: 0})
	g.AddEdge("0,0", "1,0")
	g.AddEdge("1,0", "2,0")
	return g
}

func TestDFSCoverageSingleVertex(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})

	solution, err := optimization.DFSCoverage{}.Solve(optimization.Problem{
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

func TestDFSCoverageVisitsEveryReachableVertex(t *testing.T) {
	solution, err := optimization.DFSCoverage{}.Solve(optimization.Problem{
		Graph:       lineGraph(),
		StartPoints: []graph.VertexID{"0,0"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	route := solution.Routes[0]
	visited := make(map[graph.VertexID]bool)
	for _, v := range route.Vertices {
		visited[v] = true
	}
	for _, id := range []graph.VertexID{"0,0", "1,0", "2,0"} {
		if !visited[id] {
			t.Fatalf("expected %q to be visited, walk was %v", id, route.Vertices)
		}
	}
	if route.Length != len(route.Vertices)-1 {
		t.Fatalf("route length %d does not match walk size %d", route.Length, len(route.Vertices))
	}
	if solution.Metrics.CoverageRatio != 1.0 {
		t.Fatalf("expected full coverage, got %v", solution.Metrics.CoverageRatio)
	}
}

func TestDFSCoverageRejectsUnknownStartPoint(t *testing.T) {
	_, err := optimization.DFSCoverage{}.Solve(optimization.Problem{
		Graph:       lineGraph(),
		StartPoints: []graph.VertexID{"9,9"},
	})
	if err == nil {
		t.Fatal("expected an error for a start point outside the grid")
	}
}

func TestDFSCoverageRejectsNoStartPoints(t *testing.T) {
	_, err := optimization.DFSCoverage{}.Solve(optimization.Problem{Graph: lineGraph()})
	if err == nil {
		t.Fatal("expected an error when no start points are given")
	}
}

func TestDFSCoverageRejectsNilGraph(t *testing.T) {
	_, err := optimization.DFSCoverage{}.Solve(optimization.Problem{
		StartPoints: []graph.VertexID{"0,0"},
	})
	if err == nil {
		t.Fatal("expected an error for a nil graph")
	}
}
