package optimization_test

import (
	"testing"

	"grid-route-optimizer/internal/graph"
	"grid-route-optimizer/internal/grid"
	"grid-route-optimizer/internal/optimization"
)

func wholeGraphRegion(g *graph.Graph) map[graph.VertexID]bool {
	region := make(map[graph.VertexID]bool, g.VertexCount())
	for _, v := range g.Vertices() {
		region[v.ID] = true
	}
	return region
}

// TestRouteRegion4x4ClosedFindsHamiltonianCycle is the patch's hard
// regression test: an even x even rectangle always has a Hamiltonian
// cycle, so 2-opt/Or-opt must find a route of cost exactly 16 (n=16 cells,
// closed walk of cost n). If this regresses to more than 16, there's a bug
// in two_opt/or_opt.
func TestRouteRegion4x4ClosedFindsHamiltonianCycle(t *testing.T) {
	g := buildGrid(4, 4)
	region := wholeGraphRegion(g)

	walk, err := optimization.RouteRegion(g, region, "0,0", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cost := len(walk) - 1; cost != 16 {
		t.Fatalf("expected a Hamiltonian cycle of cost 16, got %d (%v)", cost, walk)
	}
	validateWalk(t, g, "0,0", walk, true)
}

// TestRouteRegion4x4OpenFindsHamiltonianPath: a Hamiltonian path from a
// corner of a 4x4 rectangle exists, so the open-walk cost must reach
// exactly 15 (n-1).
func TestRouteRegion4x4OpenFindsHamiltonianPath(t *testing.T) {
	g := buildGrid(4, 4)
	region := wholeGraphRegion(g)

	walk, err := optimization.RouteRegion(g, region, "0,0", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cost := len(walk) - 1; cost != 15 {
		t.Fatalf("expected a Hamiltonian path of cost 15, got %d (%v)", cost, walk)
	}
	validateWalk(t, g, "0,0", walk, false)
}

func TestRouteRegionSingleAndTwoCellRegions(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})
	g.AddVertex(graph.Vertex{ID: "1,0", X: 1, Y: 0})
	g.AddEdge("0,0", "1,0")

	single := map[graph.VertexID]bool{"0,0": true}
	walk, err := optimization.RouteRegion(g, single, "0,0", false)
	if err != nil || len(walk) != 1 {
		t.Fatalf("expected a single-cell walk, got %v (err=%v)", walk, err)
	}

	pair := map[graph.VertexID]bool{"0,0": true, "1,0": true}
	walk, err = optimization.RouteRegion(g, pair, "0,0", false)
	if err != nil || len(walk)-1 != 1 {
		t.Fatalf("expected a 2-cell walk of cost 1, got %v (err=%v)", walk, err)
	}
}

func TestRouteRegionRejectsDisconnectedRegion(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})
	g.AddVertex(graph.Vertex{ID: "5,5", X: 5, Y: 5})
	g.AddVertex(graph.Vertex{ID: "6,5", X: 6, Y: 5})
	region := map[graph.VertexID]bool{"0,0": true, "5,5": true, "6,5": true}

	if _, err := optimization.RouteRegion(g, region, "0,0", false); err == nil {
		t.Fatal("expected an error for a disconnected region")
	}
}

// TestRouteRegionNeverWorseThanWalkRegion is the patch's comparison
// regression test: route_region (2-opt/Or-opt) must never cost more than
// walk_region (spanning-tree doubling) on a board with cycles.
func TestRouteRegionNeverWorseThanWalkRegion(t *testing.T) {
	sizes := [][2]int{{4, 4}, {5, 3}, {6, 6}, {3, 7}, {8, 4}}
	for _, size := range sizes {
		g := buildGrid(size[0], size[1])
		region := wholeGraphRegion(g)
		for _, returnToStart := range []bool{false, true} {
			walkWalk, err := optimization.WalkRegion(g, region, "0,0", returnToStart)
			if err != nil {
				t.Fatalf("size=%v return=%v: WalkRegion error: %v", size, returnToStart, err)
			}
			routeWalk, err := optimization.RouteRegion(g, region, "0,0", returnToStart)
			if err != nil {
				t.Fatalf("size=%v return=%v: RouteRegion error: %v", size, returnToStart, err)
			}
			walkCost, routeCost := len(walkWalk)-1, len(routeWalk)-1
			if routeCost > walkCost {
				t.Fatalf("size=%v return=%v: route cost %d > walk cost %d", size, returnToStart, routeCost, walkCost)
			}
			validateWalk(t, g, "0,0", routeWalk, returnToStart)
		}
	}
}

// TestRouteRegionCShapeRespectsHole builds a board with a notch cut out of
// the middle (forcing a detour around it) and checks RouteRegion only ever
// steps through cells inside the region, still covers everything, and is
// never worse than WalkRegion.
func TestRouteRegionCShapeRespectsHole(t *testing.T) {
	skip := map[[2]int]bool{{2, 1}: true, {2, 2}: true, {2, 3}: true}
	var active []string
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if skip[[2]int{x, y}] {
				continue
			}
			active = append(active, grid.VertexID(x, y))
		}
	}
	g, err := grid.Build(grid.GridSpec{Cols: 5, Rows: 5, ActiveVertices: active})
	if err != nil {
		t.Fatalf("unexpected error building grid: %v", err)
	}
	region := wholeGraphRegion(g)

	for _, returnToStart := range []bool{false, true} {
		walkWalk, err := optimization.WalkRegion(g, region, "0,0", returnToStart)
		if err != nil {
			t.Fatalf("return=%v: WalkRegion error: %v", returnToStart, err)
		}
		routeWalk, err := optimization.RouteRegion(g, region, "0,0", returnToStart)
		if err != nil {
			t.Fatalf("return=%v: RouteRegion error: %v", returnToStart, err)
		}

		validateWalk(t, g, "0,0", routeWalk, returnToStart)
		covered := map[graph.VertexID]bool{}
		for _, v := range routeWalk {
			covered[v] = true
		}
		if len(covered) != g.VertexCount() {
			t.Fatalf("return=%v: expected every cell covered, got %d/%d", returnToStart, len(covered), g.VertexCount())
		}
		if routeCost, walkCost := len(routeWalk)-1, len(walkWalk)-1; routeCost > walkCost {
			t.Fatalf("return=%v: route cost %d > walk cost %d", returnToStart, routeCost, walkCost)
		}
	}
}

func TestRouteRegionIsDeterministic(t *testing.T) {
	g := buildGrid(6, 5)
	region := wholeGraphRegion(g)

	first, err := optimization.RouteRegion(g, region, "0,0", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := optimization.RouteRegion(g, region, "0,0", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(first) != len(second) {
		t.Fatalf("expected identical walks across runs, got lengths %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("expected identical walks across runs, diverged at index %d: %v vs %v", i, first, second)
		}
	}
}
