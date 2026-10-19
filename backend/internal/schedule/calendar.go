// This file converts NSU calendar events and recurrence rules into daily lessons.
package schedule

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	_ "time/tzdata"

	ical "github.com/emersion/go-ical"
	"github.com/teambition/rrule-go"
)

// Location defines the timezone used for requested dates and returned lesson times.
// Embedded timezone data also supports containers without /usr/share/zoneinfo.
var Location = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Novosibirsk")
	if err != nil {
		panic(err)
	}
	return loc
}()

// parseCalendar decodes an ICS calendar and selects events starting on day.
// It applies recurrence rules, extra dates, exclusions, and cancellations, then
// returns lessons sorted by local start time. An empty result is a non-nil slice.
func parseCalendar(r io.Reader, day time.Time) ([]Lesson, error) {
	calendar, err := ical.NewDecoder(r).Decode()
	if err != nil {
		return nil, fmt.Errorf("decode NSU calendar: %w", err)
	}
	day = day.In(Location)
	from := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, Location)
	to := from.AddDate(0, 0, 1)
	lessons := make([]Lesson, 0)
	for _, event := range calendar.Events() {
		status, err := event.Status()
		if err != nil {
			return nil, err
		}
		if status == ical.EventCancelled {
			continue
		}
		start, err := event.DateTimeStart(Location)
		if err != nil || start.IsZero() {
			return nil, fmt.Errorf("invalid NSU event DTSTART: %v", err)
		}
		end, err := event.DateTimeEnd(Location)
		if err != nil || !end.After(start) {
			return nil, fmt.Errorf("invalid NSU event DTEND: %v", err)
		}
		subject, err := event.Props.Text(ical.PropSummary)
		if err != nil || strings.TrimSpace(subject) == "" {
			return nil, fmt.Errorf("invalid NSU event SUMMARY: %v", err)
		}
		room, err := event.Props.Text(ical.PropLocation)
		if err != nil {
			return nil, err
		}
		set := &rrule.Set{}
		set.DTStart(start)
		if prop := event.Props.Get(ical.PropRecurrenceRule); prop != nil {
			// NSU exports local UNTIL values without Z; retain the event timezone.
			options, err := rrule.StrToROptionInLocation(prop.Value, start.Location())
			if err != nil {
				return nil, fmt.Errorf("NSU RRULE: %w", err)
			}
			options.Dtstart = start
			rule, err := rrule.NewRRule(*options)
			if err != nil {
				return nil, err
			}
			set.RRule(rule)
		} else {
			set.RDate(start)
		}
		for _, name := range []string{ical.PropRecurrenceDates, ical.PropExceptionDates} {
			for _, prop := range event.Props.Values(name) {
				for _, value := range strings.Split(prop.Value, ",") {
					dateProp := prop
					dateProp.Value = value
					date, err := dateProp.DateTime(start.Location())
					if err != nil {
						return nil, err
					}
					if name == ical.PropExceptionDates {
						set.ExDate(date)
					} else {
						set.RDate(date)
					}
				}
			}
		}
		for _, occurrence := range set.Between(from, to, true) {
			if !occurrence.Before(to) {
				continue
			}
			lessons = append(lessons, Lesson{
				Subject:   strings.TrimSpace(subject),
				StartTime: occurrence.In(Location).Format("15:04"),
				EndTime:   occurrence.Add(end.Sub(start)).In(Location).Format("15:04"),
				Room:      strings.TrimSpace(strings.TrimPrefix(room, "Аудитория:")),
			})
		}
	}
	sort.SliceStable(lessons, func(i, j int) bool { return lessons[i].StartTime < lessons[j].StartTime })
	return lessons, nil
}
