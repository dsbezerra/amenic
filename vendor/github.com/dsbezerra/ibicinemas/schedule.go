package ibicinemas

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/ibicinemas/util"
	"github.com/gocolly/colly"
)

const (
	// VersionDubbed indicates the session is has dubbed audio
	VersionDubbed = "dubbed"
	// VersionSubtitled indicates the session has original audio with subtitles
	VersionSubtitled = "subtitled"
	// VersionNational indicates the session has original audio and the movie is national
	VersionNational = "national"
)

var (
	// DateExp is used to match dates in movie title (used to indicate preview/premiere)
	DateExp = regexp.MustCompile("\\d{2}\\/\\d{2}")
	// TimeExp is used to match time strings in session list
	TimeExp = regexp.MustCompile("\\d{1,}[:]\\d{1,}")
)

// Session represents a movie session
type Session struct {
	MovieTitle string     `json:"movie_title"`
	MovieID    string     `json:"movie_id"`
	Room       int        `json:"room"`
	Version    string     `json:"version"`
	Format     string     `json:"format"`
	StartTime  *time.Time `json:"start_time"`
}

// Schedule represents a playing schedule
type Schedule struct {
	// TODO(diego): Add period?
	Sessions Sessions `json:"sessions"`
}

// Sessions TODO
type Sessions []Session

// GetSchedule ...
func GetSchedule() (*Schedule, error) {
	return New().GetSchedule()
}

// GetSchedule ...
func (i *Ibicinemas) GetSchedule() (*Schedule, error) {
	movies, err := GetNowPlaying()
	if err != nil {
		return nil, err
	}

	if len(movies) == 0 {
		// TODO(diego): Custom error
		return nil, errors.New("movies not found")
	}

	type item struct {
		sessions Sessions
		err      error
	}

	ch := make(chan item, len(movies))
	for i := range movies {
		go func(m Movie) {
			var it item
			it.sessions, it.err = GetSessionsForMovie(m)
			ch <- it
		}(movies[i])
	}

	result := &Schedule{}
	for _, m := range movies {
		it := <-ch
		if it.err != nil {
			return nil, it.err
		}
		fmt.Printf("finished processing: %s\n", m.Title)
		result.Sessions = append(result.Sessions, it.sessions...)
	}

	return result, err
}

// GetSessionsForMovie ...
func GetSessionsForMovie(movie Movie) (Sessions, error) {
	_, err := url.Parse(movie.DetailPage)
	if err != nil {
		return nil, err
	}

	var result Sessions

	c := New().collector

	selector := "div.panel-body > div > div.col-sm-4.col-md-3"
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		e.DOM.Children().Each(func(i int, s *goquery.Selection) {
			if s.Is("h4") {
				var sessions Sessions
				sessions, err = parseSessions(s, movie)
				if err == nil {
					result = append(result, sessions...)
				}
			}
		})
	})

	c.OnError(func(r *colly.Response, e error) {
		err = e
	})

	c.Visit(movie.DetailPage)

	return result, err
}

