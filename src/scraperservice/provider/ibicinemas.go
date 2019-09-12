package provider

import (
	"strings"
	"sync"
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/movieutil"
	"github.com/dsbezerra/ibicinemas"
	"github.com/sirupsen/logrus"
)

type (
	// Ibicinemas ...
	Ibicinemas struct {
		t   *models.Theater
		log *logrus.Entry
	}
)

// NewIbicinemas ...
func NewIbicinemas() *Ibicinemas {
	// TODO: Get theater data from database
	return &Ibicinemas{
		log: logrus.WithField("provider", "ibicinemas"),
	}
}

// Init ...
func (i *Ibicinemas) Init(data persistence.DataAccessLayer) error {
	query := data.DefaultQuery().
		AddCondition("internalId", "ibicinemas").
		AddInclude("city")
	theater, err := data.FindTheater(query)
	if err != nil {
		return err
	}
	i.t = theater
	return nil
}

// GetNowPlaying ...
func (i *Ibicinemas) GetNowPlaying() ([]models.Movie, error) {
	movies, err := ibicinemas.GetNowPlaying()
	return i.fillMoviesDetailsAndMap(movies), err
}

// GetUpcoming ...
func (i *Ibicinemas) GetUpcoming() ([]models.Movie, error) {
	movies, err := ibicinemas.GetUpcoming()
	return i.fillMoviesDetailsAndMap(movies), err
}

// GetSchedule ...
func (i *Ibicinemas) GetSchedule() ([]models.Session, error) {
	schedule, err := ibicinemas.GetSchedule()
	return i.mapSessions(schedule.Sessions), err
}

// GetPrices ...
func (i *Ibicinemas) GetPrices() ([]models.Price, error) {
	prices, err := ibicinemas.GetPrices()
	if err != nil {
		return nil, err
	}

	// Mapping prices result to amenic Price model
	result := make([]models.Price, 0)
	for _, p := range prices {
		var weight uint
		var label string
		var attrs []string

		switch p.Projection {
		case ibicinemas.Projection2D:
			label = "Projeção 2D"
			attrs = []string{"2D"}
			weight = 1
		case ibicinemas.Projection3D:
			label = "Projeção 3D"
			attrs = []string{"3D"}
			weight = 2
		default:
			// TODO: logging
			continue
		}

		var includingHolidays, includingPreviews bool
		weekdays := make([]time.Weekday, 0)
		for _, d := range p.Days {
			switch d.ID {
			case ibicinemas.Holiday:
				includingHolidays = true
			case ibicinemas.Preview:
				includingPreviews = true
			default:
				w := models.NameToTimeWeekday(d.Name)
				if w >= time.Sunday && w <= time.Saturday {
					weekdays = append(weekdays, w)
				}
			}
		}

		if len(weekdays) > 0 {
			timestamp := time.Now()
			result = append(result, models.Price{
				TheaterID:         i.t.ID,
				Label:             label,
				Full:              p.Full,
				Half:              p.Half,
				Weekdays:          weekdays,
				IncludingHolidays: includingHolidays,
				IncludingPreviews: includingPreviews,
				Attributes:        attrs,
				Weight:            weight,
				CreatedAt:         &timestamp,
			})
		}
	}

	return result, err
}

// fillMoviesDetails ...
func (i *Ibicinemas) fillMoviesDetailsAndMap(movies []ibicinemas.Movie) []models.Movie {
	result := make([]models.Movie, len(movies))
	var wg sync.WaitGroup
	for ii := range movies {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			page := movies[index].DetailPage
			path := strings.Replace(page, "http://www.ibicinemas.com.br/", "", -1)
			path = strings.Replace(path, ".html", "", -1)
			movie, err := ibicinemas.GetMovie(path)
			if err != nil {
				// TODO: Handle
			} else {
				result[index] = i.mapMovie(*movie)
			}
		}(ii)
	}
	wg.Wait()
	return result
}

// mapMovie ...
func (i *Ibicinemas) mapMovie(m ibicinemas.Movie) models.Movie {
	return models.Movie{
		Slug:        movieutil.GenerateSlug(m.Title),
		PosterURL:   m.Poster,
		Title:       m.Title,
		Distributor: m.Distributor,
		Genres:      m.Genres,
		Synopsis:    m.Synopsis,
		Rating:      m.Rating,
		Runtime:     int(m.Runtime), // TODO: Convert models.Movie to uint
	}
}

func (i *Ibicinemas) mapSession(s ibicinemas.Session) models.Session {
	m := models.Movie{
		Slug:  movieutil.GenerateSlug(s.MovieTitle),
		Title: s.MovieTitle,
	}
	var tz string
	if i.t.City != nil {
		tz = i.t.City.TimeZone
	}
	return models.Session{
		Movie:     &m,
		TheaterID: i.t.ID,
		MovieSlug: m.Slug,
		StartTime: s.StartTime,
		TimeZone:  tz,
		Room:      uint(s.Room),
		Version:   s.Version,
		Format:    s.Format,
	}
}

func (i *Ibicinemas) mapSessions(sessions []ibicinemas.Session) []models.Session {
	result := make([]models.Session, len(sessions))
	for ii, session := range sessions {
		result[ii] = i.mapSession(session)
	}
	return result
}
