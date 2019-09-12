package util

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// MonthTextToMonth converts a month in pt-br text to time.Month
func MonthTextToMonth(s string) (time.Month, error) {
	var result time.Month

	lc := strings.ToLower(strings.TrimSpace(s))
	if len(lc) < 3 {
		return result, errors.New("invalid input")
	}

	start := lc[0:3]

	switch start {
	case "jan":
		result = time.January
	case "fev":
		result = time.February
	case "mar":
		result = time.March
	case "abr":
		result = time.April
	case "mai":
		result = time.May
	case "jun":
		result = time.June
	case "jul":
		result = time.July
	case "ago":
		result = time.August
	case "set":
		result = time.September
	case "out":
		result = time.October
	case "nov":
		result = time.November
	case "dez":
		result = time.December
	default:
		return result, fmt.Errorf("could't convert '%s' to month", lc)
	}

	return result, nil
}

// GetText retrieves the text contaning in the given selection.
func GetText(s *goquery.Selection) string {
	return s.Text()
}

// GetTextTrimmed is a GetText with text already trimmed.
func GetTextTrimmed(s *goquery.Selection) string {
	return strings.TrimSpace(GetText(s))
}

// SplitTextTrimmed ...
func SplitTextTrimmed(s, sep string) []string {
	splitted := strings.Split(s, sep)
	result := make([]string, len(splitted))
	for i, s := range splitted {
		result[i] = strings.TrimSpace(s)
	}
	return result
}

// ConsumeNextLine in a given reader
func ConsumeNextLine(reader *bufio.Reader) (string, bool) {
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return "", false
		}

		lineStr := string(line)
		if lineStr == "" {
			continue
		}

		return lineStr, true
	}
}

// ParseDay converts a day string to time.Weekday
func ParseDay(day string) (time.Weekday, bool) {
	day = strings.TrimSpace(day)
	if day == "" {
		return -1, false
	}

	var result time.Weekday
	day = strings.ToLower(day)
	if strings.HasPrefix(day, "dom") {
		result = time.Sunday
	} else if strings.HasPrefix(day, "seg") || strings.HasPrefix(day, "2ª") {
		result = time.Monday
	} else if strings.HasPrefix(day, "ter") || strings.HasPrefix(day, "3ª") {
		result = time.Tuesday
	} else if strings.HasPrefix(day, "qua") || strings.HasPrefix(day, "4ª") {
		result = time.Wednesday
	} else if strings.HasPrefix(day, "qui") || strings.HasPrefix(day, "5ª") {
		result = time.Thursday
	} else if strings.HasPrefix(day, "sex") || strings.HasPrefix(day, "6ª") {
		result = time.Friday
	} else if strings.HasPrefix(day, "sáb") || strings.HasPrefix(day, "sab") {
		result = time.Saturday
	} else {
		result = -1
	}

	return result, result >= time.Sunday && result <= time.Saturday
}
