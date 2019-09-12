package util

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// StringAdvance the starting index by the specified value
func StringAdvance(s *string, n int) {
	size := len(*s)
	if n < size {
		*s = (*s)[n:size]
	}
}

// CreateDateFromText creates a Time type from string date text
func CreateDateFromText(s string, delim string, hasYear, utc bool) (time.Time, error) {
	var result time.Time
	if s == "" || delim == "" {
		return result, errors.New("Date string and delimiter must be specified")
	}

	parts := strings.Split(s, delim)
	day, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])

	var year int
	if hasYear {
		year, _ = strconv.Atoi(parts[2])

		// @DumbHack
		if len(parts[2]) == 2 {
			year += 2000
		}

	} else {
		year = time.Now().Year()
	}

	loc, _ := time.LoadLocation("America/Sao_Paulo")
	result = time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
	if utc {
		result = result.UTC()
	}
	return result, nil
}

// EatSpaces removes all leading spaces
func EatSpaces(s string) string {
	index := 0
	size := len(s)
	for {
		if index >= size {
			break
		}

		if s[index] != ' ' {
			return s[index:]
		}

		index++
	}

	return ""
}

// BreakBySpaces shorthand for BreakByToken(string, ' ')
func BreakBySpaces(s string) (string, string) {
	return BreakByToken(s, ' ')
}

// BreakByToken breaks a string into two parts only if the specified token is found.
// Returns the left hand side of the string as first return value and the remainder
// of the string as the second. If token is not found then it returns the input
// string as first and an empty string as second.
func BreakByToken(s string, tok byte) (string, string) {
	s = EatSpaces(s)

	size := len(s)
	index := 0
	for {
		if index >= size {
			break
		}

		if s[index] == tok {
			return s[0:index], EatSpaces(s[index+1:])
		}

		index++
	}

	return s, ""
}
