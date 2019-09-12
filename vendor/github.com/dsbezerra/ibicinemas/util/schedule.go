package util

import (
	"time"
)

// GetNowPlayingWeek returns the now playing week for the given time
func GetNowPlayingWeek(t *time.Time, UTC bool) (*time.Time, *time.Time) {
	if t == nil {
		loc, _ := time.LoadLocation("America/Sao_Paulo")
		now := time.Now().In(loc)
		t = &now
	}
	count := DaysUntilNextWednesday(*t)
	y, m, d, loc := t.Year(), t.Month(), t.Day(), t.Location()
	s := time.Date(y, m, d, 0, 0, 0, 0, loc).AddDate(0, 0, -(6 - count))
	e := time.Date(y, m, d, 0, 0, 0, 0, loc).AddDate(0, 0, count)
	if UTC {
		s = s.UTC()
		e = e.UTC()
	}
	return &s, &e
}

// GetNowPlayingWeekAsRange ...
func GetNowPlayingWeekAsRange(t *time.Time, UTC bool) []time.Time {
	result := make([]time.Time, 7)

	start, _ := GetNowPlayingWeek(t, UTC)
	for i := range result {
		result[i] = start.AddDate(0, 0, i)
	}

	return result
}

// SetSessionTime ...
func SetSessionTime(t time.Time, hours, minutes int) *time.Time {
	result := t.
		Add(time.Duration(hours) * time.Hour).
		Add(time.Duration(minutes) * time.Minute)
	return &result
}

// DaysUntilNextWednesday calculates how many days to next wednesday
func DaysUntilNextWednesday(now time.Time) int {
	return DaysUntilNextWeekday(now, time.Wednesday)
}

// DaysUntilNextWeekday ...
func DaysUntilNextWeekday(now time.Time, weekday time.Weekday) int {
	result := -1
	nwi := int(now.Weekday())
	wi := int(weekday)
	if nwi < wi {
		result = wi - nwi
	} else {
		result = (wi + 7) - nwi
	}
	return result
}

// SafeDaysUntilNextWeekday ...
func SafeDaysUntilNextWeekday(now time.Time, weekday time.Weekday) int {
	r := DaysUntilNextWeekday(now, weekday)
	if r == 7 {
		r = 0
	}
	return r
}
