// This file tests ICS decoding, recurrence expansion, and local lesson times.
package schedule

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

const testEvent = "BEGIN:VEVENT\r\n" +
	"SUMMARY:Робототехника (лекция)\r\nLOCATION:Аудитория: 412 КПА\r\n" +
	"DTSTART;TZID=Asia/Novosibirsk:20260907T090000\r\n" +
	"DTEND;TZID=Asia/Novosibirsk:20260907T103500\r\n"

// testCalendar wraps event data in a complete calendar document.
func testCalendar(events string) string {
	return "BEGIN:VCALENDAR\r\nVERSION:2.0\r\n" + events + "END:VCALENDAR\r\n"
}

// testDay parses a test date in the same timezone as the public API.
func testDay(t *testing.T, value string) time.Time {
	t.Helper()
	day, err := time.ParseInLocation(time.DateOnly, value, Location)
	if err != nil {
		t.Fatal(err)
	}
	return day
}

// TestParseCalendarRecurrences verifies semester limits, alternating weeks, and date overrides.
func TestParseCalendarRecurrences(t *testing.T) {
	data := testCalendar(testEvent +
		"RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=MO;UNTIL=20261019T090000\r\n" +
		"EXDATE;TZID=Asia/Novosibirsk:20260921T090000,20261005T090000\r\n" +
		"RDATE;TZID=Asia/Novosibirsk:20260922T090000,20260923T090000\r\nEND:VEVENT\r\n")
	for _, tc := range []struct {
		date    string
		present bool
	}{
		{"2026-08-31", false}, {"2026-09-07", true}, {"2026-09-14", false},
		{"2026-09-21", false}, {"2026-09-22", true}, {"2026-09-23", true},
		{"2026-10-05", false}, {"2026-10-19", true}, {"2026-11-02", false},
	} {
		t.Run(tc.date, func(t *testing.T) {
			got, err := parseCalendar(strings.NewReader(data), testDay(t, tc.date))
			if err != nil {
				t.Fatal(err)
			}
			want := []Lesson{}
			if tc.present {
				want = append(want, Lesson{Subject: "Робототехника (лекция)", StartTime: "09:00", EndTime: "10:35", Room: "412 КПА"})
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v, want %#v", got, want)
			}
		})
	}
}

// TestParseCalendarOneOff verifies timezone conversion, duration, cancellation, and day boundaries.
func TestParseCalendarOneOff(t *testing.T) {
	data := testCalendar(
		"BEGIN:VEVENT\r\nSUMMARY:Поздняя\r\nDTSTART:20261019T034000Z\r\nDTEND:20261019T052500Z\r\nEND:VEVENT\r\n" +
			"BEGIN:VEVENT\r\nSUMMARY:Ранняя\r\nDTSTART:20261018T233000Z\r\nDURATION:PT1H\r\nEND:VEVENT\r\n" +
			"BEGIN:VEVENT\r\nSUMMARY:Полночь\r\nDTSTART:20261018T170000Z\r\nDURATION:PT1H\r\nEND:VEVENT\r\n" +
			"BEGIN:VEVENT\r\nSUMMARY:Завтра\r\nDTSTART:20261019T170000Z\r\nDURATION:PT1H\r\nEND:VEVENT\r\n" +
			"BEGIN:VEVENT\r\nSUMMARY:Отмена\r\nSTATUS:CANCELLED\r\nDTSTART:20261019T030000Z\r\nDURATION:PT1H\r\nEND:VEVENT\r\n")
	got, err := parseCalendar(strings.NewReader(data), testDay(t, "2026-10-19"))
	if err != nil {
		t.Fatal(err)
	}
	want := []Lesson{
		{Subject: "Полночь", StartTime: "00:00", EndTime: "01:00"},
		{Subject: "Ранняя", StartTime: "06:30", EndTime: "07:30"},
		{Subject: "Поздняя", StartTime: "10:40", EndTime: "12:25"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

// TestParseCalendarText verifies escaped text, folded lines, and parallel lesson ordering.
func TestParseCalendarText(t *testing.T) {
	data := testCalendar(
		strings.Replace(testEvent, "SUMMARY:Робототехника (лекция)", "SUMMARY:Робототехника\\, теория и\r\n практика", 1) + "END:VEVENT\r\n" +
			strings.Replace(testEvent, "SUMMARY:Робототехника (лекция)", "SUMMARY:Подгруппа 2", 1) + "END:VEVENT\r\n")
	got, err := parseCalendar(strings.NewReader(data), testDay(t, "2026-09-07"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Subject != "Робототехника, теория ипрактика" || got[1].Subject != "Подгруппа 2" {
		t.Fatalf("unexpected lessons: %#v", got)
	}
}

// TestParseCalendarEmpty verifies that a valid empty calendar produces an empty JSON-compatible slice.
func TestParseCalendarEmpty(t *testing.T) {
	got, err := parseCalendar(strings.NewReader(testCalendar("")), testDay(t, "2026-10-19"))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("expected non-nil empty slice, got %#v", got)
	}
}

// TestParseCalendarInvalid verifies that malformed source data is reported as an error.
func TestParseCalendarInvalid(t *testing.T) {
	for name, data := range map[string]string{
		"HTML instead of ICS": "<html>Unavailable</html>",
		"missing start":       testCalendar("BEGIN:VEVENT\r\nSUMMARY:Test\r\nEND:VEVENT\r\n"),
		"invalid start":       testCalendar(strings.Replace(testEvent, "20260907T090000", "invalid", 1) + "END:VEVENT\r\n"),
		"end before start":    testCalendar(strings.Replace(testEvent, "20260907T103500", "20260907T080000", 1) + "END:VEVENT\r\n"),
		"empty subject":       testCalendar(strings.Replace(testEvent, "SUMMARY:Робототехника (лекция)", "SUMMARY:", 1) + "END:VEVENT\r\n"),
		"invalid recurrence":  testCalendar(testEvent + "RRULE:FREQ=INVALID\r\nEND:VEVENT\r\n"),
		"invalid exclusion":   testCalendar(testEvent + "EXDATE:invalid\r\nEND:VEVENT\r\n"),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseCalendar(strings.NewReader(data), testDay(t, "2026-09-07")); err == nil {
				t.Fatal("expected a parsing error")
			}
		})
	}
}
