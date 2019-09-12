package extractors

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/agnivade/levenshtein"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/movieutil"
	"github.com/dsbezerra/amenic/src/lib/util/scraperutil"
	"github.com/dsbezerra/amenic/src/scraperservice/provider"
	tmdb "github.com/ryanbradynd05/go-tmdb"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	// MovieExtractor ...
	MovieExtractor struct {
		Data     persistence.DataAccessLayer
		Provider provider.Provider
		Type     string
		Run      *models.ScraperRun
		Logger   *logrus.Entry
		TMDb     *tmdb.TMDb
		Movies   []models.Movie
	}
)

// Errors
var errTmdbSearchMatch = errors.New("tmdb search match error")
var errTmdbMovieNotFound = errors.New("tmdb search movie not found")

// NewMovieExtractor creates a new extractor configured to insert movies in the Movie collection
func NewMovieExtractor(data persistence.DataAccessLayer, p provider.Provider, s *models.ScraperRun) *MovieExtractor {
	result := &MovieExtractor{
		Logger:   logrus.WithFields(logrus.Fields{"extractor": "Movie"}),
		Type:     s.Scraper.Type,
		Data:     data,
		Provider: p,
		Run:      s,
	}

	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		// TODO: Don't fetch metadata if this is missing
		log.Fatal("missing TMDB_API_KEY env variable")
	}

	result.TMDb = tmdb.Init(tmdb.Config{APIKey: apiKey})

	return result
}

// Execute extract prices information for a given provider
func (e *MovieExtractor) Execute() error {
	var movies []models.Movie
	var err error

	if e.Type == scraperutil.TypeNowPlaying {
		movies, err = e.Provider.GetNowPlaying()
	} else if e.Type == scraperutil.TypeUpcoming {
		movies, err = e.Provider.GetUpcoming()
	} else {
		err = errors.New("MovieExtractor supports only now_playing and upcoming types")
	}
	e.Movies = movies
	return err
}

// Complete ...
func (e *MovieExtractor) Complete() {
	var wg sync.WaitGroup
	for index := range e.Movies {
		wg.Add(1)
		go func(m *models.Movie) {
			defer wg.Done()
			// TODO: Create channel to handle errors
			e.FillMovieMetadata(m)
			// TODO: Handle this error.
			e.UpsertMovie(m)
		}(&e.Movies[index])
	}
	wg.Wait()

	e.Run.Movies = e.Movies
}

// ExtractedHash TODO
func (e *MovieExtractor) ExtractedHash() string {
	return GetExtractedHash(e.Movies)
}

// ExtractedCount TODO
func (e *MovieExtractor) ExtractedCount() int {
	return len(e.Movies)
}

// Configuration used to resolve image paths
var tmdbConfig *tmdb.Configuration

// FillMovieMetadata calls TMDb Api to get metadata and fill our Movie
// TODO: Add claquete?
func (e *MovieExtractor) FillMovieMetadata(movie *models.Movie) error {
	if tmdbConfig == nil {
		e.Logger.Debugln("Configuration is missing. Getting configurations...")
		config, err := e.TMDb.GetConfiguration()
		if err != nil {
			return err
		}
		tmdbConfig = config
		e.Logger.Debugln("Configuration successfully retrieved!")
	}

	y := time.Now().Year()
	// Title is the query we will use to search for a movie in TMDB (or our own database later)
	q := strings.Replace(strings.ToLower(movie.Title), "o filme", "", -1)
	if strings.HasSuffix(q, "-") {
		q = strings.TrimSpace(q[0 : len(q)-1])
	}

	// Search in TMDb
	tmdbMovie, err := e.SearchTMDBMovie(movie, q, y)
	if err != nil {
		switch err {
		default:
			return err
		case errTmdbMovieNotFound, errTmdbMovieNotFound:
			// Check in previous year
			tmdbMovie, err = e.SearchTMDBMovie(movie, q, y-1)
			if err != nil {
				return err
			}
			// NOTE(diego): This year thing only works for movies registered with release date.
			// Movies that doesn't have this info doesn't get returned by the API.
			//
			// We may consider querying without year paramter to find these movies.
		}
	}
	return e.ApplyTMDBMovieInfo(*tmdbMovie, movie)
}

