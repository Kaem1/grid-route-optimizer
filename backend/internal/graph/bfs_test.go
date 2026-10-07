package graph_test

import (
	"testing"

	"grid-route-optimizer/internal/graph"
)

// gridGraph builds a cols x rows 4-neighborhood grid graph with no walls.
func gridGraph(cols, rows int) *graph.Graph {
	g := graph.New()
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			id := graph.VertexID(coordID(x, y))
			g.AddVertex(graph.Vertex{ID: id, X: x, Y: y})
		}
	}
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			id := graph.VertexID(coordID(x, y))
			if x+1 < cols {
				g.AddEdge(id, graph.VertexID(coordID(x+1, y)))
			}
			if y+1 < rows {
				g.AddEdge(id, graph.VertexID(coordID(x, y+1)))
			}
		}
	}
	return g
}

func coordID(x, y int) string {
	return string(rune('a'+x)) + "," + string(rune('a'+y))
}

func TestMultiSourceBFSAssignsEveryVertex(t *testing.T) {
	g := gridGraph(4, 1)
	owner, dist := graph.MultiSourceBFS(g, []graph.VertexID{
		graph.VertexID(coordID(0, 0)),
		graph.VertexID(coordID(3, 0)),
	})

	if len(owner) != 4 {
		t.Fatalf("expected all 4 cells assigned, got %d", len(owner))
	}
	// midpoint cells are equidistant; first-discovery order (source 0's
	// frontier is processed before source 1's at the same BFS layer)
	// must win the tie deterministically.
	if owner[graph.VertexID(coordID(1, 0))] != 0 {
		t.Fatalf("expected cell 1 to be owned by source 0, got %d", owner[graph.VertexID(coordID(1, 0))])
	}
	if dist[graph.VertexID(coordID(1, 0))] != 1 {
		t.Fatalf("expected distance 1, got %d", dist[graph.VertexID(coordID(1, 0))])
	}
}

func TestMultiSourceBFSOwnsItsOwnSource(t *testing.T) {
	g := gridGraph(2, 2)
	sources := []graph.VertexID{
		graph.VertexID(coordID(0, 0)),
		graph.VertexID(coordID(1, 1)),
	}
	owner, dist := graph.MultiSourceBFS(g, sources)
	for i, s := range sources {
		if owner[s] != i || dist[s] != 0 {
			t.Fatalf("expected source %q to own itself at distance 0, got owner=%d dist=%d", s, owner[s], dist[s])
		}
	}
}

func TestMultiSourceBFSLeavesUnreachableCellsUnassigned(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})
	g.AddVertex(graph.Vertex{ID: "5,5", X: 5, Y: 5}) // disconnected

	owner, _ := graph.MultiSourceBFS(g, []graph.VertexID{"0,0"})
	if _, ok := owner["5,5"]; ok {
		t.Fatal("expected disconnected vertex to be unassigned")
	}
}

func TestMultiSourceBFSDeduplicatesSources(t *testing.T) {
	g := gridGraph(2, 1)
	owner, _ := graph.MultiSourceBFS(g, []graph.VertexID{
		graph.VertexID(coordID(0, 0)),
		graph.VertexID(coordID(0, 0)),
	})
	if owner[graph.VertexID(coordID(0, 0))] != 0 {
		t.Fatal("expected the earliest index to win for duplicate sources")
	}
}