func parseSessions(s *goquery.Selection, movie Movie) (Sessions, error) {
	var result Sessions
	var err error

	var room int
	var version string
	var format string

	{
		line := s.Text()
		var i int
		for {
			util.EatSpaces(line, &i)
			if i >= len(line) {
				break
			}

			b := line[i]
			if b == 'S' {
				if line[i:i+4] == "Sala" {
					i += 5 // NOTE(diego): Skipping space
					if util.IsAlpha(line[i]) {
						room, _ = strconv.Atoi(string(line[i]))
						i++
					}
				}
			} else if b == 'D' { // Dubbed
				version = VersionDubbed
				i += 7
			} else if b == 'L' { // Subtitled
				version = VersionSubtitled
				i += 9
			} else if b == 'N' { // National
				version = VersionNational
				i += 8
			} else if util.IsAlpha(b) {
				// Probably 2D or 3D labels
				if strings.Contains(line[i:], "2D") {
					format = "2D"
					i += 2
				} else if strings.Contains(line[i:], "3D") {
					format = "3D"
					i += 2
				}
			} else {
				i++
			}
		}
	}

	if room == 0 || format == "" || version == "" {
		return nil, errors.New("unexpected DOM structure")
	}

	playingRange := util.GetNowPlayingWeekAsRange(nil, false)
	weekStart := playingRange[0]

	session := &Session{
		MovieTitle: movie.Title,
		MovieID:    movie.ID,
		Room:       room,
		Format:     format,
		Version:    version,
	}

	s.Next().Find("li").Each(func(i int, s *goquery.Selection) {
		// Skip header with clock icon
		if i != 0 {
			// Prepare the text
			text := util.EatWhitespaces(s.Text())
			text = strings.ToLower(text)

			var singleDate time.Time
			var exceptDate time.Time
			{
				// NOTE(diego): We replace any dash with slashes because sometimes date
				// is in format dd-MM-yyyy and we need them in dd/MM/yyyy
				dateExpRes := DateExp.FindStringSubmatch(strings.Replace(text, "-", "/", -1))
				if len(dateExpRes) == 1 {
					var date time.Time
					date, err = util.CreateDateFromText(dateExpRes[0], "/", false, false)
					if err != nil {
						log.Print(err)
						return
					}
					if strings.Contains(text, "exceto") {
						exceptDate = date
					} else {
						singleDate = date
					}
				} else {
					log.Printf("couldn't find date in string: %s\n", text)
				}
			}

			// NOTE: only parse weekdays if we find opening times
			timeExpRes := TimeExp.FindAllString(text, -1)
			if timeExpRes == nil {
				log.Printf("couldn't find starting times in string: %s\n", text)
			} else {
				var datesToCreate []time.Time
				// Handle single date session
				if !singleDate.IsZero() {
					datesToCreate = append(datesToCreate, singleDate)
				} else if util.ContainsAny(text, []string{"estreia", "pré-estreia", "pre-estreia"}) {
					// Handle these cases:
					// ESTREIA DIA 'X' DE 'MES' (without quotes)
					// PRÉ-ESTREIA 'X' DE 'MES' (without quotes)
					log.Println("TODO")
				} else if strings.Contains(text, " a ") {
					// Handle this case:
					// QUINTA a DOMINGO
					days := strings.Split(text, " a ")
					if len(days) == 2 {
						start, startValid := util.ParseDay(days[0])
						end, endValid := util.ParseDay(days[1])
						if startValid && endValid {
							sidx := util.SafeDaysUntilNextWeekday(weekStart, start)
							eidx := util.SafeDaysUntilNextWeekday(weekStart, end)
							datesToCreate = playingRange[sidx : eidx+1] // Include eidx
						} else {
							log.Printf("couldn't find valid weekdays in string: %s\n", text)
						}
					} else {
						log.Printf("expected length equal 2 for string: %s\n", text)
					}
				} else if strings.Contains(text, ",") || strings.Contains(text, " e ") {
					// Handle these possible cases:
					// SEGUNDA, TERÇA, SÁBADO
					// SEGUNDA, TERÇA, SÁBADO e DOMINGO
					// SEGUNDA e DOMINGO
					text = strings.Replace(text, " e ", ", ", -1)
					days := strings.Split(text, ",")
					for _, day := range days {
						weekday, valid := util.ParseDay(day)
						if valid {
							num := util.SafeDaysUntilNextWeekday(weekStart, weekday)
							datesToCreate = append(datesToCreate, playingRange[num])
						} else {
							log.Printf("couldn't find valid weekday in string: %s\n", day)
						}
					}
				} else {
					// Single weekday/every day session
					day, valid := util.ParseDay(text)
					if valid {
						// Create session for this day only
						num := util.DaysUntilNextWeekday(weekStart, day)
						datesToCreate = append(datesToCreate, playingRange[num])
					} else {
						// Create session for all playing range
						datesToCreate = playingRange
					}
				}

				// Fill sessions
				for _, t := range timeExpRes {
					hours, minutes := util.ParseTime(t)
					for _, date := range datesToCreate {
						if !exceptDate.IsZero() && isSameDate(exceptDate, date) {
							continue
						}
						session.StartTime = util.SetSessionTime(date, hours, minutes)
						result = append(result, *session)
					}
				}
			}
		}
	})
	return result, err
}

func isSameDate(a time.Time, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