// SearchTMDBMovie searches for a movie with the given year and query
func (e *MovieExtractor) SearchTMDBMovie(movie *models.Movie, query string, year int) (*tmdb.MovieShort, error) {
	var result *tmdb.MovieShort
	var err error

	e.Logger.Infof("Searching metadata for movie '%s' with query '%s'", movie.Title, query)
	results, err := e.TMDb.SearchMovie(query, map[string]string{
		"language": "pt-BR",
		"region":   "BR",
		"year":     strconv.Itoa(year),
	})
	if err == nil {
		if results.TotalResults > 0 {
			slug := movieutil.GenerateSlug(query)
			for _, movieShort := range results.Results {
				// NOTE(diego): Ridiculous simple match algorithm below.
				// We should get more info about movie from theater's page, for example, the cast and genres to
				// improve this ridiculous match algorithm, *BUT* for now it works, mostly because we just care
				// about current playing and upcoming movies.
				movieShortSlug := movieutil.GenerateSlug(movieShort.Title)
				// If title matches, we say that we have found the correct movie.
				if slug == movieShortSlug {
					result = &movieShort
					break
				} else {
					// NOTE(diego): If our slug is at the index of the other slug the title may contain additional title
					// and it is probably a match...
					// TODO(diego): Change this if we get any wrong matches or report as pending so we can manually check
					// later if it a match or not....
					if strings.Index(slug, movieShortSlug) == 0 {
						result = &movieShort
						break
					} else {
						// Otherwise, try to match by using Damerau-Levensthein edit distance.
						// If we get an edit distance greater than 4 for any word in the title,
						// we report as not a match.
						if levenshtein.ComputeDistance(slug, movieShortSlug) <= 4 {
							// TODO: Report as pending so we can manually check later if it a match or not.
							result = &movieShort
							break
						}
					}
				}
			}

			if result == nil || result.ID == 0 {
				err = errTmdbSearchMatch
			}

		} else {
			e.Logger.Warnf("Nothing was found with query '%s'!", query)
			err = errTmdbMovieNotFound
		}
	} else {
		e.Logger.Errorln(err.Error())
	}

	return result, err
}

// ApplyTMDBMovieInfo applies movie information obtained from themoviedb to our scraped movie
func (e *MovieExtractor) ApplyTMDBMovieInfo(movieShort tmdb.MovieShort, movie *models.Movie) error {
	movieInfo, err := e.TMDb.GetMovieInfo(movieShort.ID, map[string]string{
		"language":           "pt-BR",
		"append_to_response": "videos,releases",
	})
	if err != nil {
		return err
	}

	e.Logger.Infoln("Updating movie metadata...")

	// Update with metadata from TMDb
	movie.TmdbID = movieInfo.ID
	movie.ImdbID = movieInfo.ImdbID
	movie.OriginalTitle = movieInfo.OriginalTitle
	movie.Title = movieInfo.Title
	if movie.Synopsis == "" && movieInfo.Overview != "" {
		movie.Synopsis = movieInfo.Overview
	}
	if movieInfo.BackdropPath != "" {
		// TODO: Ensure this image exists
		movie.BackdropURL = tmdbConfig.Images.SecureBaseURL + "w1400_and_h450_bestv2" + movieInfo.BackdropPath
	}
	if movieInfo.PosterPath == "" && movieShort.PosterPath != "" {
		movieInfo.PosterPath = movieShort.PosterPath
	}
	if movieInfo.PosterPath != "" && movie.PosterURL == "" {
		// TODO: Ensure this image exists
		movie.PosterURL = tmdbConfig.Images.SecureBaseURL + "w300_and_h450_bestv2" + movieInfo.PosterPath
	}
	if movieInfo.Runtime != 0 && movie.Runtime == 0 {
		movie.Runtime = int(movieInfo.Runtime)
	}
	movie.Genres = make([]string, len(movieInfo.Genres))
	for i, genre := range movieInfo.Genres {
		movie.Genres[i] = genre.Name
	}
	sort.Strings(movie.Genres)

	if movie.ReleaseDate == nil || movie.ReleaseDate.IsZero() {
		for _, country := range movieInfo.Releases.Countries {
			if country.Iso3166_1 == "BR" {
				loc, _ := time.LoadLocation("America/Sao_Paulo")
				t, _ := time.ParseInLocation("2006-01-02", country.ReleaseDate, loc)
				movie.ReleaseDate = &t
				break
			}
		}
	}

	if movieInfo.Videos != nil {
		for _, video := range movieInfo.Videos.Results {
			if video.Type == "Trailer" && video.Site == "YouTube" {
				movie.Trailer = video.Key
				break
			}
		}
	}

	return nil
}

