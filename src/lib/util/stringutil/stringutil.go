package stringutil

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/sajari/fuzzy"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Advance the starting index by the specified value
func Advance(s *string, n int) {
	size := len(*s)
	if n < size {
		*s = (*s)[n:size]
	}
}

// IsOnlyAlpha checks if a given string contains only alpha characters.
func IsOnlyAlpha(str string) bool {
	size := len(str)
	if size == 0 {
		return false
	}

	i := 0
	for i < size {
		if !IsAlpha(str[i]) {
			return false
		}
		i++
	}

	return true
}

// IsUppercase checks if a given character is upper case or not
func IsUppercase(character byte) bool {
	var result bool
	result = character > 64 && character < 91
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

// IsAlpha checks if a given character is alphanumeric
func IsAlpha(character byte) bool {
	var result bool
	result = character > 47 && character < 58
	return result
}

// IsCharacter checks if a given character is an character (letter)
func IsCharacter(character byte) bool {
	var result bool

	result = (character > 64 && character < 91 ||
		character > 96 && character < 123)
	return result
}

// EatSpaces removes all leading zeroes
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

	return s
}

// EatSpacesWithIndex removes all leading zeroes
func EatSpacesWithIndex(s string, index *int) {
	size := len(s)
	for {
		if *index >= size {
			break
		}

		if s[*index] != ' ' {
			break
		}

		*index++
	}
}

// BreakBySpaces shorthand for BreakByToken(string, ' ')
func BreakBySpaces(s string) (string, string) {
	return BreakByToken(s, ' ')
}

// BreakByToken breaks a string into two parts only if the specified token
// was found. Returns the left hand side of the string as first return value
// and the remainder of the string as the second.
// If token wansn't found then returns the input string as first
// and empty string as second.
func BreakByToken(s string, tok byte) (string, string) {
	s = strings.TrimSpace(s)
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

// StringAfter gets the substring after a given character is found
func StringAfter(s string, tok byte) string {
	return stringAfter(s, tok, false)
}

// StringAfterLast gets the substring after the last occurence of the given
// character
func StringAfterLast(s string, tok byte) string {
	return stringAfter(s, tok, true)
}

func stringAfter(s string, tok byte, last bool) string {
	size := len(s)
	if size < 2 {
		return ""
	}

	index := 0
	result := ""
	for {
		if index >= size {
			break
		}

		if s[index] == tok {
			if index+1 >= size {
				result = ""
				break
			}

			result = s[index+1:]
			if !last {
				break
			}
		}

		index++
	}

	return result
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

// RemoveIgnoreCase remove first occurence of old in s ignore the case
func RemoveIgnoreCase(s, old string) string {
	return remove(s, old, true)
}

// Remove remove first occurence of old in s
func Remove(s, old string) string {
	return remove(s, old, false)
}

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

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

// OnlyAlphaLetters remove anything from a given string that is not an alpha or letter character
func OnlyAlphaLetters(str string) (string, bool) {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	str, _, _ = transform.String(t, str)

	size := len(str)
	if size == 0 {
		return "", false
	}

	b := strings.Builder{}
	valid := true

	index := 0
	for index < size {
		c := str[index]
		index++

		if IsAlpha(c) || IsCharacter(c) || c == ':' || c == '-' {
			if IsUppercase(c) {
				c += 32
			}
			b.WriteByte(c)
		} else if valid {
			valid = false
		}
	}

	return b.String(), valid
}

// ContainsWithSpellCheck checks if a given string contain other string considering
// typos
func ContainsWithSpellCheck(str string, test string, model *fuzzy.Model) bool {
	if model == nil || str == "" || test == "" {
		return false
	}

	for _, w := range strings.Split(str, " ") {
		normalized, _ := OnlyAlphaLetters(w)
		result := model.SpellCheck(normalized)
		if result != "" {
			return true
		}
	}

	return false
}

// EatUntilAlpha eats everything until an alphanumeric is found
func EatUntilAlpha(s string, position *int) int {
	if *position >= len(s)-1 {
		return *position
	}

	for !IsAlpha(s[*position]) {
		*position++

		if *position >= len(s)-1 {
			break
		}
	}

	return *position
}

// EatUntilLastAlpha eats everything until the last alphanumberic returning
// the index of this alpha
func EatUntilLastAlpha(s string, position *int) int {
	result := 0

	for *position < len(s) {
		if IsAlpha(s[*position]) {
			result = *position
		}
		*position++
	}

	return result
}

// EatUntilToken eats everything until the specified token is found
func EatUntilToken(s string, tok byte, position *int) int {
	for s[*position] != tok {
		*position++

		if *position >= len(s)-1 {
			break
		}
	}
	return *position
}

// CreateDateFromText creates a Time type from string date text
func CreateDateFromText(s string, delim string, hasYear bool) (time.Time, error) {
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
	result = time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc).UTC()
	return result, nil
}

// ConsumeNextLine ...
// TODO: We should remove everything that uses this and make it use our text_file_handler.go
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
