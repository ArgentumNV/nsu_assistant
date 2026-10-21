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

	mux.Handle("GET /health", HealthHandler{})

	mux.Handle("GET /api/v1/schedule", ScheduleHandler{})

	mux.Handle("GET /api/v1/groups", GroupsHandler{})

	mux.HandleFunc("POST /api/v1/users", userHandler.CreateUser)
	mux.HandleFunc("GET /api/v1/users/{telegram_id}", userHandler.GetUserByTelegramID)
	mux.HandleFunc("DELETE /api/v1/users/{telegram_id}", userHandler.DeleteUser)
	mux.HandleFunc("PATCH /api/v1/users/{telegram_id}", userHandler.UpdateUserGroup)

	return mux
}
