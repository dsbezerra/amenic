package extractors

import (
	"log"
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/scraperutil"
	"github.com/dsbezerra/amenic/src/lib/util/timeutil"
	"github.com/dsbezerra/amenic/src/scraperservice/provider"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	ScheduleExtractor struct {
		Data     persistence.DataAccessLayer
		Provider provider.Provider
		Run      *models.ScraperRun
		Sessions []models.Session
	}
)

// NewScheduleExtractor ...
func NewScheduleExtractor(data persistence.DataAccessLayer, p provider.Provider, s *models.ScraperRun) *ScheduleExtractor {
	result := &ScheduleExtractor{
		Data:     data,
		Run:      s,
		Provider: p,
	}
	return result
}

// Execute extract schedule information for a given provider
func (e *ScheduleExtractor) Execute() error {
	result, err := e.Provider.GetSchedule()
	if err != nil {
		return err
	}

	movies := LookupMovies(e.Data, result)
	m := map[string]primitive.ObjectID{}
	for _, movie := range movies {
		m[movie.Slug] = movie.ID
	}

	for i := range result {
		v, ok := m[result[i].MovieSlug]
		if ok {
			result[i].MovieID = v
		}
		result[i].Movie = nil
	}
	e.Sessions = result
	return nil
}

// Complete TODO
func (e *ScheduleExtractor) Complete() {

	switch e.Run.ResultCode {
	case scraperutil.RunResultSuccess:
		start, end := timeutil.Now(), timeutil.Now()
		for _, session := range e.Sessions {
			if session.StartTime.Before(start) {
				start = *session.StartTime
			}
			if session.StartTime.After(end) {
				end = *session.StartTime
			}
		}
		query := e.Data.DefaultQuery().
			AddCondition("$and", []bson.D{
				bson.D{
					{Key: "theaterId", Value: e.Run.Scraper.TheaterID},
					{Key: "startTime", Value: bson.D{{Key: "$gte", Value: start}}},
					{Key: "startTime", Value: bson.D{{Key: "$lte", Value: end}}},
				},
			})
		_, err := e.Data.DeleteSessions(query)
		if err != nil {
			// TODO: Handle
			log.Fatal(err)
		}
		err = e.Data.InsertSessions(e.Sessions...)
		if err != nil {
			// TODO: Handle
			log.Fatal(err)
		}
	case scraperutil.RunResultNotModified:
		fallthrough
	default:
		// Do nothing.
	}
}

// ExtractedHash TODO
func (e *ScheduleExtractor) ExtractedHash() string {
	return GetExtractedHash(e.Sessions)
}

// ExtractedCount TODO
func (e *ScheduleExtractor) ExtractedCount() int {
	return len(e.Sessions)
}

// LookupMovies ...
func LookupMovies(data persistence.DataAccessLayer, sessions []models.Session) []models.Movie {
	defer timeutil.TimeTrack(time.Now(), "LookupMovies")

	claquete := []int{}
	slugs := []string{}

	movies := make([]models.Movie, 0)

	seen := map[string]models.Movie{}
	for _, s := range sessions {
		_, ok := seen[s.MovieSlug]
		if !ok {
			seen[s.MovieSlug] = *s.Movie
			if s.Movie.ClaqueteID != 0 {
				claquete = append(claquete, s.Movie.ClaqueteID)
			} else if s.MovieSlug != "" {
				slugs = append(slugs, s.MovieSlug)
			}
		}
	}

	if len(claquete) > 0 {
		// @Refactor support other databases.
		m, err := data.GetMovies(data.DefaultQuery().
			AddCondition("_id", bson.M{"$in": claquete}))
		if err == nil {
			movies = append(movies, m...)
		}
	}

	if len(slugs) > 0 {
		// @Refactor support other databases.
		m, err := data.GetMovies(data.DefaultQuery().
			AddCondition("slug", bson.M{"$in": slugs}))
		if err == nil {
			movies = append(movies, m...)
		}
	}

	// @Improve this check
	if len(movies) == len(seen) {
		return movies
	}

	// We still need to find movies
	f := map[string]models.Movie{}
	for _, m := range movies {
		f[m.Slug] = m
	}

	for k, m := range seen {
		_, ok := f[k]
		if !ok {
			found, movie := FindMovieMatch(data, &m)
			if found {
				movies = append(movies, *movie)
			}
		}
	}

	return movies
}
