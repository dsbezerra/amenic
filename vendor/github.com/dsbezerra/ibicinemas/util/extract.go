package util

import (
	"strconv"
	"strings"
	"time"
)

// ParseDay ...
func ParseDay(day string) (time.Weekday, bool) {
	var result time.Weekday
	day = strings.ToLower(day)
	day = strings.TrimSpace(day)
	if strings.HasPrefix(day, "dom") {
		result = time.Sunday
	} else if strings.HasPrefix(day, "seg") {
		result = time.Monday
	} else if strings.HasPrefix(day, "ter") {
		result = time.Tuesday
	} else if strings.HasPrefix(day, "qua") {
		result = time.Wednesday
	} else if strings.HasPrefix(day, "qui") {
		result = time.Thursday
	} else if strings.HasPrefix(day, "sex") {
		result = time.Friday
	} else if strings.HasPrefix(day, "sáb") {
		result = time.Saturday
	} else {
		result = -1
	}
	return result, result >= time.Sunday && result <= time.Saturday
}

// ParseTime extract hours and minutes from a time string in
// format HH:MM and return hours as first and minutes as seconds.
func ParseTime(text string) (hours int, minutes int) {
	if len(text) == 5 {
		hours, _ = strconv.Atoi(text[0:2])
		minutes, _ = strconv.Atoi(text[3:])
	}
	return hours, minutes
}
