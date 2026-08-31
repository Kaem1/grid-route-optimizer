package graph_test

import (
	"testing"

	"grid-route-optimizer/internal/graph"
)

func TestAddEdgeRequiresBothVertices(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})

	// "1,0" was never added, so this must stay a no-op.
	g.AddEdge("0,0", "1,0")

	if g.HasEdge("0,0", "1,0") {
		t.Fatal("expected no edge to a vertex that was never added")
	}
}

func TestAddEdgeIsSymmetric(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})
	g.AddVertex(graph.Vertex{ID: "1,0", X: 1, Y: 0})
	g.AddEdge("0,0", "1,0")

	if !g.HasEdge("0,0", "1,0") || !g.HasEdge("1,0", "0,0") {
		t.Fatal("expected edge to be visible from both endpoints")
	}
}

func TestNeighborsOrderedByCoordinate(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "1,1", X: 1, Y: 1})
	g.AddVertex(graph.Vertex{ID: "2,1", X: 2, Y: 1})
	g.AddVertex(graph.Vertex{ID: "1,0", X: 1, Y: 0})
	g.AddVertex(graph.Vertex{ID: "1,2", X: 1, Y: 2})
	g.AddEdge("1,1", "2,1")
	g.AddEdge("1,1", "1,0")
	g.AddEdge("1,1", "1,2")

	got := g.Neighbors("1,1")
	want := []graph.VertexID{"1,0", "2,1", "1,2"}

	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestNeighborsOfUnknownVertexIsNil(t *testing.T) {
	g := graph.New()
	if got := g.Neighbors("0,0"); got != nil {
		t.Fatalf("expected nil neighbors for unknown vertex, got %v", got)
	}
}
