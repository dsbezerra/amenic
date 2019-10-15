/** DEPRECATED */

package v1

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNotification(t *testing.T) {
	data := NewMockDataAccessLayer()
	r := NewMockRouter(data)
	r.Use(rest.Init())

	testNotification := models.Notification{
		ID:    primitive.NewObjectID(),
		Title: "Test",
		Type:  "premiere",
	}
	err := data.InsertNotification(testNotification)
	assert.NoError(t, err)

	s := RESTService{data: data}
	s.ServeNotifications(&r.RouterGroup)

	testCases := []apiTestCase{
		newAPITestCase("Get all notifications", "GET", "/notifications", "", http.StatusUnauthorized, false, nil),
		newAPITestCase("Get single notification with ID 1", "GET", "/notifications/notification/1", "", http.StatusBadRequest, true, nil),
		newAPITestCase("Get single notification with ID 5c353e8cebd54428b4f25447", "GET", "/notifications/notification/5c353e8cebd54428b4f25447", "", http.StatusNotFound, true, nil),
		newAPITestCase(
			fmt.Sprintf("Get single notification with ID %s", testNotification.ID.Hex()),
			"GET", fmt.Sprintf("/notifications/%s", testNotification.ID.Hex()), "", http.StatusOK, true, nil),
	}
	r.RunTests(t, testCases)

	err = data.DeleteNotification(testNotification.ID.Hex())
	assert.NoError(t, err)
}
