package api

import (
	"encoding/json"
	"log"
	"net/http"
)

type ScheduleHandler struct {
}

type Lesson struct {
	Subject   string `json:"subject"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Room      string `json:"room"`
}

type ScheduleResponse struct {
	Group   string   `json:"group"`
	Date    string   `json:"date"`
	Lessons []Lesson `json:"lessons"`
}

func (h ScheduleHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	group := r.URL.Query().Get("group")
	date := r.URL.Query().Get("date")

	if group == "" {
		http.Error(w, "Missing group", http.StatusBadRequest)
		return
	} else if date == "" {
		http.Error(w, "Missing date", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// TODO: get info from web-site
	// if the source web-site is closed return 502 error

	// test struct
	lesson := Lesson{
		Subject:   "subject",
		StartTime: "starttime",
		EndTime:   "endtime",
		Room:      "room",
	}
	// test struct
	response := ScheduleResponse{
		Group:   group,
		Date:    date,
		Lessons: []Lesson{lesson},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("JSON encoding error:", err)
	}

}
