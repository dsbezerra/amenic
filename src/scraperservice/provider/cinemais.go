package provider

import (
	"fmt"
	"sync"
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/movieutil"
	"github.com/dsbezerra/cinemais"
	"github.com/sirupsen/logrus"
)

type (
	// ComplexCode ...
	ComplexCode string

	// Complex represents a Cinemais Complex.
	Complex struct {
		Code ComplexCode
		Name string
		City string
		UF   string
	}

	// Cinemais ...
	Cinemais struct {
		t       *models.Theater
		complex Complex
		log     *logrus.Entry
	}
)

// CinemaisComplexes is a list of supported complexes
var CinemaisComplexes = map[ComplexCode]Complex{
	"34": Complex{Code: "34", Name: "Montes Claros", City: "Montes Claros", UF: "MG"},
}

// NewCinemais ...
func NewCinemais(cc ComplexCode) *Cinemais {
	complex, ok := CinemaisComplexes[cc]
	if !ok {
		// Defaulting to Montes Claros
		complex = CinemaisComplexes["34"]
	}

	return &Cinemais{
		complex: complex,
		log:     logrus.WithField("provider", fmt.Sprintf("cinemais-%s", complex.Code)),
	}
}

// Init ...
func (c *Cinemais) Init(data persistence.DataAccessLayer) error {
	query := data.DefaultQuery().
		AddCondition("internalId", c.complex.Code).
		AddCondition("shortName", "Cinemais").
		AddInclude("city")
	theater, err := data.FindTheater(query)
	if err != nil {
		return err
	}
	c.t = theater
	return nil
}

// GetNowPlaying ...
func (c *Cinemais) GetNowPlaying() ([]models.Movie, error) {
	movies, err := cinemais.GetNowPlaying()
	return c.fillMoviesDetailsAndMap(movies), err
}

// GetUpcoming ...
func (c *Cinemais) GetUpcoming() ([]models.Movie, error) {
	movies, err := cinemais.GetUpcoming()
	return c.fillMoviesDetailsAndMap(movies), err
}

// GetSchedule ...
func (c *Cinemais) GetSchedule() ([]models.Session, error) {
	schedule, err := cinemais.GetSchedule(string(c.complex.Code))
	return c.mapSessions(schedule.Sessions), err
}

// GetPrices ...
func (c *Cinemais) GetPrices() ([]models.Price, error) {
	prices, err := cinemais.GetPrices(string(c.complex.Code))
	return c.mapPrices(prices), err
}

// fillMoviesDetails ...
func (c *Cinemais) fillMoviesDetailsAndMap(movies []cinemais.Movie) []models.Movie {
	result := make([]models.Movie, len(movies))
	var wg sync.WaitGroup
	for i := range movies {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			movie, err := cinemais.GetMovie(movies[index].ID)
			if err != nil {
				// TODO: Handle
			} else {
				result[index] = c.mapMovie(*movie)
			}
		}(i)
	}
	wg.Wait()
	return result
}

func (c *Cinemais) mapMovies(movies []cinemais.Movie) []models.Movie {
	result := make([]models.Movie, len(movies))
	for i, m := range movies {
		result[i] = c.mapMovie(m)
	}
	return result
}

func (c *Cinemais) mapMovie(m cinemais.Movie) models.Movie {
	posterURL := ""
	if m.PosterURLs != nil {
		posterURL = m.PosterURLs[cinemais.PosterSizeLarge]
	}
	return models.Movie{
		Slug:          movieutil.GenerateSlug(m.Title),
		ClaqueteID:    m.ID,
		PosterURL:     posterURL,
		OriginalTitle: m.OriginalTitle,
		Title:         m.Title,
		Cast:          m.Cast,
		Distributor:   m.Distributor,
		Genres:        m.Genres,
		Synopsis:      m.Synopsis,
		ReleaseDate:   m.ReleaseDate,
		Rating:        m.Rating,
		Runtime:       int(m.Runtime), // TODO: Convert models.Movie to uint
	}
}

func (c *Cinemais) mapPrices(prices []cinemais.Price) []models.Price {
	var result []models.Price
	for _, p := range prices {
		price := c.mapPrice(p)
		if price.Weight != 0 {
			result = append(result, price)
		}
	}
	return result
}

func (c *Cinemais) mapPrice(price cinemais.Price) models.Price {
	timestamp := time.Now()
	return models.Price{
		TheaterID:         c.t.ID,
		Label:             price.Label,
		Full:              price.Full,
		Half:              price.Half,
		Weekdays:          price.Weekdays,
		ExceptHolidays:    price.ExceptHolidays,
		ExceptPreviews:    price.ExceptPreviews,
		IncludingHolidays: price.IncludingHolidays,
		IncludingPreviews: price.IncludingPreviews,
		Attributes:        price.Attributes,
		Weight:            getWeightForAttributes(price.Attributes),
		CreatedAt:         &timestamp,
	}
}

func (c *Cinemais) mapSession(s cinemais.Session) models.Session {
	m := c.mapMovie(*s.Movie)
	var tz string
	if c.t.City != nil {
		tz = c.t.City.TimeZone
	}
	return models.Session{
		TheaterID: c.t.ID,
		MovieSlug: m.Slug,
		Movie:     &m,
		StartTime: s.StartTime,
		TimeZone:  tz,
		Room:      s.Room,
		Version:   s.Version,
		Format:    s.Format,
	}
}

func (c *Cinemais) mapSessions(sessions []cinemais.Session) []models.Session {
	result := make([]models.Session, len(sessions))
	for i, session := range sessions {
		result[i] = c.mapSession(session)
	}
	return result
}

func getWeightForAttributes(attrs []string) uint {
	equal := func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}

	if equal(attrs, []string{"2D"}) {
		return 1
	} else if equal(attrs, []string{"3D"}) {
		return 2
	} else if equal(attrs, []string{"2D", "Magic D", "Poltrona Tradicional"}) {
		return 3
	} else if equal(attrs, []string{"3D", "Magic D", "Poltrona Tradicional"}) {
		return 4
	} else if equal(attrs, []string{"2D", "3D", "Magic D", "Poltrona VIP"}) {
		return 5
	}
	return 0
}
