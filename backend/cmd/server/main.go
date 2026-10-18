package main

/// Starts server

import (
	"log"
	"net/http"

	"github.com/ArgentumNV/nsu_assistant/backend/internal/api"
)

func main() {
	handler := api.HealthHandler{}
	http.Handle("/health", handler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
