package v2

import (
	"net/http"
	"testing"

	"github.com/dsbezerra/amenic/src/lib/middlewares"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTheater(t *testing.T) {
	data := NewMockDataAccessLayer()

	r := NewMockRouter(data)
	r.Use(rest.Init(), middlewares.ValidObjectIDHex(), middlewares.BaseParseQuery())

	s := RESTService{data: data}
	s.ServeTheaters(&r.RouterGroup)

	// Add test data
	testTheater := models.Theater{
		ID:        primitive.NewObjectID(),
		Name:      "Fake Theater",
		ShortName: "Fake",
	}
	err := data.InsertTheater(testTheater)
	assert.NoError(t, err)

	HexID := testTheater.ID.Hex()

	clientAuthToken := getClientAuthToken(t)
	adminAuthToken := getAdminAuthToken(t)

	cases := []apiTestCase{
		apiTestCase{
			name:   "It should return Unauthorized",
			method: "GET",
			url:    "/theaters",
			status: http.StatusUnauthorized,
		},
		apiTestCase{
			name:      "It should return BadRequest since ID is not a valid ObjectId",
			method:    "GET",
			url:       "/theaters/theater/invalid-theater-id",
			status:    http.StatusBadRequest,
			authToken: clientAuthToken,
		},
		apiTestCase{
			name:      "It should return a Theater with ID " + HexID,
			method:    "GET",
			url:       "/theaters/theater/" + HexID,
			status:    http.StatusOK,
			authToken: clientAuthToken,
		},
		apiTestCase{
			name:      "It should return all prices of Theater with ID " + HexID,
			method:    "GET",
			url:       "/theaters/theater/" + HexID + "/prices",
			status:    http.StatusOK,
			authToken: clientAuthToken,
		},
		apiTestCase{
			name:      "It should return OK because the token is a valid admin token",
			method:    "GET",
			url:       "/theaters/count",
			status:    http.StatusOK,
			authToken: adminAuthToken,
		},
	}

	r.RunTests(t, cases)

	err = data.DeleteTheater(testTheater.ID.Hex())
	assert.NoError(t, err)
}
