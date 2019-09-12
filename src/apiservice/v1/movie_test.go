package v1

import (
	"net/http"
	"testing"

	"github.com/dsbezerra/amenic/src/lib/middlewares"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMovie(t *testing.T) {
	data := NewMockDataAccessLayer()

	r := NewMockRouter(data)
	r.Use(rest.Init(), middlewares.ValidObjectIDHex(), middlewares.BaseParseQuery())

	testMovie := models.Movie{
		ID:    primitive.NewObjectID(),
		Title: "Test Movie",
	}
	err := data.InsertMovie(testMovie)
	assert.NoError(t, err)

	s := RESTService{data: data}
	s.ServeMovies(&r.RouterGroup)

	HexID := testMovie.ID.Hex()

	cases := []apiTestCase{
		apiTestCase{
			name:         "It should return Unauthorized",
			method:       "GET",
			url:          "/movies",
			status:       http.StatusUnauthorized,
			appendAPIKey: false,
		},
		apiTestCase{
			name:         "It should return BadRequest since ID is not a valid ObjectId",
			method:       "GET",
			url:          "/movies/movie/invalid-movie-id",
			status:       http.StatusBadRequest,
			appendAPIKey: true,
		},
		apiTestCase{
			name:         "It should return NotFound since Movie with ID 5c353e8cebd54428b4f25447 doesn't exist",
			method:       "GET",
			url:          "/movies/movie/5c353e8cebd54428b4f25447",
			status:       http.StatusNotFound,
			appendAPIKey: true,
		},
		apiTestCase{
			name:         "It should return a Movie with ID " + HexID,
			method:       "GET",
			url:          "/movies/movie/" + HexID,
			status:       http.StatusOK,
			appendAPIKey: true,
		},
	}

	r.RunTests(t, cases)

	err = data.DeleteMovie(testMovie.ID.Hex())
	assert.NoError(t, err)
}
