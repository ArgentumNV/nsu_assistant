// Command server starts the NSU Assistant HTTP API on port 8080.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ArgentumNV/nsu_assistant/backend/internal/api"
	"github.com/ArgentumNV/nsu_assistant/backend/internal/user"
	"github.com/ArgentumNV/nsu_assistant/backend/storage"
)

// main registers the API routes and runs the HTTP server until it fails.
func main() {
	log.Println("Starting NSU Assistant...")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	log.Println("Connecting to PostgreSQL...")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	pool, err := storage.NewPostgres(ctx, databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL: ", err)
	}
	defer pool.Close()

	log.Println("Connected to PostgreSQL!")

	repo := user.NewRepository(pool)
	router := api.NewRouter(repo)

	log.Println("Starting HTTP server on port 8080...")

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Println("HTTP server error:", err)
	}
}
