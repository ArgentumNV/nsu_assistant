package main

/// Starts server

import (
	"log"
	"net/http"

	"github.com/ArgentumNV/nsu_assistant/backend/internal/api"
)

func main() {
	router := api.NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}
