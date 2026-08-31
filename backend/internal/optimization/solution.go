package optimization

import "grid-route-optimizer/internal/routes"

// Metrics summarizes solution quality (see project instructions, section
// 40). More metrics (execution time, repeated vertices, ...) can be added
// without breaking existing callers. This is a domain type, not a wire
// DTO (see project instructions, section 46) — the API layer has its own
// MetricsDTO and converts explicitly.
type Metrics struct {
	TotalLength     int
	NumberOfRoutes  int
	CoveredVertices int
	CoverageRatio   float64
}

// Solution is the result of running an Algorithm on a Problem (see project
// instructions, section 39).
type Solution struct {
	Routes  []routes.Route
	Metrics Metrics
}
