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

func TestSession(t *testing.T) {
	data := NewMockDataAccessLayer()

	r := NewMockRouter(data)
	r.Use(rest.Init(), middlewares.ValidObjectIDHex(), middlewares.BaseParseQuery())

	testSession := models.Session{
		ID: primitive.NewObjectID(),
	}
	err := data.InsertSession(testSession)
	assert.NoError(t, err)

	s := RESTService{data: data}
	s.ServeSessions(&r.RouterGroup)

	HexID := testSession.ID.Hex()

	cases := []apiTestCase{
		apiTestCase{
			name:         "It should return Unauthorized",
			method:       "GET",
			url:          "/sessions",
			status:       http.StatusUnauthorized,
			appendAPIKey: false,
		},
		apiTestCase{
			name:         "It should return BadRequest since ID is not a valid ObjectId",
			method:       "GET",
			url:          "/sessions/session/invalid-session-id",
			status:       http.StatusBadRequest,
			appendAPIKey: true,
		},
		apiTestCase{
			name:         "It should return NotFound since Session with ID 5c353e8cebd54428b4f25447 doesn't exist",
			method:       "GET",
			url:          "/sessions/session/5c353e8cebd54428b4f25447",
			status:       http.StatusNotFound,
			appendAPIKey: true,
		},
		apiTestCase{
			name:         "It should return a Session with ID " + HexID,
			method:       "GET",
			url:          "/sessions/session/" + HexID,
			status:       http.StatusOK,
			appendAPIKey: true,
		},
		apiTestCase{
			name:         "It should return a list of Session models",
			method:       "GET",
			url:          "/sessions",
			status:       http.StatusOK,
			appendAPIKey: true,
		},
	}

	r.RunTests(t, cases)

	err = data.DeleteSession(testSession.ID.Hex())
	assert.NoError(t, err)
}
