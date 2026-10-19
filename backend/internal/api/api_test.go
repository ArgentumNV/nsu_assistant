// This file tests the public HTTP API using an isolated local NSU source.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ArgentumNV/nsu_assistant/backend/internal/schedule"
)

// newTestRouter replaces the shared source for one test and restores it during cleanup.
// Tests using this helper must remain sequential because scheduleSource is package-level state.
func newTestRouter(t *testing.T, upstream http.Handler) *http.ServeMux {
	t.Helper()
	server := httptest.NewServer(upstream)
	t.Cleanup(server.Close)
	original := scheduleSource
	scheduleSource = schedule.NewClient(server.URL, server.Client())
	t.Cleanup(func() { scheduleSource = original })
	// Schedule and group tests do not use the user endpoint or a PostgreSQL repository.
	return NewRouter(nil)
}

// TestHTTPAPIResponses verifies routes, JSON fields, empty arrays, and source failures.
func TestHTTPAPIResponses(t *testing.T) {
	calendar := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nBEGIN:VEVENT\r\n" +
		"SUMMARY:Робототехника (лекция)\r\nLOCATION:Аудитория: 412 КПА\r\n" +
		"DTSTART;TZID=Asia/Novosibirsk:20261019T090000\r\n" +
		"DTEND;TZID=Asia/Novosibirsk:20261019T103500\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	for _, tc := range []struct {
		name, path                        string
		wantStatus, sourceStatus          int
		invalidDirectory, invalidCalendar bool
	}{
		{name: "groups", path: "/api/v1/groups", wantStatus: 200},
		{name: "lessons", path: "/api/v1/schedule?group=24943&date=2026-10-19", wantStatus: 200},
		{name: "trimmed group", path: "/api/v1/schedule?group=%2024943%20&date=2026-10-19", wantStatus: 200},
		{name: "empty day", path: "/api/v1/schedule?group=24943&date=2026-10-25", wantStatus: 200},
		{name: "unknown group", path: "/api/v1/schedule?group=99999&date=2026-10-19", wantStatus: 404},
		{name: "groups source failure", path: "/api/v1/groups", sourceStatus: 503, wantStatus: 502},
		{name: "schedule source failure", path: "/api/v1/schedule?group=24943&date=2026-10-19", sourceStatus: 503, wantStatus: 502},
		{name: "invalid group directory", path: "/api/v1/groups", invalidDirectory: true, wantStatus: 502},
		{name: "invalid calendar", path: "/api/v1/schedule?group=24943&date=2026-10-19", invalidCalendar: true, wantStatus: 502},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := newTestRouter(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.sourceStatus != 0 {
					w.WriteHeader(tc.sourceStatus)
					return
				}
				switch r.URL.Path {
				case "/faculty/iir":
					if tc.invalidDirectory {
						fmt.Fprint(w, "<html>Error</html>")
						return
					}
					fmt.Fprint(w, `<a href="/group/26940">26940</a><a href="/group/24943">24943</a><a href="/group/24943">24943</a>`)
				case "/ics/group/24943":
					if tc.invalidCalendar {
						fmt.Fprint(w, "<html>Error</html>")
						return
					}
					fmt.Fprint(w, calendar)
				default:
					t.Errorf("unexpected source path: %s", r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if w.Code != tc.wantStatus {
				t.Fatalf("got %d: %s; want %d", w.Code, w.Body.String(), tc.wantStatus)
			}
			if tc.wantStatus != 200 {
				return
			}
			if got := w.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Fatalf("unexpected content type: %q", got)
			}
			if tc.name == "groups" {
				var body struct {
					Groups []string `json:"groups"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(body.Groups, []string{"24943", "26940"}) {
					t.Fatalf("unexpected groups: %#v", body.Groups)
				}
				return
			}
			var body struct {
				Group   string `json:"group"`
				Date    string `json:"date"`
				Lessons []struct {
					Subject   string `json:"subject"`
					StartTime string `json:"start_time"`
					EndTime   string `json:"end_time"`
					Room      string `json:"room"`
				} `json:"lessons"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			wantDate := "2026-10-19"
			if tc.name == "empty day" {
				wantDate = "2026-10-25"
			}
			if body.Group != "24943" || body.Date != wantDate || body.Lessons == nil {
				t.Fatalf("unexpected response: %s", w.Body.String())
			}
			if tc.name == "empty day" {
				if len(body.Lessons) != 0 {
					t.Fatalf("expected empty lessons: %s", w.Body.String())
				}
				return
			}
			if len(body.Lessons) != 1 {
				t.Fatalf("unexpected lesson count: %s", w.Body.String())
			}
			lesson := body.Lessons[0]
			if lesson.Subject != "Робототехника (лекция)" || lesson.StartTime != "09:00" || lesson.EndTime != "10:35" || lesson.Room != "412 КПА" {
				t.Fatalf("unexpected lesson: %#v", lesson)
			}
		})
	}
}

// TestHTTPAPIValidation verifies bad input and unsupported methods are rejected before contacting NSU.
func TestHTTPAPIValidation(t *testing.T) {
	var calls atomic.Int32
	router := newTestRouter(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); http.Error(w, "unexpected request", 500) }))
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/api/v1/schedule?date=2026-10-19", 400},
		{"GET", "/api/v1/schedule?group=%20%20&date=2026-10-19", 400},
		{"GET", "/api/v1/schedule?group=24943", 400},
		{"GET", "/api/v1/schedule?group=24943&date=wrong", 400},
		{"GET", "/api/v1/schedule?group=24943&date=2026-02-30", 400},
		{"GET", "/api/v1/schedule?group=24943&date=19.10.2026", 400},
		{"POST", "/api/v1/schedule", 405}, {"DELETE", "/api/v1/schedule", 405},
		{"POST", "/api/v1/groups", 405}, {"PUT", "/api/v1/groups", 405},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("got %d: %s, want %d", w.Code, w.Body.String(), tc.status)
			}
			if tc.status == 405 && w.Header().Get("Allow") != "GET" {
				t.Fatal("missing Allow: GET")
			}
			if tc.status == 400 && strings.TrimSpace(w.Body.String()) == "" {
				t.Fatal("missing validation error")
			}
		})
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("invalid requests contacted NSU %d times", got)
	}
}
