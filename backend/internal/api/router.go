// Package api exposes HTTP endpoints for health checks, IIR groups, and schedules.
package api

import (
	"net/http"
)

// NewRouter registers all API handlers and returns the HTTP request multiplexer.
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/health", HealthHandler{})
	mux.Handle("/api/v1/schedule", ScheduleHandler{})
	mux.Handle("/api/v1/groups", GroupsHandler{})

	return mux
}
