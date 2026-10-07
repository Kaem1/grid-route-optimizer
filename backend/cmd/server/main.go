// Command server runs the Grid Route Optimizer backend HTTP API.
package main

import (
	"log"
	"net/http"

	"grid-route-optimizer/internal/api"
	"grid-route-optimizer/internal/optimization"
)

const (
	addr          = ":8080"
	allowedOrigin = "http://localhost:5173"
)

func main() {
	// Register additional methods (M3, M6 from the algorithm spec need a
	// CP-SAT binding, M5 needs a VRP solver — none wired into the project)
	// here as they are implemented; the API/frontend already list whatever
	// is registered without further changes.
	registry := optimization.NewRegistry(
		optimization.VoronoiTree{},
		optimization.WeightedVoronoi{},
		optimization.LocalSearch{},
		optimization.TSPBaseline{},
	)
	service := optimization.NewService(registry)
	server := api.NewServer(service, allowedOrigin)

	log.Println("Backend running on http://localhost:8080")
	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