// UpsertMovie finds movie in database and use it to fill missing fields.
// It tries to find by title/slug.
func (e *MovieExtractor) UpsertMovie(movie *models.Movie) {
	if movie.Title == "" {
		e.Logger.Warn("Aborted movie upsert due to empty title")
		return
	}

	var result *models.Movie
	if movie.Slug == "" {
		movie.Slug = movieutil.GenerateSlug(movie.Title)
	}

	// TODO: Improve those logging messages
	found, result := FindMovieMatch(e.Data, movie)
	if found {
		// NOTE(diego):
		// Only updating movies if the provider is VeloxTickets/Claquete/Cinemais because
		// IBICINEMAS may contain wrong data. But this can make movies releasing only in
		// IBICINEMAS outdated.
		// TODO: Elaborate this logic.
		e.Logger.Infof("Movie '%s' is in database. Checking for update...", movie.Title)
		update, u := movieutil.ShouldUpdate(result, movie)
		if update && e.Run.Scraper.Provider != provider.ProviderIbicinemas {
			e.Logger.Debugf("Updating movie: %s (%s)...", u.ID.Hex(), movie.Title)
			updatedAt := time.Now()
			u.UpdatedAt = &updatedAt
			_, err := e.Data.UpdateMovie(u.ID.Hex(), u)
			if err != nil {
				e.Logger.Error(err.Error())
			} else {
				e.Logger.Info("Updated.")
			}
		} else {
			e.Logger.Info("No update needed.")
		}

		// Let's use whatever we found as the correct data for now on.
		*movie = u
	} else {
		e.Logger.Infof("Movie '%s' is not in database.", movie.Title)

		createdAt := time.Now()
		movie.ID = primitive.NewObjectID()
		movie.CreatedAt = &createdAt

		err := e.Data.InsertMovie(*movie)
		if err != nil {
			e.Logger.Error(err.Error())
		} else {
			e.Logger.Infof("Movie '%s' is now in database", movie.Title)
		}
	}
}

// FindMovieMatch ...
func FindMovieMatch(data persistence.DataAccessLayer, movie *models.Movie) (bool, *models.Movie) {
	var result *models.Movie
	// Try to find it by Claquete ID
	if movie.ClaqueteID != 0 {
		result, _ := data.FindMovie(data.DefaultQuery().
			AddCondition("claqueteId", movie.ClaqueteID))
		if result != nil {
			return true, result
		}
	}

	// Try to find it by TMDb ID
	if movie.TmdbID != 0 {
		result, _ := data.FindMovie(data.DefaultQuery().AddCondition("tmdbId", movie.TmdbID))
		if result != nil {
			return true, result
		}
	}

	// Try to find it by IMDb ID
	if movie.ImdbID != "" {
		result, _ := data.FindMovie(data.DefaultQuery().AddCondition("imdbId", movie.ImdbID))
		if result != nil {
			return true, result
		}
	}

	// Try to find it by slug
	result, err := data.FindMovie(data.DefaultQuery().AddCondition("slug", movie.Slug))
	if result != nil {
		return true, result
	}

	// Query the films that has the closest title sorted by matching score
	query := data.DefaultQuery().
		AddCondition("$text", bson.M{"$search": movie.Title}).
		SetFields(bson.M{
			"score": bson.M{"$meta": "textScore"},
		}).
		SetSort("$testScore:score")
	possible, err := data.GetMovies(query)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}
	// TODO: Improve the match algorithm here if necessary
	for _, m := range possible {
		distance := levenshtein.ComputeDistance(m.Title, movie.Title)
		if distance <= 2 {
			result = &m
			break
		}
	}
	// NOTE: Two loops because we want to make sure we check every possible
	// movie before checking if part of title exists.
	for _, m := range possible {
		// If possible result title exists in scraped movie and it's
		// right in the beginning of string we treat as a match.
		//
		// NOTE: This can be wrong, but this code will only be executed if we don't
		// have the Claquete ID or couldn't find the movie in TMDb.
		//
		// Was added to make possible to match movies missing optional text in title.
		// eg: 'Lino - O Filme' missing '- O Filme' is still the same movie.
		if strings.Index(movie.Title, m.Title) == 0 {
			result = &m
			break
		}
	}
	if result != nil {
		return true, result
	}
	return false, nil
}
