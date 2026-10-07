package optimization_test

import (
	"testing"

	"grid-route-optimizer/internal/graph"
	"grid-route-optimizer/internal/optimization"
)

// chainTree builds root -> a -> b -> c (a simple 4-node chain).
func chainTree() map[graph.VertexID][]graph.VertexID {
	return map[graph.VertexID][]graph.VertexID{
		"root": {"a"},
		"a":    {"b"},
		"b":    {"c"},
	}
}

func TestWalkTreeReturnToStartCost(t *testing.T) {
	tree := chainTree()
	walk := optimization.WalkTree(tree, "root", true)
	// n=4, cost = 2(n-1) = 6, so walk length = 7.
	if len(walk)-1 != 6 {
		t.Fatalf("expected cost 6, got %d (%v)", len(walk)-1, walk)
	}
	if walk[0] != "root" || walk[len(walk)-1] != "root" {
		t.Fatalf("expected walk to start and end at root, got %v", walk)
	}
}

func TestWalkTreeNoReturnCostUsesHeight(t *testing.T) {
	tree := chainTree()
	walk := optimization.WalkTree(tree, "root", false)
	// height(root) = 3 (a chain), cost = 2(n-1) - height = 6 - 3 = 3.
	if len(walk)-1 != 3 {
		t.Fatalf("expected cost 3, got %d (%v)", len(walk)-1, walk)
	}
	if walk[0] != "root" || walk[len(walk)-1] != "c" {
		t.Fatalf("expected walk to end at the deepest leaf, got %v", walk)
	}
}

func TestWalkTreeSingleNode(t *testing.T) {
	tree := map[graph.VertexID][]graph.VertexID{}
	for _, returnToStart := range []bool{true, false} {
		walk := optimization.WalkTree(tree, "root", returnToStart)
		if len(walk) != 1 || walk[0] != "root" {
			t.Fatalf("expected a single-node walk, got %v", walk)
		}
	}
}

func TestWalkTreeBranchingVisitsEveryNode(t *testing.T) {
	// root has two children: a (leaf) and b (which has leaf child c).
	tree := map[graph.VertexID][]graph.VertexID{
		"root": {"a", "b"},
		"b":    {"c"},
	}
	for _, returnToStart := range []bool{true, false} {
		walk := optimization.WalkTree(tree, "root", returnToStart)
		visited := map[graph.VertexID]bool{}
		for _, v := range walk {
			visited[v] = true
		}
		for _, id := range []graph.VertexID{"root", "a", "b", "c"} {
			if !visited[id] {
				t.Fatalf("returnToStart=%v: expected %q visited, got %v", returnToStart, id, walk)
			}
		}
		for i := 1; i < len(walk); i++ {
			if walk[i] == walk[i-1] {
				t.Fatalf("walk has a stationary step at %d: %v", i, walk)
			}
		}
	}
}

func lineRegion(ids ...graph.VertexID) (*graph.Graph, map[graph.VertexID]bool) {
	g := graph.New()
	for i, id := range ids {
		x, y := i, 0
		g.AddVertex(graph.Vertex{ID: id, X: x, Y: y})
	}
	for i := 0; i+1 < len(ids); i++ {
		g.AddEdge(ids[i], ids[i+1])
	}
	region := make(map[graph.VertexID]bool, len(ids))
	for _, id := range ids {
		region[id] = true
	}
	return g, region
}

func TestBuildTreeRejectsDisconnectedRegion(t *testing.T) {
	g := graph.New()
	g.AddVertex(graph.Vertex{ID: "0,0", X: 0, Y: 0})
	g.AddVertex(graph.Vertex{ID: "5,5", X: 5, Y: 5})
	region := map[graph.VertexID]bool{"0,0": true, "5,5": true}

	if _, err := optimization.BuildTree(g, region, "0,0", optimization.TreeBFS); err == nil {
		t.Fatal("expected an error for a disconnected region")
	}
}

func TestWalkRegionCorridorNoReturn(t *testing.T) {
	g, region := lineRegion("0,0", "1,0", "2,0", "3,0")
	walk, err := optimization.WalkRegion(g, region, "0,0", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// a straight corridor is a Hamiltonian path: cost = n-1 = 3.
	if len(walk)-1 != 3 {
		t.Fatalf("expected cost 3 for a corridor, got %d (%v)", len(walk)-1, walk)
	}
}

func TestWalkRegionCorridorReturn(t *testing.T) {
	g, region := lineRegion("0,0", "1,0", "2,0", "3,0")
	walk, err := optimization.WalkRegion(g, region, "0,0", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// cost = 2(n-1) = 6.
	if len(walk)-1 != 6 {
		t.Fatalf("expected cost 6 for a corridor with return, got %d (%v)", len(walk)-1, walk)
	}
	if walk[len(walk)-1] != "0,0" {
		t.Fatalf("expected walk to return to root, got %v", walk)
	}
}
