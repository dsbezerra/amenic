/** DEPRECATED */

package v1

import (
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/mongolayer"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// APIKeyService ...
type APIKeyService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeAPIKeys ...
func (r *RESTService) ServeAPIKeys(rg *gin.RouterGroup) {
	s := &APIKeyService{r.data, r.emitter}

	// Apply AdminAuth only to /apikeys path
	apikeys := rg.Group("/apikeys", rest.AdminAuth(r.data))
	apikeys.GET("/", s.GetAll)
	apikeys.GET("/:id", s.Get)
}

// Get returns the APIKey with the specified ID
func (s *APIKeyService) Get(c *gin.Context) {
	opts := s.ParseQuery(c)
	permitted := checkPermissions(c, opts)
	if !permitted {
		apiutil.SendUnauthorized(c)
		return
	}
	apiKey, err := s.data.GetAPIKey(c.Param("id"), opts)
	apiutil.SendSuccessOrError(c, apiKey, err)
}

// GetAll returns the APIKey matching the query options
func (s *APIKeyService) GetAll(c *gin.Context) {
	opts := s.ParseQuery(c)
	permitted := checkPermissions(c, opts)
	if !permitted {
		apiutil.SendUnauthorized(c)
		return
	}
	apiKeys, err := s.data.GetAPIKeys(opts)
	apiutil.SendSuccessOrError(c, apiKeys, err)
}

// ParseQuery builds the conditional Mongo query
func (s *APIKeyService) ParseQuery(c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(persistence.Query)

	if owner := c.Query("owner"); owner != "" {
		query.AddCondition("owner", owner)
	}

	if userType := c.Query("user_type"); userType != "" {
		query.AddCondition("user_type", userType)
	}

	return query
}

// DefaultOptions returns the default options used when querying the collection
func (s *APIKeyService) defaultOptions() *mongolayer.QueryOptions {
	return &mongolayer.QueryOptions{
		Conditions: bson.M{},
		Fields:     bson.M{"_id": 1, "key": 1, "owner": 1, "name": 1, "user_type": 1, "iat": 1},
	}
}

func checkPermissions(c *gin.Context, query persistence.Query) bool {
	rs := rest.GetRequestScope(c)

	// UserCredentials is always valid at this point
	rowner := rs.UserCredentials().ID
	// Only retrieve apiKeys owned by the user
	qowner := query.GetCondition("owner")
	if qowner != "" && qowner != rowner {
		return false
	}
	// Update our options to reflect the owner.
	if qowner == "" {
		query.AddCondition("owner", rowner)
	}
	return true
}
