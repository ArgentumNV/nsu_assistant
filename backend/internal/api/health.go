// This file implements the health endpoint used to check server availability.
package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// HealthResponse contains the server health status returned as JSON.
type HealthResponse struct {
	Status string `json:"status"`
}

// HealthHandler handles requests to the /health endpoint.
type HealthHandler struct{}

// ServeHTTP returns the server health status for GET requests.
func (h HealthHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := HealthResponse{
		Status: "ok",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("JSON encoding error:", err)
	}
}
