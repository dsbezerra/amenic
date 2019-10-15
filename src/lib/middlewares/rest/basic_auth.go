package rest

import (
	"sync"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/persistence/mongolayer"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// Credentials ...
type Credentials struct {
	ID  string // Holds username
	Key string // Holds API Key
}

// APIKeysTable keeps API keys in memory to avoid querying database everytime.
type APIKeysTable struct {
	sync.RWMutex
	m map[string]models.APIKey
}

var apiKeysTable = &APIKeysTable{
	m: make(map[string]models.APIKey),
}

// ClientAuth is a short function for BasicAuth(UserTypeClient)
// must be used to authenticate any route used by public client
// applications
func ClientAuth(data persistence.DataAccessLayer) gin.HandlerFunc {
	return BasicAuth(data, models.UserTypeClient)
}

// AdminAuth is a short function for BasicAuth(UserTypeAdmin)
// must be used to authenticate any private route that is used
// by internal apps
func AdminAuth(data persistence.DataAccessLayer) gin.HandlerFunc {
	return BasicAuth(data, models.UserTypeAdmin)
}

// BasicAuth is a simple middleware that checks if the query contains our
// api key and see if it's valid or not.
func BasicAuth(data persistence.DataAccessLayer, minUserType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.Query("api_key")
		if apiKey == "" {
			apiutil.SendUnauthorized(c)
			return
		}

		result, err := getAPIKey(data, apiKey)
		if err != nil {
			apiutil.SendUnauthorized(c)
			return
		}

		if result.UserTypeLevel() < models.UserTypeLevel(minUserType) {
			apiutil.SendUnauthorized(c)
			return
		}

		rs := GetRequestScope(c)
		rs.SetUserCredentials(Credentials{
			ID:  result.Owner,
			Key: result.Key,
		})

		c.Next()
	}
}

func getAPIKey(data persistence.DataAccessLayer, key string) (*models.APIKey, error) {
	c := apiKeysTable.get(key)
	if c.Key != "" {
		return &c, nil
	}

	opts := &mongolayer.QueryOptions{Conditions: bson.M{"key": key}}
	result, err := data.FindAPIKey(opts)
	if err != nil {
		return nil, err
	}

	if result.Key != "" {
		apiKeysTable.put(result)
	}

	return result, err
}

func (t *APIKeysTable) put(apiKey *models.APIKey) {
	apiKeysTable.Lock()
	apiKeysTable.m[apiKey.Key] = *apiKey
	apiKeysTable.Unlock()
}

func (t *APIKeysTable) get(key string) models.APIKey {
	result := models.APIKey{}

	apiKeysTable.RLock()
	result = apiKeysTable.m[key]
	apiKeysTable.RUnlock()

	return result
}
