package v2

import (
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
)

// PriceService ...
type PriceService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServePrices ...
func (r *RESTService) ServePrices(rg *gin.RouterGroup) {
	s := &PriceService{r.data, r.emitter}

	client := rg.Group("/prices", rest.JWTAuth(nil))
	client.GET("/price/:id", s.Get)

	admin := rg.Group("/prices", rest.JWTAuth(&rest.Endpoint{AdminOnly: true}))
	admin.GET("", s.GetAll)
}

// Get gets the price corresponding the requested ID.
func (s *PriceService) Get(c *gin.Context) {
	price, err := s.data.GetPrice(c.Param("id"), s.ParseQuery(c))
	apiutil.SendSuccessOrError(c, price, err)
}

// GetAll gets all prices.
func (s *PriceService) GetAll(c *gin.Context) {
	prices, err := s.data.GetPrices(s.ParseQuery(c))
	apiutil.SendSuccessOrError(c, prices, err)
}

// ParseQuery builds the conditional Mongo query
func (s *PriceService) ParseQuery(c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(persistence.Query)

	// This was added for back compat with app versions belo 1.0.22
	if cinemaID := c.Query("cinema"); cinemaID != "" {
		query.AddCondition("cinema_id", cinemaID)
	}

	if cinemaID := c.Query("cinema_id"); cinemaID != "" {
		query.AddCondition("cinema_id", cinemaID)
	}

	return query
}
