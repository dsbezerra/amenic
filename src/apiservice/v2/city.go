package v2

import (
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
)

// CityService ...
type CityService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeCities ...
func (r *RESTService) ServeCities(rg *gin.RouterGroup) {
	s := &CityService{r.data, r.emitter}

	client := rg.Group("/cities", rest.AdminAuth(r.data))
	client.GET("/", s.GetAll)
	client.GET("/city/:id", s.Get)
}

// Get gets the city corresponding the requested ID.
func (s *CityService) Get(c *gin.Context) {
	city, err := s.data.GetCity(c.Param("id"), BuildCityQuery(s.data, c))
	apiutil.SendSuccessOrError(c, city, err)
}

// GetAll gets all cities.
func (s *CityService) GetAll(c *gin.Context) {
	cities, err := s.data.GetCities(BuildCityQuery(s.data, c))
	apiutil.SendSuccessOrError(c, cities, err)
}

// BuildCityQuery builds City query from request query string
func BuildCityQuery(data persistence.DataAccessLayer, c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(map[string]string)
	return data.BuildCityQuery(query)
}
