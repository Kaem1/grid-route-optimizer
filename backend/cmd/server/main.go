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
	registry := optimization.NewRegistry(optimization.DFSCoverage{})
	service := optimization.NewService(registry)
	server := api.NewServer(service, allowedOrigin)

	log.Println("Backend running on http://localhost:8080")
	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
