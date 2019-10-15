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

func TestNotification(t *testing.T) {
	data := NewMockDataAccessLayer()

	r := NewMockRouter(data)
	r.Use(rest.Init(), middlewares.ValidObjectIDHex(), middlewares.BaseParseQuery())

	testNotification := models.Notification{
		ID:    primitive.NewObjectID(),
		Title: "Test",
		Type:  "premiere",
	}
	err := data.InsertNotification(testNotification)
	assert.NoError(t, err)

	s := RESTService{data: data}
	s.ServeNotifications(&r.RouterGroup)

	HexID := testNotification.ID.Hex()

	clientAuthToken := getClientAuthToken(t)
	adminAuthToken := getAdminAuthToken(t)

	testCases := []apiTestCase{
		apiTestCase{
			name:   "It should return Unauthorized",
			method: "GET",
			url:    "/notifications/notification/5c353e8cebd54428b4f25447",
			status: http.StatusUnauthorized,
		},
		apiTestCase{
			name:      "It should return BadRequest since ID is not a valid ObjectId",
			method:    "GET",
			url:       "/notifications/notification/invalid-notification-id",
			status:    http.StatusBadRequest,
			authToken: clientAuthToken,
		},
		apiTestCase{
			name:      "It should return NotFound since Notification with ID 5c353e8cebd54428b4f25447 doesn't exist",
			method:    "GET",
			url:       "/notifications/notification/5c353e8cebd54428b4f25447",
			status:    http.StatusNotFound,
			authToken: clientAuthToken,
		},
		apiTestCase{
			name:      "It should return a Notification with ID " + HexID,
			method:    "GET",
			url:       "/notifications/notification/" + HexID,
			status:    http.StatusOK,
			authToken: clientAuthToken,
		},
		apiTestCase{
			name:      "It should return Unauthorized because endpoint can be accessed only by admins",
			method:    "GET",
			url:       "/notifications",
			status:    http.StatusUnauthorized,
			authToken: clientAuthToken,
		},
		apiTestCase{
			name:      "It should return OK because the token is a valid admin token",
			method:    "GET",
			url:       "/notifications",
			status:    http.StatusOK,
			authToken: adminAuthToken,
		},
	}
	r.RunTests(t, testCases)

	err = data.DeleteNotification(testNotification.ID.Hex())
	assert.NoError(t, err)
}
