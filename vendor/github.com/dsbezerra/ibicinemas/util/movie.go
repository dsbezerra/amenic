package util

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// Date in dd/MM/YYYY
	reReleaseDate = regexp.MustCompile("\\d{2}/\\d{2}/\\d{4}")
	// Alternative date
	reAltReleaseDate = regexp.MustCompile("(\\d{1,}/\\d{1,}/\\d{2,})")
	// Alternative date 2
	reAlt2ReleaseDate = regexp.MustCompile("(\\d{1,}-\\d{1,}-\\d{2,})")
	// Preview showtime date, matches 'DIA 03 DE ABRIL' in movie title for IBICINEMAS
	reShowtimeDateExp = regexp.MustCompile("(?i)(dia*\\s*(\\d{1,})\\sde*\\s*(janeiro|fevereiro|março|abril|maio|junho|julho|agosto|setembro|outubro|novembro|dezembro))")
	// Date period in some movie titles for IBICINEMAS
	reShowtimePeriodInTitle = regexp.MustCompile("(?i)(,*\\s*(?:do)*\\sdia\\s(\\d{2}\\/\\d{2})\\sao*\\s(\\d{2}\\/\\d{2}))")
)

// TrimExtraTitleText removes any extra text that is not part of a movie title like
// preview/release dates
func TrimExtraTitleText(title string) string {
	result := title

	// Remove any dates
	if res := reReleaseDate.FindStringSubmatch(result); len(res) > 1 {
		result = strings.Replace(result, res[1], "", -1)
	} else if res := reAltReleaseDate.FindStringSubmatch(result); len(res) > 1 {
		result = strings.Replace(result, res[1], "", -1)
	} else if res := reAlt2ReleaseDate.FindStringSubmatch(result); len(res) > 1 {
		result = strings.Replace(result, res[1], "", -1)
	} else if res := reShowtimePeriodInTitle.FindStringSubmatch(result); len(res) > 1 {
		result = strings.Replace(result, res[1], "", -1)
	}

	if ContainsAny(strings.ToLower(result), []string{"estreia", "pré-estreia", "pre-estreia", "pré-venda", "pre-venda"}) {
		if res := reShowtimeDateExp.FindStringSubmatch(result); len(res) > 1 {
			result = strings.Replace(result, res[1], "", -1)
		}
	}

	// Remove premiere texts
	result = RemoveIgnoreCase(result, "pré-estreia")
	result = RemoveIgnoreCase(result, "pre-estreia")
	result = RemoveIgnoreCase(result, "pré-venda")
	result = RemoveIgnoreCase(result, "pre-venda")

	// Trim any possible dashes and spaces
	result = strings.TrimFunc(result, func(r rune) bool {
		return r == ',' || r == '-' || unicode.IsSpace(r)
	})

	return result
}

// GenerateMovieSlug ...
func GenerateMovieSlug(str string) string {
	size := len(str)
	if size == 0 {
		return ""
	}
	start := 0
	end := size - 1
	var b strings.Builder
	// Skip left spaces
	for {
		if !IsWhitespace(str[start]) {
			break
		}
		start++
	}

	// Skip right spaces
	for {
		if !IsWhitespace(str[end]) {
			break
		}
		end--
	}

	i := start
	for i <= end {
		c := str[i]
		i++
		if IsWhitespace(c) {
			continue
		}
		// We add only if it is a-z0-9
		if IsAlpha(c) || IsCharacter(c) {
			if IsUppercase(c) {
				c += 32
			}
			b.WriteByte(c)
		}
	}
	return b.String()
}

var prepList = []string{
	"e",
	"o", "os",
	"a", "as", "à", "às",
	"um", "uns", "uma", "umas",
	"de", "do", "dos", "da", "das", "dum", "duns", "duma", "dumas",
	"em", "no", "nos", "na", "nas", "num", "nuns", "numa", "numas",
	"por", "pelo", "pelos", "pela", "pelas",
}

// CapTitle TODO
func CapTitle(str string) string {
	if str == "" {
		return str
	}

	result := strings.Builder{}
	words := strings.Split(str, " ")
	size := len(words)
	for i, w := range words {
		w = strings.ToLower(w)

		cap := true
		if Contains(prepList, w) {
			if i > 0 {
				prev := words[i-1]
				if !strings.HasSuffix(prev, ":") && !strings.HasSuffix(prev, "-") {
					cap = false
				}
			}
		}
		if cap {
			w = strings.Title(w)
		}
		result.WriteString(w)
		if i < size-1 {
			result.WriteString(" ")
		}
	}
	return result.String()
}

func FixTitle(input string) string {
	return CapTitle(TrimExtraTitleText(input))
}
