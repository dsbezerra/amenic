package task

import (
	"log"
	"testing"
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/persistence/mongolayer"
	"github.com/dsbezerra/amenic/src/lib/util/scheduleutil"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	testConnection = "mongodb://localhost/amenic-test"
)

func TestCheckOpeningMovies(t *testing.T) {
	data := newMockDataAccessLayer()

	mockTheater := &models.Cinema{
		ID:   "mock-theater",
		Name: "Test Theater",
	}

	err := data.InsertCinema(mockTheater)
	exitIfErr(err)

	tt := time.Now().UTC()
	mockMovie := &models.Movie{
		ID:          bson.NewObjectId(),
		Title:       "Test Movie",
		ReleaseDate: &tt,
	}

	period := scheduleutil.GetWeekPeriod(nil)
	mockShowtimes := []models.Showtime{
		models.Showtime{
			ID:       bson.NewObjectId(),
			MovieID:  mockMovie.ID,
			CinemaID: mockTheater.ID,
			Period: struct {
				Start time.Time `json:"start,omitempty" bson:"start,omitempty"`
				End   time.Time `json:"end,omitempty" bson:"end,omitempty"`
			}{
				Start: period.Start,
				End:   period.End,
			},
		},
	}

	err = data.InsertMovie(mockMovie)
	exitIfErr(err)

	for _, s := range mockShowtimes {
		err = data.InsertShowtime(&s)
		exitIfErr(err)
	}

	CheckOpeningMovies(data)

	err = data.RemoveAllShowtimes(data.DefaultQuery().
		AddCondition("movieId", mockMovie.ID))
	exitIfErr(err)

	err = data.RemoveCinema(mockTheater.ID)
	exitIfErr(err)

	err = data.RemoveMovie(mockMovie.ID.Hex())
	exitIfErr(err)
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
