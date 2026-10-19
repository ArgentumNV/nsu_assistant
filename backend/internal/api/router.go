// Package api exposes HTTP endpoints for health checks, IIR groups, and schedules.
package api

import (
	"net/http"

	"github.com/ArgentumNV/nsu_assistant/backend/internal/user"
)

// NewRouter registers all API handlers and returns the HTTP request multiplexer.
func NewRouter(repo *user.Repository) *http.ServeMux {
	mux := http.NewServeMux()

	userHandler := NewUserHandler(repo)

	mux.Handle("/health", HealthHandler{})
	mux.Handle("/api/v1/schedule", ScheduleHandler{})
	mux.Handle("/api/v1/groups", GroupsHandler{})
	mux.HandleFunc("/api/v1/users", userHandler.CreateUser)
	mux.HandleFunc("/api/v1/users/{telegram_id}", userHandler.GetUserByTelegramID)

	return mux
}
