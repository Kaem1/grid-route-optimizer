// Package optimization contains the optimization problem/solution domain
// model and the algorithms that solve it. Nothing here depends on HTTP,
// JSON, React, or PixiJS (see project instructions, sections 37/68), so
// algorithms can be unit tested and benchmarked in isolation.
package optimization

import "grid-route-optimizer/internal/graph"

// Problem is what an Algorithm must solve: cover the graph's vertices with
// routes starting at the given points (see project instructions, section
// 35). Constraints (max routes, max length, ...) can be added here later
// without changing the Algorithm interface.
type Problem struct {
	Graph       *graph.Graph
	StartPoints []graph.VertexID
}
