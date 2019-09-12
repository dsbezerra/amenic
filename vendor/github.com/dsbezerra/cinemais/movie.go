package cinemais

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/cinemais/util"
	"github.com/gocolly/colly"
)

const (
	// Selector of the container that contains all movie detail
	scDetail = "#filmeContainer"

	// Selector of the container that contains premiere movie list.
	scPremieres = "#indexContainer > div.estreiasContainer > div.estreiasContainerSide"
	scRemaining = "#indexContainer > div.continuacaoContainer > ul"

	// Selector of the container that contains upcoming movie list.
	scUpcoming = "#LancamentosContainer > div"

	// Selector of upcoming movies
	sUpcomingItem = "div.Poster"

	// Selector of upcoming movie title
	sUpcomingItemPage    = "a"
	sUpcomingItemTitle   = "h5"
	sUpcomingItemPoster  = "img"
	sUpcomingItemRelease = "small"
)

const (
	// RatingL is the Livre image path
	RatingL      = "ICO_LIV_programacao_GR.png"
	RatingLSmall = "ICO_LIV_programacao.png"
	// Rating10 is the age 10 image path
	Rating10      = "ICO_10_programacao_GR.png"
	Rating10Small = "ICO_10A_programacao.png"
	// Rating12 is the age 12 image path
	Rating12      = "ICO_12_programacao_GR.png"
	Rating12Small = "ICO_12A_programacao.png"
	// Rating14 is the age 14 image path
	Rating14      = "ICO_14_programacao_GR.png"
	Rating14Small = "ICO_14A_programacao.png"
	// Rating16 is the age 16 image path
	Rating16      = "ICO_16_programacao_GR.png"
	Rating16Small = "ICO_16A_programacao.png"
	// Rating18 is the age 18 image path
	Rating18      = "ICO_18_programacao_GR.png"
	Rating18Small = "ICO_18A_programacao.png"
)

const (
	// SchemeHTTP is the index used to access http poster urls
	SchemeHTTP = 0
	// SchemeHTTPS is the index used to access https poster urls
	SchemeHTTPS = 1
)

// PosterSize ...
type PosterSize uint

const (
	// PosterSizeSmall used to indicate claquete small poster image
	PosterSizeSmall PosterSize = iota
	// PosterSizeMedium used to indicate claquete medium poster image
	PosterSizeMedium
	// PosterSizeLarge used to indicate claquete large poster image
	PosterSizeLarge
	// PosterSizeCount used to know the count of sizes, must be the last
	PosterSizeCount
)

// Movie ...
type Movie struct {
	ID                  int                      `json:"id"`
	Title               string                   `json:"title,omitempty"`
	OriginalTitle       string                   `json:"original_title,omitempty"`
	Synopsis            string                   `json:"synopsis,omitempty"`
	Cast                []string                 `json:"cast,omitempty"`
	Screenplay          []string                 `json:"screenplay,omitempty"`
	ExecutiveProduction []string                 `json:"executive_production,omitempty"`
	Production          []string                 `json:"production,omitempty"`
	Direction           []string                 `json:"direction,omitempty"`
	PosterURLs          *[PosterSizeCount]string `json:"poster_urls,omitempty"`
	SecurePosterURLs    *[PosterSizeCount]string `json:"secure_poster_urls,omitempty"`
	Rating              int                      `json:"rating,omitempty"`
	Country             string                   `json:"country,omitempty"`
	Genres              []string                 `json:"genres,omitempty"`
	Runtime             uint                     `json:"runtime,omitempty"`
	ReleaseDate         *time.Time               `json:"release_date,omitempty"`
	Distributor         string                   `json:"distributor,omitempty"`
	DetailPage          string                   `json:"detail_url,omitempty"`
}

// Movies ...
type Movies []Movie

// GetMovie ...
func GetMovie(ID int) (*Movie, error) {
	return New().GetMovie(ID)
}

