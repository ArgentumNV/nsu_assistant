// This file tests the NSU client against a local HTTP source without internet access.
package schedule

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// TestClientGroups verifies nested links, sorting, duplicate removal, and link filtering.
func TestClientGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/faculty/iir" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		fmt.Fprint(w, `<html><body><section><a href="/group/26940">26940</a>
  <div><a href="/group/24943">24943</a><a href="/group/24943">duplicate</a></div>
  <a href="/faculty/fit">other faculty</a><a href="/group/">empty</a>
  <a href="/group/24943/extra">extra path</a><a href="/group/24943?q=1">query</a>
  <a href="/group/24943#anchor">anchor</a></section></body></html>`)
	}))
	defer server.Close()
	client := NewClient(server.URL+"/", server.Client())
	got, err := client.Groups(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"24943", "26940"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestClientLessons verifies membership checks, requested paths, and calendar error propagation.
func TestClientLessons(t *testing.T) {
	for _, tc := range []struct {
		name, group, directory, calendar string
		directoryStatus, calendarStatus  int
		wantErr, wantMissing             bool
		wantPaths                        []string
	}{
		{name: "success", group: "24943", wantPaths: []string{"/faculty/iir", "/ics/group/24943"}},
		{name: "unknown group", group: "99999", wantErr: true, wantMissing: true, wantPaths: []string{"/faculty/iir"}},
		{name: "directory unavailable", group: "24943", directoryStatus: 503, wantErr: true, wantPaths: []string{"/faculty/iir"}},
		{name: "directory format changed", group: "24943", directory: "<html>No groups</html>", wantErr: true, wantPaths: []string{"/faculty/iir"}},
		{name: "calendar unavailable", group: "24943", calendarStatus: 500, wantErr: true, wantPaths: []string{"/faculty/iir", "/ics/group/24943"}},
		{name: "invalid calendar", group: "24943", calendar: "<html>Error</html>", wantErr: true, wantPaths: []string{"/faculty/iir", "/ics/group/24943"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			var paths []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				paths = append(paths, r.URL.Path)
				mu.Unlock()
				switch r.URL.Path {
				case "/faculty/iir":
					if tc.directoryStatus != 0 {
						w.WriteHeader(tc.directoryStatus)
						return
					}
					body := tc.directory
					if body == "" {
						body = `<a href="/group/24943">24943</a>`
					}
					fmt.Fprint(w, body)
				case "/ics/group/24943":
					if tc.calendarStatus != 0 {
						w.WriteHeader(tc.calendarStatus)
						return
					}
					body := tc.calendar
					if body == "" {
						body = testCalendar(testEvent + "END:VEVENT\r\n")
					}
					fmt.Fprint(w, body)
				default:
					t.Errorf("unexpected path: %s", r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			got, err := NewClient(server.URL, server.Client()).Lessons(context.Background(), tc.group, testDay(t, "2026-09-07"))
			if (err != nil) != tc.wantErr {
				t.Fatalf("got error %v, want error %v", err, tc.wantErr)
			}
			if errors.Is(err, ErrGroupNotFound) != tc.wantMissing {
				t.Fatalf("unexpected membership error: %v", err)
			}
			if !tc.wantErr {
				want := []Lesson{{Subject: "Робототехника (лекция)", StartTime: "09:00", EndTime: "10:35", Room: "412 КПА"}}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("got %#v, want %#v", got, want)
				}
			}
			mu.Lock()
			defer mu.Unlock()
			if !reflect.DeepEqual(paths, tc.wantPaths) {
				t.Fatalf("got paths %v, want %v", paths, tc.wantPaths)
			}
		})
	}
}

// TestClientResponseLimit verifies both the accepted size boundary and oversized source rejection.
func TestClientResponseLimit(t *testing.T) {
	for _, tc := range []struct {
		name    string
		size    int
		wantErr bool
	}{
		{"at limit", 4 << 20, false}, {"above limit", (4 << 20) + 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, strings.Repeat("x", tc.size)) }))
			defer server.Close()
			body, err := NewClient(server.URL, server.Client()).get(context.Background(), "/large")
			if (err != nil) != tc.wantErr {
				t.Fatalf("got error %v, want error %v", err, tc.wantErr)
			}
			if !tc.wantErr && len(body) != tc.size {
				t.Fatalf("got %d bytes, want %d", len(body), tc.size)
			}
		})
	}
}

// TestClientContextCancellation verifies that cancellation reaches an in-flight HTTP request.
func TestClientContextCancellation(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { _, err := NewClient(server.URL, server.Client()).Groups(ctx); result <- err }()
	<-started
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}
