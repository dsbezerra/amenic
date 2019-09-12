package cinemais

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/cinemais/util"
	"github.com/gocolly/colly"
)

const (
	// Selector of the container that contains schedule list.
	scSchedule = "#programacaoContainer > div.tableContainer > div"

	// Disclaimer stuff
	DisclaimerOnly   DisclaimerType = "only"
	DisclaimerExcept DisclaimerType = "except"

	// Format3D ...
	Format3D = "3D"
	// Format2D ...
	Format2D = "2D"

	// VersionSubtitled indicate showtime has original audio and subtitles
	VersionSubtitled = "subtitled"
	// VersionDubbed indicate showtimes has dubbed audio
	VersionDubbed = "dubbed"
	// VersionNational indicate showtime doesn't have either one of those
	VersionNational = "national"

	// Icon3D is the name of the 3D icon image
	Icon3D = "ICO_3d_programacao"
	// IconMagicD is the name of the MagicD icon image
	IconMagicD = "magic"
	// IconVIP is the name of the VIP icon image
	IconVIP = "vip"
)

var PeriodExp = regexp.MustCompile("\\((\\d{2}\\/\\d{2})\\)")

// Schedule ...
type Schedule struct {
	Sessions []Session `json:"sessions"`
}

// Session ...
type Session struct {
	Movie     *Movie     `json:"movie,omitempty"`
	Room      uint       `json:"room,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	Version   string     `json:"version,omitempty"`
	Format    string     `json:"format,omitempty"`
	MagicD    bool       `json:"magicd"`
	VIP       bool       `json:"vip"`
}

// Period ...
type Period struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

// DisclaimerType ...
type DisclaimerType string

// DisclaimerEntry ...
type DisclaimerEntry struct {
	Letter   string         `json:"letter,omitempty"`
	Type     DisclaimerType `json:"type,omitempty"`
	Dates    []Period       `json:"periods,omitempty"`
	Weekdays []time.Weekday `json:"weekdays,omitempty"`
	Text     string         `json:"text,omitempty"`
	// Time        string         `json:"time"`
}

// Disclaimer ...
type Disclaimer map[string]DisclaimerEntry

// Sessions ...
type Sessions []Session

// GetSchedule ...
func GetSchedule(ID string) (*Schedule, error) {
	return New(CinemaCode(ID)).GetSchedule()
}

// GetShowtimes ...
func (c *Cinemais) GetSchedule() (*Schedule, error) {
	var result Schedule
	var err error

	// NOTE: Improve this check!
	if c.CinemaCode == "" {
		return nil, ErrNoCinemaCode
	}

	c.collector.OnHTML(scSchedule, func(e *colly.HTMLElement) {
		sessions, parseErr := parseSessions(e.DOM)
		if parseErr != nil {
			err = parseErr
			return
		}
		result.Sessions = sessions
	})

	c.collector.OnError(func(r *colly.Response, e error) {
		err = e
	})

	c.collector.Visit(fmt.Sprintf("%s/programacao/cinema.php?cc=%s", c.BaseURL.String(), c.CinemaCode))

	return &result, err
}

func parseSessions(s *goquery.Selection) (Sessions, error) {
	d, err := parseDisclaimer(s.Find("div.disclaimer"))
	if err != nil {
		return nil, err
	}

	var result Sessions
	var nowPlayingWeek = getNowPlayingWeek(nil, false)
	s.Find("table > tbody > tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// Skip header
		if i == 0 {
			return true
		}

		session := Session{
			Format: Format2D,
		}

		// Get room information
		room, err := strconv.Atoi(util.GetTextTrimmed(s.Find("td:nth-child(1)")))
		if err != nil {
			return false
		}
		session.Room = uint(room)

		// Get movie information
		mc := s.Find("td:nth-child(2)")
		movie, err := parseSessionMovieInfo(mc)
		if err != nil {
			return false
		}

		rating, err := parseRating(s.Find("td:nth-child(3)"))
		if err != nil {
			return false
		}
		movie.Rating = rating

		session.Movie = movie

		// Get format of session (MagicD, 3D, VIP, etc)
		mc.Find("img").Each(func(i int, s *goquery.Selection) {
			title := strings.TrimSpace(s.AttrOr("title", ""))
			if title != "" {
				if title == "Em 3D" {
					session.Format = Format3D
				} else if title == "Magic D" {
					session.MagicD = true
				} else if title == "Vip" {
					session.VIP = true
				} else {
					// TODO: Logging
				}
			} else {
				src := s.AttrOr("src", "")
				c := strings.LastIndex(src, "/")
				if c > -1 && c+1 < len(src) {
					name := src[c+1:]
					if strings.Contains(name, Icon3D) {
						session.Format = Format3D
					} else if strings.Contains(name, IconMagicD) {
						session.MagicD = true
					} else if strings.Contains(name, IconVIP) {
						session.VIP = true
					} else {
						// TODO: Logging
					}
				}
			}
		})

		// Get starting times
		times := util.GetTextTrimmed(s.Find("td:nth-child(4)"))
		v, r := util.BreakByToken(times, '-')
		if v = strings.TrimSpace(v); v == "" {
			err = ErrUnexpectedStructure
		}
		if err != nil {
			return false
		}

		switch v {
		case "Dub.":
			session.Version = VersionDubbed
		case "Leg.":
			session.Version = VersionSubtitled
		default:
			if r == "" {
				r = v
				session.Version = VersionNational
			} else {
				err = ErrUnexpectedStructure
			}
		}

		if err != nil {
			return false
		}

		var t string
		for r != "" {
			t, r = util.BreakByToken(r, ',')

			var letter string
			var size = len(t)
			// If time has a letter as last character remove it and update time string
			if size > 0 && unicode.IsLetter(rune(t[size-1])) {
				letter = t[size-1:]
				t = t[0 : size-1]
			}

			h, m := util.BreakByToken(t, 'h')
			if h == "" || m == "" {
				err = ErrUnexpectedStructure
			}

			if err != nil {
				return false
			}

			hours, err := strconv.Atoi(h)
			if err != nil {
				return false
			}

			minutes, err := strconv.Atoi(m)
			if err != nil {
				return false
			}

			if letter != "" {
				entry, ok := d[letter]
				if ok {
					// Create showtimes accordingly with disclaimer
					if entry.Type == DisclaimerOnly {
						for _, date := range entry.Dates {
							y, m, d := date.Start.Date()
							st := time.Date(y, m, d, hours, minutes, 0, 0, date.Start.Location())
							session.StartTime = &st
							result.AppendUniq(&session)
						}
					} else if entry.Type == DisclaimerExcept {
						for d := 0; d < 7; d++ {
							nd := nowPlayingWeek.Start.AddDate(0, 0, d)

							add := true
							for _, date := range entry.Dates {
								// NOTE: Only day and month are enough here because we use
								// AddDate above which guarantees the correctness of time
								// when we deal with dates near New Year's Eve
								_, m, d := date.Start.Date()
								if nd.Day() == d && nd.Month() == m {
									add = false
									break
								}
							}

							if add {
								session.StartTime = setTime(&nd, hours, minutes)
								result.AppendUniq(&session)
							}
						}
					}
				} else {
					err = ErrUnexpectedStructure
				}
			} else {
				// Build sessions for the whole week
				for d := 0; d < 7; d++ {
					nd := nowPlayingWeek.Start.AddDate(0, 0, d)
					session.StartTime = setTime(&nd, hours, minutes)
					result.AppendUniq(&session)
				}
			}

			if err != nil {
				return false
			}
		}

		return true
	})

	return result, err
}

func parseSessionMovieInfo(s *goquery.Selection) (*Movie, error) {
	var result Movie
	var err error

	c := s.Find("a")
	result.Title = util.GetTextTrimmed(c)

	URL, err := parseMovieURL(c)
	if err != nil {
		return nil, err
	}
	ID, err := strconv.Atoi(URL.Query().Get("cf"))
	if err != nil {
		return nil, err
	}
	result.ID = ID

	if result.ID == 0 || result.Title == "" {
		return nil, ErrUnexpectedStructure
	}

	return &result, err
}

// This parses each entry of something like this:
//
//  A - Somente sábado (16/02)(16/02), domingo (17/02)(17/02)
//  B - Somente sábado (16/02)(16/02), domingo (17/02)(17/02), quarta (20/02)(20/02)
//
//  to a disclaimer map like map[string]DisclaimerEntry
func parseDisclaimer(s *goquery.Selection) (Disclaimer, error) {
	s.Find("br").ReplaceWithHtml("\n")

	d := make(Disclaimer)
	t := util.GetTextTrimmed(s)
	if t == "" {
		// NOTE: there's no disclaimer
		return d, nil
	}

	r := bufio.NewReader(strings.NewReader(t))
	for {
		line, found := util.ConsumeNextLine(r)
		if !found {
			break
		}

		// NOTE:
		// Found case:
		// A - Somente terça (19/02) - CINEMATERNA
		//
		// Should we handle CINEMATERNA?

		parts := strings.Split(line, " - ")
		count := len(parts)
		if count >= 2 {
			k := parts[0]
			v := parts[1]

			if len(k) == 1 && unicode.IsLetter(rune(k[0])) {
				item, err := parseDisclaimerEntry(v)
				if err != nil {
					return nil, err
				}
				item.Letter = k
				item.Text = line
				d[k] = *item
			}
		}
	}

	return d, nil
}

func parseDisclaimerEntry(s string) (*DisclaimerEntry, error) {
	scopy := s
	scopy = strings.ToLower(scopy)

	var err error
	var result DisclaimerEntry

	// Get the type
	t, r := util.BreakBySpaces(scopy)
	if t != "" {
		if t == "somente" {
			result.Type = DisclaimerOnly
		} else if t == "exceto" {
			result.Type = DisclaimerExcept
		} else {
			// TODO: Make an error for this
			err = errors.New("unknown value")
		}
	} else {
		err = ErrUnexpectedStructure
	}

	if err == nil {
		// Replace dots with space to handle cases like this:
		// Somente qua.(15/05)
		// Somente quarta (15/05)
		r = strings.Replace(r, ".", " ", -1)
		r = util.EatSpaces(r)

		var day, date string
		// Get the day(s) now
		for r != "" {
			day, r = util.BreakBySpaces(r)
			if day == "" {
				err = ErrUnexpectedStructure
				break
			}

			// Get the day number
			weekday, ok := util.ParseDay(day)
			if !ok {
				err = ErrUnexpectedStructure
				break
			}

			result.Weekdays = append(result.Weekdays, weekday)
			if r == "" {
				err = ErrUnexpectedStructure
				break
			}

			date, r = util.BreakBySpaces(r)
			if date == "" {
				err = ErrUnexpectedStructure
				break
			}

			// Parse period date
			//
			// NOTE: Year is ignored here.
			//
			// ex: If we have a now playing week starting in 2019 and ending
			// in 2020 all dates will be created in the year the code executed.
			//
			// This works fine because we don't need to check years when we
			// look for entries in the disclaimer map.
			res := PeriodExp.FindStringSubmatch(date)
			length := len(res)
			if length == 3 { // Example text (DD/MM)(DD/MM)
				start, err := util.CreateDateFromText(res[1], "/", false, false)
				if err != nil {
					break
				}

				end, err := util.CreateDateFromText(res[2], "/", false, false)
				if err != nil {
					break
				}

				result.Dates = append(result.Dates, Period{
					Start: &start,
					End:   &end,
				})
			} else if length == 2 { // Example text (DD/MM)
				single, err := util.CreateDateFromText(res[1], "/", false, false)
				if err != nil {
					break
				}
				result.Dates = append(result.Dates, Period{
					Start: &single,
					End:   &single,
				})
			} else if length == 1 {
				lhs, r := util.BreakByToken(date, '/')
				if lhs != "" {
					if lhs[0] == '(' {
						util.StringAdvance(&lhs, 1)
					}

					start, err := util.CreateDateFromText(lhs, "/", false, false)
					if err != nil {
						break
					}

					lhs, _ := util.BreakByToken(r, ')')
					if lhs != "" {
						end, err := util.CreateDateFromText(lhs, "/", false, false)
						if err != nil {
							break
						}

						result.Dates = append(result.Dates, Period{
							Start: &start,
							End:   &end,
						})
					}
				} else {
					// TODO(diego): Proper error logging.
					fmt.Printf("Failed to create date from text '%s'\n", date)
				}
			} else {
				// TODO(diego): Handle cases without dates
			}
		}
	}

	return &result, err
}

// IsEqual ...
func (s *Session) IsEqual(session Session) bool {
	return s.Format == session.Format &&
		s.MagicD == session.MagicD &&
		s.Movie == session.Movie &&
		s.Room == session.Room &&
		s.StartTime == session.StartTime &&
		s.VIP == session.VIP &&
		s.Version == session.Version
}

// AppendUniq ...
func (s *Sessions) AppendUniq(session *Session) {
	exists := false

	for _, e := range *s {
		if e.IsEqual(*session) {
			exists = true
			break
		}
	}

	if !exists {
		*s = append(*s, *session)
	}
}

// returns the now playing week for the given time
func getNowPlayingWeek(t *time.Time, utc bool) Period {
	if t == nil {
		loc, _ := time.LoadLocation("America/Sao_Paulo")
		now := time.Now().In(loc)
		t = &now
	}
	count := daysUntilNextWednesday(t)
	y, m, d, loc := t.Year(), t.Month(), t.Day(), t.Location()
	s := time.Date(y, m, d, 0, 0, 0, 0, loc).AddDate(0, 0, -(6 - count))
	e := time.Date(y, m, d, 0, 0, 0, 0, loc).AddDate(0, 0, count)
	if utc {
		s = s.UTC()
		e = e.UTC()
	}
	return Period{Start: &s, End: &e}
}

// daysUntilNextWednesday calculates how many days we are to next wednesday
func daysUntilNextWednesday(now *time.Time) int {
	result := -1

	w := int(now.Weekday())
	if w < 4 {
		result = (4 - 1) - w
	} else {
		result = (4 + 7 - 1) - w
	}

	return result
}

func setTime(t *time.Time, hours, minutes int) *time.Time {
	result := *t
	result = result.
		Add(time.Duration(hours) * time.Hour).
		Add(time.Duration(minutes) * time.Minute)
	return &result
}
