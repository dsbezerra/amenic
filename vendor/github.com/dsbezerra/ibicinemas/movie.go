package ibicinemas

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/ibicinemas/util"
	"github.com/gocolly/colly"
)

const (
	RatingL  = "images/censura/ibicinemas-censura-livre-1.jpg"
	Rating10 = "images/censura/ibicinemas-censura-10-anos-2.jpg"
	Rating12 = "images/censura/ibicinemas-censura-12-anos-3.jpg"
	Rating14 = "images/censura/ibicinemas-censura-14-anos-4.jpg"
	Rating16 = "images/censura/ibicinemas-censura-16-anos-5.jpg"
	Rating18 = "images/censura/ibicinemas-censura-18-anos-6.jpg"
)

// Movie TODO
type Movie struct {
	// NOTE(diego): This is not really a ID.
	ID          string   `json:"id,omitempty"`
	Title       string   `json:"title,omitempty"`
	Poster      string   `json:"poster_url,omitempty"`
	Synopsis    string   `json:"synopsis,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	Runtime     uint     `json:"runtime,omitempty"`
	Rating      int      `json:"rating,omitempty"`
	Distributor string   `json:"distributor,omitempty"`
	DetailPage  string   `json:"detail_url,omitempty"`
	Trailer     *Trailer `json:"trailer,omitempty"`
}

// Trailer TODO
type Trailer struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

const (
	scMovie    = "body > div:nth-child(4) > div > div"
	scPlaying  = "body > div:nth-child(6) > div > div.panel.panel-default > div > div"
	scUpcoming = ".proxfilm"
)

// Movies ...
type Movies []Movie

// GetMovie ...
func (i *Ibicinemas) GetMovie(id string) (*Movie, error) {
	return i.getMovie(id)
}

// GetNowPlaying ...
func (i *Ibicinemas) GetNowPlaying() (Movies, error) {
	return i.getSliderMovies(scPlaying, "")
}

// GetUpcoming ...
func (i *Ibicinemas) GetUpcoming() (Movies, error) {
	return i.getSliderMovies(scUpcoming, "")
}

// GetMovie ...
func GetMovie(id string) (*Movie, error) {
	return New().getMovie(id)
}

// GetNowPlaying ...
func GetNowPlaying() (Movies, error) {
	return New().getSliderMovies(scPlaying, "")
}

// GetUpcoming ...
func GetUpcoming() (Movies, error) {
	return New().getSliderMovies(scUpcoming, "ibicinemas-proximos-lancamentos-7.html")
}

func (i *Ibicinemas) getMovie(id string) (*Movie, error) {
	var result *Movie
	var err error
	url := fmt.Sprintf("%s/%s.html", baseURL, id)
	i.collector.OnHTML(scMovie, func(e *colly.HTMLElement) {
		result, err = parseMovie(e.DOM)
		if err == nil {
			result.DetailPage = url
		}
	})
	i.collector.OnError(func(r *colly.Response, e error) {
		err = e
	})
	visitErr := i.collector.Visit(url)
	if visitErr != nil {
		err = visitErr
	}
	return result, err
}

func (i *Ibicinemas) getSliderMovies(container string, path string) (Movies, error) {
	var result Movies
	var err error
	i.collector.OnHTML(container, func(e *colly.HTMLElement) {
		result, err = parseSliderMovies(e.DOM)
	})
	i.collector.OnError(func(r *colly.Response, e error) {
		err = e
	})
	url := fmt.Sprintf("%s/%s", baseURL, path)
	visitErr := i.collector.Visit(url)
	if visitErr != nil {
		err = visitErr
	}
	return result, err
}

func parseMovie(s *goquery.Selection) (*Movie, error) {
	var result *Movie
	var err error

	title := util.FixTitle(s.Find("h2").Text())
	if title == "" {
		// TODO: Better error handling
		return nil, errors.New("couldn't find movie title")
	}

	result = &Movie{
		ID:       util.GenerateMovieSlug(title),
		Title:    title,
		Synopsis: s.Find("div.video-description > p").Text(),
	}

	{
		// Get side bar info
		selector := "div.panel-body > div > div.col-sm-4.col-md-3"
		s.Find(selector).Children().Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				result.Poster = getPoster(s)
			} else {
				if result.Rating == 0 {
					result.Rating = getRating(s)
				}
				if s.Is("ul") {
					s.Find("li").Each(func(i int, s *goquery.Selection) {
						content := strings.TrimSpace(s.Text())
						if content != "" {
							icon := s.Find("i")
							if icon.HasClass("fa-check-square-o") {
								result.Genres = append(result.Genres, content)
							} else if icon.HasClass("fa-clock-o") {
								content = strings.TrimFunc(content, func(r rune) bool {
									return !unicode.IsDigit(r)
								})
								value, err := strconv.Atoi(content)
								if err == nil {
									result.Runtime = uint(value)
								} else {
									// TODO: Logging
								}
							} else if icon.HasClass("fa-university") {
								result.Distributor = content
							}
						}
					})
				} else if s.Is("h4") {
					// TODO: Include showtime information?
				}
			}
		})
	}

	{
		// Get a trailer
		if url := s.Find("div.video-container > iframe").AttrOr("src", ""); url != "" {
			var start, end int
			if start = strings.Index(url, "embed/"); start > -1 {
				start += len("embed/")
			}
			end = strings.Index(url, "?")
			if start < end {
				ID := url[start:end]
				result.Trailer = &Trailer{
					ID:  ID,
					URL: fmt.Sprintf("https://www.youtube.com/watch?v=%s", ID),
				}
			}
		}
	}

	return result, err
}

func parseSliderMovies(s *goquery.Selection) (Movies, error) {
	var result Movies
	var err error
	s.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if page := s.AttrOr("href", ""); page != "" {
			URL, err := util.ResolveRelativeURL(baseURL, page)
			if err != nil {
				// TODO: Logging
				return false
			}
			// Get "ID"
			title := util.FixTitle(s.Find(".film-thumb-info").Text())
			result = append(result, Movie{
				ID:         util.GenerateMovieSlug(title),
				Title:      title,
				Poster:     getPoster(s),
				DetailPage: URL,
			})
			return true
		}
		return false
	})
	return result, err
}

func getPoster(s *goquery.Selection) string {
	var result string
	url, err := util.ResolveRelativeURL(baseURL, s.Find("img").AttrOr("src", ""))
	if err == nil {
		result = url
	} else {
		// TODO: Logging
	}
	return result
}

func getRating(s *goquery.Selection) int {
	var result int

	s.Find("img").EachWithBreak(func(i int, s *goquery.Selection) bool {
		switch s.AttrOr("src", "") {
		case RatingL:
			result = -1
		case Rating10:
			result = 10
		case Rating12:
			result = 12
		case Rating14:
			result = 14
		case Rating16:
			result = 16
		case Rating18:
			result = 18
		default:
			return true
		}
		return false
	})

	return result
}
