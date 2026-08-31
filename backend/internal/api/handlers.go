package api

import (
	"encoding/json"
	"log"
	"net/http"

	"grid-route-optimizer/internal/grid"
	"grid-route-optimizer/internal/optimization"
)

// Server wires HTTP handlers to the optimization service. It has no domain
// logic of its own (see project instructions, section 45).
type Server struct {
	simulation    *optimization.Service
	allowedOrigin string
}

// NewServer builds a Server that answers requests from allowedOrigin only.
func NewServer(simulation *optimization.Service, allowedOrigin string) *Server {
	return &Server{simulation: simulation, allowedOrigin: allowedOrigin}
}

// Routes builds the HTTP router for the whole API.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/algorithms", s.handleAlgorithms)
	mux.HandleFunc("/api/simulate", s.handleSimulate)
	return mux
}

func (s *Server) withCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", s.allowedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// preflight answers a CORS preflight OPTIONS request. It returns true if
// the caller should stop processing the request.
func (s *Server) preflight(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodOptions {
		return false
	}
	w.WriteHeader(http.StatusNoContent)
	return true
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("api: failed to encode response: %v", err)
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, ErrorResponse{Error: message})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.withCORS(w)
	if s.preflight(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "only GET is supported")
		return
	}

	s.writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func (s *Server) handleAlgorithms(w http.ResponseWriter, r *http.Request) {
	s.withCORS(w)
	if s.preflight(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "only GET is supported")
		return
	}

	s.writeJSON(w, http.StatusOK, AlgorithmsResponse{Algorithms: s.simulation.AlgorithmNames()})
}

func (s *Server) handleSimulate(w http.ResponseWriter, r *http.Request) {
	s.withCORS(w)
	if s.preflight(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "only POST is supported")
		return
	}

	var req SimulateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	spec := grid.GridSpec{
		Cols:           req.Grid.Cols,
		Rows:           req.Grid.Rows,
		ActiveVertices: req.Grid.ActiveVertices,
		Walls:          req.Grid.Walls,
	}

	solution, err := s.simulation.Simulate(spec, req.StartPoints, req.Algorithm)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, toSimulateResponse(req.Algorithm, solution))
}

func toSimulateResponse(algorithm string, solution optimization.Solution) SimulateResponse {
	routeDTOs := make([]RouteDTO, 0, len(solution.Routes))
	for _, route := range solution.Routes {
		vertices := make([]string, 0, len(route.Vertices))
		for _, v := range route.Vertices {
			vertices = append(vertices, string(v))
		}
		routeDTOs = append(routeDTOs, RouteDTO{
			StartPoint: string(route.StartPoint),
			Vertices:   vertices,
			Length:     route.Length,
		})
	}

	return SimulateResponse{
		Algorithm: algorithm,
		Routes:    routeDTOs,
		Metrics: MetricsDTO{
			TotalLength:     solution.Metrics.TotalLength,
			NumberOfRoutes:  solution.Metrics.NumberOfRoutes,
			CoveredVertices: solution.Metrics.CoveredVertices,
			CoverageRatio:   solution.Metrics.CoverageRatio,
		},
	}
}