// GetMovie ...
func (c *Cinemais) GetMovie(ID int) (*Movie, error) {
	var result *Movie
	var err error

	url := fmt.Sprintf("%s/filmes/filme.php?cf=%d", c.BaseURL.String(), ID)

	c.collector.OnHTML(scDetail, func(e *colly.HTMLElement) {
		movie, parseErr := parseMovie(e.DOM)
		if parseErr != nil {
			err = parseErr
			return
		}

		getMoviePosterURLs(ID, movie)

		movie.ID = ID
		movie.DetailPage = url

		result = movie
	})
	c.collector.OnError(func(r *colly.Response, e error) {
		err = e
	})

	c.collector.Visit(url)
	return result, err
}

// GetPremieres ...
func GetPremieres() (Movies, error) {
	return New().GetPremieres()
}

// GetPremieres ...
func (c *Cinemais) GetPremieres() (Movies, error) {
	var result Movies
	var err error

	c.collector.OnHTML(scPremieres, func(e *colly.HTMLElement) {
		e.DOM.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
			movie, parseErr := parseNowPlayingMovie(s)
			if parseErr != nil {
				err = parseErr
				return false
			}
			result.AppendUniq(movie)
			return true
		})
	})

	c.collector.OnError(func(r *colly.Response, e error) {
		err = e
	})

	url := fmt.Sprintf("%s/programacao", c.BaseURL.String())
	c.collector.Visit(url)
	return result, err
}

// GetNowPlaying ...
func GetNowPlaying() (Movies, error) {
	return New().GetNowPlaying()
}

// GetNowPlaying ...
func (c *Cinemais) GetNowPlaying() (Movies, error) {
	var result Movies
	var err error

	c.collector.OnHTML(scPremieres, func(e *colly.HTMLElement) {
		e.DOM.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
			movie, parseErr := parseNowPlayingMovie(s)
			if parseErr != nil {
				err = parseErr
				return false
			}
			result.AppendUniq(movie)
			return true
		})
	})

	c.collector.OnHTML(scRemaining, func(e *colly.HTMLElement) {
		e.DOM.Find("li").EachWithBreak(func(i int, s *goquery.Selection) bool {
			movie, parseErr := parseNowPlayingMovie(s.Find("a"))
			if parseErr != nil {
				err = parseErr
				return false
			}
			result.AppendUniq(movie)
			return true
		})
	})

	c.collector.OnError(func(r *colly.Response, e error) {
		err = e
	})

	url := fmt.Sprintf("%s/programacao", c.BaseURL.String())
	c.collector.Visit(url)
	return result, err
}

// GetUpcoming ...
func GetUpcoming() (Movies, error) {
	return New().GetUpcoming()
}

// GetUpcoming ...
func (c *Cinemais) GetUpcoming() (Movies, error) {
	var result Movies
	var err error

	c.collector.OnHTML(scUpcoming, func(e *colly.HTMLElement) {
		e.DOM.Find(sUpcomingItem).EachWithBreak(func(i int, s *goquery.Selection) bool {
			movie, parseErr := parseUpcomingMovie(s)
			if parseErr != nil {
				err = parseErr
				return false
			}
			result.AppendUniq(movie)
			return true
		})
	})

	c.collector.OnError(func(r *colly.Response, e error) {
		err = e
	})

	url := fmt.Sprintf("%s/proximos_lancamentos", c.BaseURL.String())
	c.collector.Visit(url)
	return result, err
}

// IsEqual ...
func (m *Movie) IsEqual(movie Movie) bool {
	return m.ID == movie.ID
}

// AppendUniq ...
func (m *Movies) AppendUniq(movie *Movie) {
	exists := false

	for _, e := range *m {
		if e.IsEqual(*movie) {
			exists = true
			break
		}
	}

	if !exists {
		*m = append(*m, *movie)
	}
}

