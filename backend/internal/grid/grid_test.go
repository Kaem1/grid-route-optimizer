package grid_test

import (
	"testing"

	"grid-route-optimizer/internal/grid"
)

func TestBuildConnectsAdjacentActiveVertices(t *testing.T) {
	spec := grid.GridSpec{
		Cols:           2,
		Rows:           1,
		ActiveVertices: []string{"0,0", "1,0"},
	}

	g, err := grid.Build(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !g.HasEdge("0,0", "1,0") {
		t.Fatal("expected adjacent active cells to be connected")
	}
}

func TestBuildLeavesInactiveCellsOut(t *testing.T) {
	spec := grid.GridSpec{
		Cols:           2,
		Rows:           1,
		ActiveVertices: []string{"0,0"},
	}

	g, err := grid.Build(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.HasVertex("1,0") {
		t.Fatal("expected inactive cell to be absent from the graph")
	}
}

func TestBuildWallBlocksPassage(t *testing.T) {
	spec := grid.GridSpec{
		Cols:           2,
		Rows:           1,
		ActiveVertices: []string{"0,0", "1,0"},
		Walls:          []string{grid.EdgeKey("0,0", "1,0")},
	}

	g, err := grid.Build(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.HasEdge("0,0", "1,0") {
		t.Fatal("expected a wall to remove the passage between two cells")
	}
	if !g.HasVertex("0,0") || !g.HasVertex("1,0") {
		t.Fatal("a wall must not remove the cells themselves")
	}
}

func TestBuildRejectsNonPositiveDimensions(t *testing.T) {
	_, err := grid.Build(grid.GridSpec{Cols: 0, Rows: 5})
	if err == nil {
		t.Fatal("expected an error for a zero-width grid")
	}
}

func TestBuildRejectsOutOfBoundsVertex(t *testing.T) {
	_, err := grid.Build(grid.GridSpec{
		Cols:           2,
		Rows:           2,
		ActiveVertices: []string{"5,5"},
	})
	if err == nil {
		t.Fatal("expected an error for a vertex outside the grid bounds")
	}
}

func TestBuildRejectsWallOnInactiveCell(t *testing.T) {
	_, err := grid.Build(grid.GridSpec{
		Cols:           2,
		Rows:           1,
		ActiveVertices: []string{"0,0"},
		Walls:          []string{grid.EdgeKey("0,0", "1,0")},
	})
	if err == nil {
		t.Fatal("expected an error for a wall touching a cell that is not on the board")
	}
}

func TestEdgeKeyIsOrderIndependent(t *testing.T) {
	if grid.EdgeKey("0,0", "1,0") != grid.EdgeKey("1,0", "0,0") {
		t.Fatal("expected EdgeKey to be symmetric")
	}
}
