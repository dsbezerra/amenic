package v1

import (
	"net/http"
	"testing"

	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBasicAuth(t *testing.T) {
	r := NewMockRouter(NewMockDataAccessLayer())
	r.Use(rest.Init())

	// Add test data
	apikey := models.APIKey{
		ID:       primitive.NewObjectID(),
		Key:      "test-api-key",
		UserType: "admin",
		Owner:    "test",
	}
	result, err := r.data.FindAPIKey(r.data.DefaultQuery().AddCondition("key", "test-api-key"))
	assert.NoError(t, err)
	if result != nil {
		r.data.DeleteAPIKey(result.ID.Hex())
	}
	err = r.data.InsertAPIKey(apikey)
	assert.NoError(t, err)

	// 200 expected without middleware
	r.GET("/not-authenticated", func(c *gin.Context) {
		c.String(http.StatusOK, "...")
	})

	r.Use(rest.AdminAuth(r.data))
	r.GET("/authenticated", func(c *gin.Context) {
		c.String(http.StatusOK, "...")
	})

	res := r.Call("GET", "/not-authenticated", "")
	assert.Equal(t, http.StatusOK, res.Code)

	res = r.Call("GET", "/authenticated", "")
	assert.Equal(t, http.StatusUnauthorized, res.Code)

	// NOTE: This shouldn't work since we passed a invalid api_key
	res = r.Call("GET", "/authenticated?api_key=aksdjalksdjaskdaj", "")
	assert.Equal(t, http.StatusUnauthorized, res.Code)

	// NOTE: This should work since we passed the fake api_key
	res = r.Call("GET", "/authenticated?api_key=test-api-key", "")
	assert.Equal(t, http.StatusOK, res.Code)

	err = r.data.DeleteAPIKey(apikey.ID.Hex())
	assert.NoError(t, err)
}
