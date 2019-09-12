package task

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/dsbezerra/amenic/src/lib/config"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/persistence/mongolayer"
	"github.com/dsbezerra/amenic/src/lib/util/scraperutil"
	"github.com/dsbezerra/amenic/src/scraperservice/provider"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	testConnection = "mongodb://localhost/amenic-test"
)

func TestStartScraperCinemais(t *testing.T) {
	// Change our wd because .env is in the upper dir
	os.Chdir("../")
	// Setup env variables
	_, err := config.LoadConfiguration()
	assert.NoError(t, err)

	data := newMockDataAccessLayer()
	theater, scraper, err := ensureCinemaisTestDataExists(data)
	assert.NoError(t, err)

	// Run scraper for theater
	run, err := StartScraper(data, ScraperOptions{
		TheaterID:     theater.ID.Hex(),
		Type:          scraper.Type,
		Provider:      scraper.Provider,
		IgnoreLastRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	printRun(run)
}

func TestStartScraperIbicinemas(t *testing.T) {
	// Change our wd because .env is in the upper dir
	os.Chdir("../")
	// Setup env variables
	_, err := config.LoadConfiguration()
	assert.NoError(t, err)

	data := newMockDataAccessLayer()
	theater, scraper, err := ensureIbicinemaisTestDataExists(data)
	assert.NoError(t, err)

	// Run scraper for theater
	run, err := StartScraper(data, ScraperOptions{
		TheaterID:     theater.ID.Hex(),
		Type:          scraper.Type,
		Provider:      scraper.Provider,
		IgnoreLastRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	printRun(run)
}

func ensureCinemaisTestDataExists(data persistence.DataAccessLayer) (*models.Theater, *models.Scraper, error) {
	theater, err := data.GetTheater("5d4e00db1b3e2d231434d147", data.DefaultQuery())
	if err != nil && err == mongo.ErrNoDocuments {
		ID, _ := primitive.ObjectIDFromHex("5d4e00db1b3e2d231434d147") // Ignoring error since it's a valid hex
		testTheater := models.Theater{
			ID:         ID,
			Name:       "Cinemais Montes Claros",
			ShortName:  "Cinemais",
			InternalID: "34",
		}
		theater = &testTheater
		err = data.InsertTheater(testTheater)
	}

	if err != nil {
		return nil, nil, err
	}

	// Ensure we have a scraper
	scraper, err := data.FindScraper(data.DefaultQuery().
		AddCondition("theaterId", theater.ID).
		AddCondition("type", scraperutil.TypeSchedule).
		AddCondition("provider", provider.ProviderCinemais))
	if err != nil && err == mongo.ErrNoDocuments {
		scraper = &models.Scraper{
			ID:        primitive.NewObjectID(),
			TheaterID: theater.ID,
			Provider:  provider.ProviderCinemais,
			Type:      scraperutil.TypeSchedule,
		}
		err = data.InsertScraper(*scraper)
	}

	return theater, scraper, nil
}

func ensureIbicinemaisTestDataExists(data persistence.DataAccessLayer) (*models.Theater, *models.Scraper, error) {
	theater, err := data.GetTheater("5d4e01661b3e2d231434d148", data.DefaultQuery())
	if err != nil && err == mongo.ErrNoDocuments {
		ID, _ := primitive.ObjectIDFromHex("5d4e01661b3e2d231434d148") // Ignoring error since it's a valid hex
		testTheater := models.Theater{
			ID:         ID,
			Name:       "IBICINEMAS",
			ShortName:  "IBICINEMAS",
			InternalID: "ibicinemas",
		}
		theater = &testTheater
		err = data.InsertTheater(testTheater)
	}

	if err != nil {
		return nil, nil, err
	}

	// Ensure we have a scraper
	scraper, err := data.FindScraper(data.DefaultQuery().
		AddCondition("theaterId", theater.ID).
		AddCondition("type", scraperutil.TypeSchedule).
		AddCondition("provider", provider.ProviderIbicinemas))
	if err != nil && err == mongo.ErrNoDocuments {
		scraper = &models.Scraper{
			ID:        primitive.NewObjectID(),
			TheaterID: theater.ID,
			Provider:  provider.ProviderIbicinemas,
			Type:      scraperutil.TypeSchedule,
		}
		err = data.InsertScraper(*scraper)
	}

	return theater, scraper, nil
}

func newMockDataAccessLayer() persistence.DataAccessLayer {
	data, err := mongolayer.NewMongoDAL(testConnection)
	exitIfErr(err)
	return data
}

func exitIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func printRun(run *models.ScraperRun) {
	fmt.Printf(`
	---  RUN STATS ---

	_id = %s
	scraper_id = %s
	result_code = %s
	error = %s
	start_time = %s
	complete_time = %s
	extracted_count = %d

	--- END STATS ---
	`,
		run.ID,
		run.ScraperID,
		run.ResultCode,
		run.Error,
		run.StartTime,
		run.CompleteTime,
		run.ExtractedCount)

	synopsis := func(synopsis string) string {
		limit := 200
		count := len(synopsis)
		if count > limit {
			synopsis = synopsis[:limit]
			more := count - limit
			return fmt.Sprintf("%s... %d+", synopsis, more)
		}
		return synopsis
	}

	if run.Scraper.Type == "now_playing" || run.Scraper.Type == "upcoming" {
		for _, m := range run.Movies {
			fmt.Printf(`
				id = %s
				claquete_id = %d
				tmdb_id = %d
				imdb_id = %s
				slug = %s
				title = %s
				poster = %s
				synopsis = %s
				genres = %v
				distributor = %s
				release_date = %v
				showtimes =
			`, m.ID.Hex(), m.ClaqueteID, m.TmdbID, m.ImdbID, m.Slug, m.Title, m.PosterURL, synopsis(m.Synopsis), m.Genres, m.Distributor, m.ReleaseDate)

			for _, s := range m.Sessions {
				fmt.Printf(`
				  	cinema_id = %s
				  	movie_id = %s
				  	movie_slug = %s
				  	room = %d
						start_time = %s
				  	version = %s
						format = %s
			`, s.TheaterID, s.MovieID.Hex(), s.MovieSlug, s.Room, s.StartTime, s.Version, s.Format)
			}
		}
	}

	if run.Scraper.Type == "prices" {
		for _, p := range run.Prices {
			fmt.Printf(`
				theater_id = %s
				attrs = %v
				full = %f
				half = %f
				weekdays = %v
			`, p.TheaterID, p.Attributes, p.Full, p.Half, p.Weekdays)
		}
	}
}
