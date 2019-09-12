package util

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// Contains checks if a given array of strings contains a given test string.
func Contains(values []string, test string) bool {
	if len(values) == 0 {
		return false
	}

	for _, s := range values {
		if s == test {
			return true
		}
	}

	return false
}

func ContainsAny(s string, strs []string) bool {
	if len(strs) == 0 {
		return false
	}

	for _, i := range strs {
		if strings.Contains(s, i) {
			return true
		}
	}

	return false
}

func remove(s string, old string, ignoreCase bool) string {
	if s == "" || old == "" {
		return s
	}

	w := s
	if ignoreCase {
		w = strings.ToLower(s)
		old = strings.ToLower(old)
	}

	if index := strings.Index(w, old); index > -1 {
		s = s[:index] + s[index+len(old):]
	}

	return s
}

func RemoveIgnoreCase(s, old string) string {
	return remove(s, old, true)
}

func Remove(s, old string) string {
	return remove(s, old, false)
}

// IsTimeString checks if a given string is in format HH:MM
func IsTimeString(s string) bool {
	if len(s) != 5 || s[2] != ':' {
		return false
	}

	hh := s[0:2]
	mm := s[3:]
	if !ContainsOnlyAlpha(hh) && !ContainsOnlyAlpha(mm) {
		return false
	}

	hours, _ := strconv.Atoi(hh)
	minutes, _ := strconv.Atoi(mm)
	if hours < 0 || hours > 24 || minutes < 0 || minutes > 59 {
		return false
	}

	return true
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

// ContainsOnlyAlpha checks if a given string is made only of characters [0-9]
func ContainsOnlyAlpha(s string) bool {
	result := true
	for i := range s {
		if !IsAlpha(s[i]) {
			return false
		}
	}
	return result
}

// IsCharacter checks if a given character is an character (letter)
func IsCharacter(character byte) bool {
	var result bool

	result = (character > 64 && character < 91 ||
		character > 96 && character < 123)
	return result
}

// IsUppercase checks if a given character is upper case or not
func IsUppercase(character byte) bool {
	var result bool
	result = character > 64 && character < 91
	return result
}

// IsAlpha checks if a given character is alphanumeric
func IsAlpha(character byte) bool {
	var result bool
	result = character > 47 && character < 58
	return result
}

// IsWhitespace checks if a given character is a whitespace
func IsWhitespace(character byte) bool {
	var result bool
	result = (character == ' ' ||
		character == '\n' ||
		character == '\r' ||
		character == '\v' ||
		character == '\f' ||
		character == '\t')
	return result
}

func EatSpaces(s string, i *int) {
	if *i >= len(s)-1 {
		return
	}

	for IsWhitespace(s[*i]) {
		*i++
	}
}

func EatWhitespaces(s string) string {
	for i := range s {
		if !IsWhitespace(s[i]) {
			return s[i:]
		}
	}
	return s
}

func PrintJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}
