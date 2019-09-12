package task

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/movieutil"
	"github.com/dsbezerra/amenic/src/scoreservice/moviescore"
)

var (
	// ErrNowPlayingMoviesNotFound indicates now playing movies not found
	ErrNowPlayingMoviesNotFound = errors.New("now playing movies not found")
	// ErrScoresNotFound indicates scores were not found
	ErrScoresNotFound = errors.New("scores were not found")
)

// SyncScores first ensures that only movies currently in theaters will have their
// scores synced.
//
// After that, it calls FindScoresForNowPlaying, which makes sure our movie has a
// score document.
//
// Then it finally calls Sync to actually sync the scores and updates the database.
func SyncScores(data persistence.DataAccessLayer) error {
	_, err := EnsureSyncOnlyNowPlayingMovies(data)
	if err != nil {
		return err
	}

	err = FindScoresForNowPlayingMovies(data)
	if err != nil {
		return err
	}

	// Now actually sync.
	return Sync(data, 10, 0, true)
}

// Sync retrieves all scores that should be kept sync and recursively syncs
// each one of them.
func Sync(data persistence.DataAccessLayer, limit, skip int64, r bool) error {
	// Limit max sync documents to 10.
	if limit > 10 {
		limit = 10
	}

	query := data.DefaultQuery().
		AddCondition("keepSynced", true).
		SetLimit(limit).
		SetSkip(skip)

	scores, err := data.GetScores(query)
	if err != nil {
		return err
	}

	len := int64(len(scores))
	if len == 0 {
		return ErrScoresNotFound
	}

	// Sync scores...
	wg := sync.WaitGroup{}

	for _, score := range scores {
		wg.Add(1)

		go func(inner models.Score) {
			defer wg.Done()

			original := inner
			synced := &inner

			log.Printf("Syncing score for movie %s...\n", inner.MovieID.Hex())

			result, _ := StartMovieScore(inner, moviescore.IMDB)
			applyScoreResult(result, synced)

			result, _ = StartMovieScore(inner, moviescore.RottenT)
			applyScoreResult(result, synced)

			if original != *synced {
				_, err := data.UpdateScore(synced.ID.Hex(), *synced)
				if err != nil {
					// TODO(diego): Better logging.
					log.Printf("Error: %s", err.Error())
				} else {
					log.Println("Score syncing was successful.")
				}
			} else {
				log.Println("Score syncing was successful, but nothing changed.")
			}

		}(score)
	}

	wg.Wait()

	if len < limit || !r {
		// We don't need to call this again.
		log.Println("End of score collection reached.\nClosing session...")
		return nil
	}

	return Sync(data, limit, skip+len, r)
}

