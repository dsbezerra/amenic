package models

import (
	"strings"
	"time"
)

// Weekday represents a weekday or special day used in Showtime and
// Price models to indicate which days their data apply
type Weekday uint

const (
	// ALL means all days, including any holiday and premiere ones
	ALL Weekday = iota
	// SUNDAY weekday
	SUNDAY
	// MONDAY weekday
	MONDAY
	// TUESDAY weekday
	TUESDAY
	// WEDNESDAY weekday
	WEDNESDAY
	// THURSDAY weekday
	THURSDAY
	// FRIDAY weekday
	FRIDAY
	// SATURDAY weekday
	SATURDAY
	// HOLIDAY weekday
	HOLIDAY
	// PREMIERE premiere days (mostly thursdays)
	PREMIERE
	// INVALID This is used as none of above.
	INVALID
)

// NameToWeekday ...
func NameToWeekday(name string) Weekday {
	name = strings.ToLower(name)
	switch name {
	case "domingo", "sunday":
		return SUNDAY
	case "segunda", "monday":
		return MONDAY
	case "terça", "terca", "tuesday":
		return TUESDAY
	case "quarta", "wednesday":
		return WEDNESDAY
	case "quinta", "thursday":
		return THURSDAY
	case "sexta", "friday":
		return FRIDAY
	case "sábado", "sabado", "saturday":
		return SATURDAY
	case "feriado", "holiday":
		return HOLIDAY
	case "pré-estreia", "pre-estreia", "preview":
		return PREMIERE
	default:
		return INVALID
	}
}

// NameToTimeWeekday ...
func NameToTimeWeekday(name string) time.Weekday {
	name = strings.ToLower(name)
	switch name {
	case "domingo", "sunday":
		return time.Sunday
	case "segunda", "monday":
		return time.Monday
	case "terça", "terca", "tuesday":
		return time.Tuesday
	case "quarta", "wednesday":
		return time.Wednesday
	case "quinta", "thursday":
		return time.Thursday
	case "sexta", "friday":
		return time.Friday
	case "sábado", "sabado", "saturday":
		return time.Saturday
	}

	return -1
}

// TimeWeekdayToWeekday ...
func TimeWeekdayToWeekday(w time.Weekday) Weekday {

	switch w {
	case time.Sunday:
		return SUNDAY
	case time.Monday:
		return MONDAY
	case time.Tuesday:
		return TUESDAY
	case time.Wednesday:
		return WEDNESDAY
	case time.Thursday:
		return THURSDAY
	case time.Friday:
		return FRIDAY
	case time.Saturday:
		return SATURDAY
	}

	return INVALID
}
