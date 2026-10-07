// Package graph provides the generic graph domain model shared by the grid
// builder and the optimization algorithms. It has no knowledge of HTTP,
// JSON, or any rendering technology (see project instructions, sections
// 19/65/68).
package graph

import "sort"

// VertexID uniquely identifies a vertex. For 2D grids it is the "x,y"
// string produced by grid.VertexID, matching the format the frontend
// already uses for its own vertex/edge identifiers.
type VertexID string

// Vertex is the domain representation of a single grid cell (see project
// instructions, section 17). Z is intentionally omitted for now but adding
// it later would not require touching this package's API shape.
type Vertex struct {
	ID VertexID
	X  int
	Y  int
}

// Graph is an undirected graph: a passage between A and B is symmetric.
// Walls/missing passages are represented simply by the absence of an edge,
// not by a separate "blocked" flag (see project instructions, section 14).
type Graph struct {
	vertices  map[VertexID]Vertex
	adjacency map[VertexID]map[VertexID]struct{}
}

// New creates an empty graph.
func New() *Graph {
	return &Graph{
		vertices:  make(map[VertexID]Vertex),
		adjacency: make(map[VertexID]map[VertexID]struct{}),
	}
}

// AddVertex registers a vertex. Re-adding the same ID is a no-op.
func (g *Graph) AddVertex(v Vertex) {
	if _, exists := g.vertices[v.ID]; exists {
		return
	}
	g.vertices[v.ID] = v
	g.adjacency[v.ID] = make(map[VertexID]struct{})
}

// HasVertex reports whether id is part of the graph.
func (g *Graph) HasVertex(id VertexID) bool {
	_, ok := g.vertices[id]
	return ok
}

// Vertex returns the vertex for id, if present.
func (g *Graph) Vertex(id VertexID) (Vertex, bool) {
	v, ok := g.vertices[id]
	return v, ok
}

// VertexCount returns the number of vertices in the graph.
func (g *Graph) VertexCount() int {
	return len(g.vertices)
}

// AddEdge connects two existing vertices. Both endpoints must already be
// present; otherwise the call is a no-op, which keeps callers from having to
// special-case grid boundaries.
func (g *Graph) AddEdge(a, b VertexID) {
	if a == b {
		return
	}
	if _, ok := g.adjacency[a]; !ok {
		return
	}
	if _, ok := g.adjacency[b]; !ok {
		return
	}
	g.adjacency[a][b] = struct{}{}
	g.adjacency[b][a] = struct{}{}
}

// HasEdge reports whether a passage exists between a and b.
func (g *Graph) HasEdge(a, b VertexID) bool {
	neighbors, ok := g.adjacency[a]
	if !ok {
		return false
	}
	_, ok = neighbors[b]
	return ok
}

// Neighbors returns the vertex IDs directly reachable from id, sorted by
// coordinate so traversals (e.g. the DFS coverage algorithm) are
// deterministic and reproducible across runs.
func (g *Graph) Neighbors(id VertexID) []VertexID {
	neighbors, ok := g.adjacency[id]
	if !ok {
		return nil
	}

	out := make([]VertexID, 0, len(neighbors))
	for n := range neighbors {
		out = append(out, n)
	}

	sort.Slice(out, func(i, j int) bool {
		vi, vj := g.vertices[out[i]], g.vertices[out[j]]
		if vi.Y != vj.Y {
			return vi.Y < vj.Y
		}
		if vi.X != vj.X {
			return vi.X < vj.X
		}
		return out[i] < out[j]
	})

	return out
}

// Vertices returns every vertex in the graph, sorted by (Y,X,ID) for
// deterministic iteration (matching Neighbors' ordering convention).
func (g *Graph) Vertices() []Vertex {
	out := make([]Vertex, 0, len(g.vertices))
	for _, v := range g.vertices {
		out = append(out, v)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Y != out[j].Y {
			return out[i].Y < out[j].Y
		}
		if out[i].X != out[j].X {
			return out[i].X < out[j].X
		}
		return out[i].ID < out[j].ID
	})

	return out
}
