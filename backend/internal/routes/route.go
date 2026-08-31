// Package routes holds the Route domain model (see project instructions,
// section 31): a route is a real domain object, not just a visual effect.
package routes

import "grid-route-optimizer/internal/graph"

// Route is a walk performed by one simulated agent starting at StartPoint.
// It may revisit vertices/edges while covering the board, so Vertices is a
// walk (a sequence of steps), not a simple set.
type Route struct {
	StartPoint graph.VertexID
	Vertices   []graph.VertexID
	Length     int // number of edges traversed, i.e. len(Vertices)-1
}
