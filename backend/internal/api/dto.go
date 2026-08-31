// Package api contains the HTTP layer: wire DTOs and handlers. Domain logic
// itself lives in internal/optimization and internal/grid; handlers only
// translate between JSON and those packages (see project instructions,
// section 45).
package api

// GridDTO mirrors the frontend's BoardSnapshot: which cells are part of the
// board and where the walls are, using the same "x,y" / "a|b" string
// formats the frontend already produces (frontend/src/graph/grid.ts), so no
// translation layer is needed on either side of the wire.
type GridDTO struct {
	Cols           int      `json:"cols"`
	Rows           int      `json:"rows"`
	ActiveVertices []string `json:"activeVertices"`
	Walls          []string `json:"walls"`
}

// SimulateRequest is the payload sent when the user starts a simulation:
// the board, the chosen algorithm, and the agents' start points.
type SimulateRequest struct {
	Grid        GridDTO  `json:"grid"`
	StartPoints []string `json:"startPoints"`
	Algorithm   string   `json:"algorithm"`
}

// RouteDTO is one agent's computed walk.
type RouteDTO struct {
	StartPoint string   `json:"startPoint"`
	Vertices   []string `json:"vertices"`
	Length     int      `json:"length"`
}

// MetricsDTO is the wire form of optimization.Metrics.
type MetricsDTO struct {
	TotalLength     int     `json:"totalLength"`
	NumberOfRoutes  int     `json:"numberOfRoutes"`
	CoveredVertices int     `json:"coveredVertices"`
	CoverageRatio   float64 `json:"coverageRatio"`
}

// SimulateResponse is returned once the requested algorithm has finished.
type SimulateResponse struct {
	Algorithm string     `json:"algorithm"`
	Routes    []RouteDTO `json:"routes"`
	Metrics   MetricsDTO `json:"metrics"`
}

// AlgorithmsResponse lists the algorithms the frontend can choose from.
type AlgorithmsResponse struct {
	Algorithms []string `json:"algorithms"`
}

// ErrorResponse is returned for any request that fails validation.
type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthResponse is returned by GET /api/health.
type HealthResponse struct {
	Status string `json:"status"`
}