// EnsureSyncOnlyNowPlayingMovies makes sure we are syncing only now playing movies
func EnsureSyncOnlyNowPlayingMovies(data persistence.DataAccessLayer) ([]models.Movie, error) {
	playing, err := data.GetNowPlayingMovies(nil)
	if err != nil {
		return nil, err
	}

	scores, err := data.GetScores(
		data.DefaultQuery().AddCondition("keepSynced", true).SetLimit(-1)) // -1 ensure we get all scores
	if err != nil {
		return playing, err
	}

	for _, score := range scores {
		keepSynced := false

		for _, m := range playing {
			if m.ID == score.MovieID {
				keepSynced = true
				break
			}
		}

		if !keepSynced {
			score.KeepSynced = false
			_, err := data.UpdateScore(score.ID.Hex(), score)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return playing, nil
}

// FindScoresForNowPlayingMovies gets now playing movies and checks if they have
// a corresponding score document in Scores collection. If it does not have one,
// a find score for movie operation will be executed to retrieve and create a score
// document. Otherwise, it ensures the score is marked to sync.
func FindScoresForNowPlayingMovies(data persistence.DataAccessLayer) error {
	playing, err := data.GetNowPlayingMovies(nil)
	if err != nil {
		return err
	}

	size := len(playing)
	if size == 0 {
		return ErrNowPlayingMoviesNotFound
	}

	// TODO(diego): Goroutines?
	sq := data.DefaultQuery()
	for _, m := range playing {
		sq.AddCondition("movieId", m.ID)
		scores, err := data.GetScores(sq)
		if err != nil {
			// TODO: Log error to logging system
			continue
		}

		missingScores := getMissingScores(scores)
		if len(missingScores) > 0 {
			log.Printf("We don't have all scores for movie %s.", m.Title)

			// We don't have scores for this movie, let's find it!
			movie := models.Movie{
				ID:            m.ID,
				ImdbID:        m.ImdbID,
				Title:         m.Title,
				OriginalTitle: m.OriginalTitle,
				ReleaseDate:   m.ReleaseDate,
			}

			found, result := FindScoresForMovie(movie, missingScores)
			if found {
				log.Printf("Found scores for movie %s (%s)", movie.Title, movie.OriginalTitle)

				isNew := len(scores) == 0

				now := time.Now()
				score := models.Score{
					MovieID:    movie.ID,
					KeepSynced: true,
					CreatedAt:  &now,
				}

				if !isNew {
					item := scores[0]
					score.ID = item.ID
					score.Imdb = item.Imdb
					score.Rotten = item.Rotten
					score.CreatedAt = item.CreatedAt
				}

				if result.IMDb.ID != "" {
					score.Imdb.ID = result.IMDb.ID
					score.Imdb.Score = result.IMDb.Score
				}

				if result.Rotten.ID != "" {
					score.Rotten.Path = result.Rotten.ID
					score.Rotten.Score = int(result.Rotten.Score)
					score.Rotten.Class = result.Rotten.ScoreClass
				}

				// If it's a new score we insert.
				if isNew {
					err = data.InsertScore(score)
				} else {
					_, err = data.UpdateScore(score.ID.Hex(), score)
				}

				if err != nil {
					log.Printf("Error: %s", err.Error())
				}

			} else {
				ot := movie.OriginalTitle
				if ot == "" {
					ot = "missing original title"
				}
				log.Printf("Couldn't find scores for movie %s (%s)", movie.Title, ot)
			}

			continue
		}

		if len(scores) != 0 {
			// We have scores, let's check if we are setting them to sync.
			doc := scores[0]
			if !doc.KeepSynced {
				doc.KeepSynced = true
				_, err := data.UpdateScore(doc.ID.Hex(), doc)
				if err != nil {
					log.Printf("Error: %s\n", err.Error())
				}
			}
		}
	}

	return err
}

// FindScoresForMovie tries to search scores for the given movie and missing providers.
// Currently supported providers are: Rotten Tomatoes (rotten) and Internet Movie Database (IMDb)
func FindScoresForMovie(movie models.Movie, providers []string) (bool, moviescore.FindScoresResult) {
	result := moviescore.FindScoresResult{}
	if movie.OriginalTitle == "" {
		return false, result
	}

	movieID := movie.ID.Hex()
	log.Printf("Finding score for movie %s (%s).", movie.Title, movie.OriginalTitle)

	foundRotten := false
	foundImdb := false

	originalTitle := strings.ToLower(movie.OriginalTitle)
	title := strings.ToLower(movie.Title)

	// Get score for Rotten (we can get this from search API so it's one step less than IMDb)
	if isScoreMissing("rotten", providers) {
		r, _ := moviescore.RunRottenSearch(originalTitle, movieID)
		if r == nil || len(r.Items) == 0 {
			r, _ = moviescore.RunRottenSearch(title, movieID)
		}
		if r != nil {
			found, r := findCorrectMovie(movie, r.Items)
			if found {
				result.Rotten = r
				foundRotten = true
			}
		}
	}

	if isScoreMissing("imdb", providers) {
		// Get score for IMDb by searching if we don't have an ID
		if movie.ImdbID == "" {
			r, _ := moviescore.RunIMDbSearch(originalTitle, movieID)
			if r != nil {
				found, r := findCorrectMovie(movie, r.Items)
				if found {
					movie.ImdbID = r.ID
					result.IMDb = r
				}
			}
		}

		// If we have we just need to get.
		if movie.ImdbID != "" {
			r, _ := moviescore.RunIMDbScore(movie.ImdbID, movieID)
			if r != nil {
				if found := len(r.Items) == 1; found {
					result.IMDb = r.Items[0]
					foundImdb = true
				}
			}
		}
	}

	found := foundRotten || foundImdb

	return found, result
}

func isScoreMissing(provider string, missing []string) bool {
	for _, p := range missing {
		if p == provider {
			return true
		}
	}

	return false
}

func findCorrectMovie(movie models.Movie, possibleResults []moviescore.ResultItem) (bool, moviescore.ResultItem) {
	var found moviescore.ResultItem

	year := time.Now().Year()
	if movie.ReleaseDate != nil {
		year = movie.ReleaseDate.Year()
	}

	Abs := func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	}

	originalSlug := movieutil.GenerateSlug(movie.OriginalTitle)
	nationalSlug := movieutil.GenerateSlug(movie.Title)

	// TODO: use cast members here once we include cast in our movie data
	for _, r := range possibleResults {
		resultSlug := movieutil.GenerateSlug(r.Title)
		slugMatched := resultSlug == originalSlug || resultSlug == nationalSlug
		// NOTE: Here we compare year of releases, but we are comparing year of release in Brazil vs original country.
		// This makes this code wrong for movies released in different years.
		// (ex: Spider-Man: Into the Spider-Verse and Mortal Engines)
		//
		// To make this easier we should also store original (or domestic) release date.
		//
		// Still, this helps us ignore older movies of recent remakes, such as Predator (2018).
		//
		// dbezerra - 11 jan. 2019
		yearMatched := r.Year == year
		if slugMatched && yearMatched {
			found = r
			break
		} else {
			// TODO(diego): Improve this. Use cast or any other information (beyond title) to help use match a movie.
			if slugMatched {
				// NOTE(diego):
				//
				// Temporary fix to handle movie that was released in another year in Brazil.
				// Remove this once the original movie release date is included in our movie model.
				// This should be enough to ignore older movies and get the recent remakes.
				diff := Abs(year - r.Year)
				if diff <= 1 {
					found = r
					break
				}
			}
		}
	}

	return found.ID != "", found
}

func getMissingScores(scores []models.Score) []string {
	if len(scores) == 0 {
		return []string{"rotten", "imdb"}
	}

	missing := make([]string, 0)

	item := scores[0]
	if item.Imdb.ID == "" {
		missing = append(missing, "imdb")
	}
	if item.Rotten.Path == "" {
		missing = append(missing, "rotten")
	}
	return missing
}

// StartMovieScore executes the movie score fetcher to obtain updated scores
func StartMovieScore(score models.Score, provider string) (*moviescore.Result, error) {
	id := ""
	switch provider {
	case "imdb":
		id = score.Imdb.ID
	case "rotten":
		id = score.Rotten.Path
	}

	if id == "" {
		return nil, nil
	}

	result, err := moviescore.RunScore(provider, id, score.MovieID.Hex())
	return result, err
}

func applyScoreResult(data *moviescore.Result, s *models.Score) {
	if data == nil || len(data.Items) != 1 {
		return
	}

	item := data.Items[0]

	switch data.Provider {
	case "imdb":
		s.Imdb.Score = item.Score
	case "rotten":
		s.Rotten.Class = item.ScoreClass
		s.Rotten.Score = int(item.Score)
	default:
	}
}