func parseMovie(s *goquery.Selection) (*Movie, error) {
	var result = &Movie{}

	title := util.GetTextTrimmed(s.Find("h1"))
	if title == "" {
		return nil, ErrUnexpectedStructure
	}
	result.Title = title

	// Get text between parentheses and last comma (Original Title, Year)
	originalTitle := util.GetTextTrimmed(s.Find("small"))
	if originalTitle == "" {
		return nil, ErrUnexpectedStructure
	}
	end := strings.LastIndex(originalTitle, ",")
	if end < 0 {
		return nil, ErrUnexpectedStructure
	}
	result.OriginalTitle = originalTitle[1:end]

	// Get synopsis, cast, screenwriters, producers, production, direction
	s.Find("#filmesContainer").Children().Each(func(i int, s *goquery.Selection) {
		if s.Is("h3") {
			label := strings.ToLower(util.GetTextTrimmed(s))
			content := util.GetTextTrimmed(s.Next())
			switch label {
			case "sinopse":
				result.Synopsis = content
			case "elenco", "roteiro", "produção executiva", "produção", "direção":
				if content != "" {
					content = strings.Replace(content, "Vozes de:", "", -1)
					content = strings.TrimFunc(content, func(r rune) bool {
						return unicode.IsSpace(r) || r == ','
					})
					content = strings.Replace(content, "\n", " ", -1)

					r := util.SplitTextTrimmed(content, ", ")
					if label == "elenco" {
						result.Cast = r
					} else if label == "roteiro" {
						result.Screenplay = r
					} else if label == "produção executiva" {
						result.ExecutiveProduction = r
					} else if label == "produção" {
						result.Production = r
					} else if label == "direção" {
						result.Direction = r
					}
				}
			}
		}
	})

	// Information rows
	rows := s.Find("#filmes_conteudo > table tr")
	rows.Each(func(i int, s *goquery.Selection) {
		lcolumns := rows.First().Find("td")
		ccolumns := rows.Last().Find("td")
		lcolumns.Each(func(i int, sel *goquery.Selection) {
			label := util.GetTextTrimmed(lcolumns.Eq(i))
			content := util.GetTextTrimmed(ccolumns.Eq(i - 1))

			switch label {
			case "País":
				result.Country = content
			case "Gênero":
				result.Genres = util.SplitTextTrimmed(content, ", ")
			case "Duração":
				content = strings.TrimFunc(content, func(r rune) bool {
					return !unicode.IsDigit(r)
				})
				runtime, err := strconv.Atoi(content)
				if err != nil {
					fmt.Printf("couldn't convert runtime %s to integer\n", content)
				} else {
					result.Runtime = uint(runtime)
				}
			case "Lançamento Nacional":
				if content != "" {
					parts := strings.Split(content, "/")
					if len(parts) == 3 {
						d, _ := strconv.Atoi(parts[0])
						m, _ := strconv.Atoi(parts[1])
						y, _ := strconv.Atoi(parts[2])

						loc, _ := time.LoadLocation("America/Sao_Paulo")
						date := time.Date(y, time.Month(m), d, 0, 0, 0, 0, loc)
						result.ReleaseDate = &date
					}
				}
			case "Distribuição":
				result.Distributor = content
			default:
				rating, err := parseRating(s)
				if err != nil {
					log.Printf("Error: %s", err.Error())
				} else {
					result.Rating = rating
				}
			}
		})
	})

	return result, nil
}

func parseNowPlayingMovie(s *goquery.Selection) (*Movie, error) {
	var result *Movie

	page := s.AttrOr("href", "")
	if page == "" {
		return nil, errors.New("couldn't find movie page url")
	}

	URL, err := url.Parse(page)
	if err != nil {
		return nil, err
	}

	ID, err := strconv.Atoi(URL.Query().Get("cf"))
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(s.AttrOr("title", ""))
	if title == "" {
		return nil, errors.New("couldn't find movie title")
	}

	result = &Movie{
		ID:         ID,
		Title:      title,
		DetailPage: page,
	}
	getMoviePosterURLs(ID, result)
	return result, nil
}

