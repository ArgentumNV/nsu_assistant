// Command server starts the NSU Assistant HTTP API on port 8080.
package main

import (
	"log"
	"net/http"

	"github.com/ArgentumNV/nsu_assistant/backend/internal/api"
)

// main registers the API routes and runs the HTTP server until it fails.
func main() {
	router := api.NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}
