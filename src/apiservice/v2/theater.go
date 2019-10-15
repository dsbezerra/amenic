package v2

import (
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
)

// TheaterService ...
type TheaterService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeTheaters ...
func (r *RESTService) ServeTheaters(rg *gin.RouterGroup) {
	s := &TheaterService{r.data, r.emitter}

	client := rg.Group("/theaters", rest.JWTAuth(nil))
	client.GET("", s.GetAll)
	client.GET("/theater/:id", s.Get)
	client.GET("/theater/:id/prices", s.GetPrices)
	client.GET("/theater/:id/sessions", s.GetSessions)

	admin := rg.Group("/theaters", rest.JWTAuth(&rest.Endpoint{AdminOnly: true}))
	admin.GET("/count", s.Count)
	admin.PUT("/theater/:id", s.Update)
	admin.DELETE("/theater/:id", s.Delete)
}

// Get gets the theater corresponding the requested ID.
func (s *TheaterService) Get(c *gin.Context) {
	theater, err := s.data.GetTheater(c.Param("id"), BuildTheaterQuery(s.data, c))
	apiutil.SendSuccessOrError(c, theater, err)
}

// GetAll gets all theaters.
func (s *TheaterService) GetAll(c *gin.Context) {
	theaters, err := s.data.GetTheaters(BuildTheaterQuery(s.data, c))
	apiutil.SendSuccessOrError(c, theaters, err)
}

// GetPrices gets theater prices.
func (s *TheaterService) GetPrices(c *gin.Context) {
	prices, err := s.data.GetPrices(
		BuildTheaterQuery(s.data, c).
			AddCondition("theaterId", c.Param("id")))
	apiutil.SendSuccessOrError(c, prices, err)
}

// GetSessions gets theater sessions.
func (s *TheaterService) GetSessions(c *gin.Context) {
	sessions, err := s.data.GetSessions(
		BuildSessionQuery(s.data, c).
			AddCondition("theaterId", c.Param("id")))
	apiutil.SendSuccessOrError(c, sessions, err)
}

// Update apply to Theater with the given ID the given body data
func (s *TheaterService) Update(c *gin.Context) {
	theater := models.Theater{}
	err := c.ShouldBindJSON(&theater)
	if err != nil {
		apiutil.SendBadRequest(c)
		return
	}
	_, err = s.data.UpdateTheater(c.Param("id"), theater)
	apiutil.SendSuccessOrError(c, theater, err)
}

// Delete the Theater with the given ID
func (s *TheaterService) Delete(c *gin.Context) {
	err := s.data.DeleteTheater(c.Param("id"))
	apiutil.SendSuccessOrError(c, 1, err)
}

// Count returns the total count of Theater matching the given query
func (s *TheaterService) Count(c *gin.Context) {
	count, err := s.data.CountTheaters(BuildTheaterQuery(s.data, c))
	apiutil.SendSuccessOrError(c, count, err)
}

// BuildTheaterQuery builds theater query from request query string
func BuildTheaterQuery(data persistence.DataAccessLayer, c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(map[string]string)
	return data.BuildTheaterQuery(query)
}
