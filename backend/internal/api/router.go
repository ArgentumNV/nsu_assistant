package api

import (
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/health", HealthHandler{})
	mux.Handle("/api/v1/schedule", ScheduleHandler{})

	return mux
}
