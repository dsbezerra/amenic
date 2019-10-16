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

func TestScore(t *testing.T) {
	data := NewMockDataAccessLayer()

	r := NewMockRouter(data)
	r.Use(rest.Init(), middlewares.ValidObjectIDHex(), middlewares.BaseParseQuery())

	s := RESTService{data: data}
	s.ServeScores(&r.RouterGroup)

	// Add test data
	testScore := models.Score{ID: primitive.NewObjectID()}
	err := data.InsertScore(testScore)
	assert.NoError(t, err)

	HexID := testScore.ID.Hex()
	defer data.DeleteScore(HexID)

	clientAuthToken := getClientAuthToken(t)
	adminAuthToken := getAdminAuthToken(t)

	cases := []apiTestCase{
		apiTestCase{
			name:   "It should return Unauthorized",
			method: "GET",
			url:    "/scores",
			status: http.StatusUnauthorized,
		},
		apiTestCase{
			name:      "It should return Unauthorized since client cannot access this",
			method:    "GET",
			url:       "/scores",
			status:    http.StatusUnauthorized,
			authToken: clientAuthToken,
		},
		apiTestCase{
			name:      "It should return BadRequest since ID is not a valid ObjectId",
			method:    "GET",
			url:       "/scores/score/invalid-score-id",
			status:    http.StatusBadRequest,
			authToken: adminAuthToken,
		},
		apiTestCase{
			name:      "It should return OK since ID is the previously created test score",
			method:    "GET",
			url:       "/scores/score/" + HexID,
			status:    http.StatusOK,
			authToken: adminAuthToken,
		},
	}

	r.RunTests(t, cases)
}
