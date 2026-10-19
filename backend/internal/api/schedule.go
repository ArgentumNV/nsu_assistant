// This file implements the group directory and daily schedule HTTP endpoints.
package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ArgentumNV/nsu_assistant/backend/internal/schedule"
)

// scheduleSource is the shared client for the official NSU schedule website.
var scheduleSource = schedule.NewClient("https://table.nsu.ru", nil)

// ScheduleHandler serves the daily schedule of a selected IIR group.
type ScheduleHandler struct{}

// Lesson aliases the schedule model used in API responses.
type Lesson = schedule.Lesson

// ScheduleResponse contains a group, a requested date, and its scheduled lessons.
type ScheduleResponse struct {
	Group   string   `json:"group"`
	Date    string   `json:"date"`
	Lessons []Lesson `json:"lessons"`
}

// ServeHTTP validates the group and date, fetches lessons, and returns JSON.
// Dates use YYYY-MM-DD in the Novosibirsk timezone; source failures return HTTP 502.
func (h ScheduleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	group := strings.TrimSpace(r.URL.Query().Get("group"))
	date := r.URL.Query().Get("date")

	if group == "" {
		http.Error(w, "Missing group", http.StatusBadRequest)
		return
	}

	day, err := time.ParseInLocation(time.DateOnly, date, schedule.Location)
	if err != nil {
		http.Error(w, "Invalid date: expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	lessons, err := scheduleSource.Lessons(r.Context(), group, day)
	if errors.Is(err, schedule.ErrGroupNotFound) {
		http.Error(w, "Group not found in IIR", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("Schedule source error:", err)
		http.Error(w, "Schedule source unavailable", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(ScheduleResponse{Group: group, Date: date, Lessons: lessons}); err != nil {
		log.Println("JSON encoding error:", err)
	}
}

// GroupsHandler lists groups of the IIR faculty.
type GroupsHandler struct{}

// ServeHTTP returns the sorted IIR group directory as JSON for GET requests.
func (h GroupsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	groups, err := scheduleSource.Groups(r.Context())
	if err != nil {
		log.Println("Schedule groups error:", err)
		http.Error(w, "Schedule source unavailable", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(struct {
		Groups []string `json:"groups"`
	}{Groups: groups}); err != nil {
		log.Println("JSON encoding error:", err)
	}
}