func parseUpcomingMovie(s *goquery.Selection) (*Movie, error) {
	var result *Movie

	URL, err := parseMovieURL(s.Find(sUpcomingItemPage))
	if err != nil {
		return nil, err
	}

	ID, err := strconv.Atoi(URL.Query().Get("cf"))
	if err != nil {
		return nil, err
	}

	title := util.GetTextTrimmed(s.Find(sUpcomingItemTitle))
	if title == "" {
		// NOTE: create a generic error to handle this
		return nil, errors.New("couldn't find movie title")
	}

	release := util.GetTextTrimmed(s.Find(sUpcomingItemRelease))
	date, err := parseUpcomingDate(release)
	if err != nil {
		return nil, err
	}

	result = &Movie{
		ID:          ID,
		Title:       title,
		ReleaseDate: date,
		DetailPage:  URL.String(),
	}
	getMoviePosterURLs(ID, result)
	return result, nil
}

func parseUpcomingDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, errors.New("invalid input")
	}

	parts := strings.Split(s, " de ")
	if len(parts) != 3 {
		return nil, errors.New("invalid input")
	}

	day, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}

	month, err := util.MonthTextToMonth(parts[1])
	if err != nil {
		return nil, err
	}

	year, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, err
	}

	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &t, nil
}

func parseMovieURL(s *goquery.Selection) (*url.URL, error) {
	page := s.AttrOr("href", "")
	if page == "" {
		return nil, errors.New("couldn't find movie page url")
	}

	URL, err := url.Parse(page)
	if err != nil {
		return nil, err
	}

	return URL, err
}

func parseRating(s *goquery.Selection) (int, error) {
	var result int
	var err error

	src, exists := s.Find("img").Attr("src")
	if exists {
		start := strings.LastIndex(src, "/")
		if start > -1 {
			switch src[start+1:] {
			case RatingL, RatingLSmall:
				result = -1
			case Rating10, Rating10Small:
				result = 10
			case Rating12, Rating12Small:
				result = 12
			case Rating14, Rating14Small:
				result = 14
			case Rating16, Rating16Small:
				result = 16
			case Rating18, Rating18Small:
				result = 18
			default:
				err = ErrUnexpectedStructure
			}
		} else {
			err = ErrUnexpectedStructure
		}
	} else {
		err = ErrUnexpectedStructure
	}

	return result, err
}

// NOTE: Cinemais content comes from Claquete that's why this
// function has claquete's host hardcoded.
//
// Claquete also supports HTTPS
// 07 feb. 2019
func buildPosterURLs(ID int) [2][PosterSizeCount]string {
	var result [2][PosterSizeCount]string

	host := "claquete.com"
	path := "fotos/filmes/poster"

	schemes := []string{"http", "https"}
	sizes := []string{"pequeno", "medio", "grande"}
	for _, scheme := range schemes {
		si := SchemeHTTP
		if scheme == "https" {
			si = SchemeHTTPS
		}

		for i, size := range sizes {
			url := fmt.Sprintf("%s://www.%s/%s/%d_%s.jpg", scheme, host, path, ID, size)
			result[si][i] = url
		}
	}

	return result
}

func getMoviePosterURLs(ID int, movie *Movie) {
	urls := buildPosterURLs(ID)
	movie.PosterURLs = &urls[SchemeHTTP]
	movie.SecurePosterURLs = &urls[SchemeHTTPS]
}

func ensureImageInClaquete(s *goquery.Selection) bool {
	poster := s.AttrOr("src", "")
	if poster == "" {
		return false
	}

	posterURL, err := url.Parse(poster)
	if err != nil {
		return false
	}

	return strings.Contains(posterURL.Host, "claquete")
}
