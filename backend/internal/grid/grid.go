// Package grid builds the domain graph for a rectangular 2D board out of a
// wire-friendly GridSpec. It is the "Krata 2D" from project instructions
// section 16, and doubles as the Subgraph builder from section 6/34: the
// frontend already resolved which cells are selected, so ActiveVertices here
// plays the role of the target subgraph's vertex set.
package grid

import (
	"fmt"
	"strconv"
	"strings"

	"grid-route-optimizer/internal/graph"
)

// GridSpec is the wire-friendly definition of a 2D grid board sent by the
// frontend: its dimensions, which cells are part of the board, and which
// adjacent cell pairs have a wall between them. IDs use the same "x,y" /
// "a|b" string formats produced by frontend/src/graph/grid.ts, so the API
// layer can pass request data through unchanged.
type GridSpec struct {
	Cols           int
	Rows           int
	ActiveVertices []string
	Walls          []string
}

// right/down are the only directions checked while connecting vertices;
// walking every active cell and only looking right/down still visits every
// adjacent pair exactly once.
var neighborOffsets = [2][2]int{{1, 0}, {0, 1}}

// VertexID formats grid coordinates into the canonical vertex identifier.
func VertexID(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

// ParseVertexID parses a canonical vertex identifier back into coordinates.
func ParseVertexID(id string) (x, y int, err error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid vertex id %q", id)
	}
	x, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid vertex id %q: %w", id, err)
	}
	y, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid vertex id %q: %w", id, err)
	}
	return x, y, nil
}

// EdgeKey builds the canonical, direction-independent key for the edge
// between two adjacent vertices (matches frontend/src/graph/grid.ts).
func EdgeKey(a, b string) string {
	if a < b {
		return a + "|" + b
	}
	return b + "|" + a
}

// ParseEdgeKey splits a canonical edge key into its two endpoint IDs.
func ParseEdgeKey(key string) (a, b string, err error) {
	parts := strings.SplitN(key, "|", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid edge key %q", key)
	}
	return parts[0], parts[1], nil
}

// Build converts a GridSpec into a domain graph: every active vertex
// becomes a graph.Vertex, and every pair of orthogonally adjacent active
// vertices is connected unless a wall blocks the passage between them (see
// project instructions, sections 8-14). The backend re-validates everything
// instead of trusting the frontend (section 47).
func Build(spec GridSpec) (*graph.Graph, error) {
	if spec.Cols <= 0 || spec.Rows <= 0 {
		return nil, fmt.Errorf("grid dimensions must be positive, got %dx%d", spec.Cols, spec.Rows)
	}

	active := make(map[string]bool, len(spec.ActiveVertices))
	g := graph.New()

	for _, id := range spec.ActiveVertices {
		x, y, err := ParseVertexID(id)
		if err != nil {
			return nil, err
		}
		if x < 0 || y < 0 || x >= spec.Cols || y >= spec.Rows {
			return nil, fmt.Errorf("vertex %q is outside the %dx%d grid", id, spec.Cols, spec.Rows)
		}
		if active[id] {
			continue
		}
		active[id] = true
		g.AddVertex(graph.Vertex{ID: graph.VertexID(id), X: x, Y: y})
	}

	walls := make(map[string]bool, len(spec.Walls))
	for _, key := range spec.Walls {
		a, b, err := ParseEdgeKey(key)
		if err != nil {
			return nil, err
		}
		if !active[a] || !active[b] {
			return nil, fmt.Errorf("wall %q references a cell that is not part of the board", key)
		}
		walls[EdgeKey(a, b)] = true
	}

	for id := range active {
		x, y, _ := ParseVertexID(id)
		for _, offset := range neighborOffsets {
			neighborID := VertexID(x+offset[0], y+offset[1])
			if !active[neighborID] {
				continue
			}
			if walls[EdgeKey(id, neighborID)] {
				continue
			}
			g.AddEdge(graph.VertexID(id), graph.VertexID(neighborID))
		}
	}

	return g, nil
}
