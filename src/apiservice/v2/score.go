package v2

import (
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
)

// ScoreService ...
type ScoreService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeScores ...
func (r *RESTService) ServeScores(rg *gin.RouterGroup) {
	s := &ScoreService{r.data, r.emitter}

	scores := rg.Group("/scores", rest.JWTAuth(&rest.Endpoint{AdminOnly: true}))
	scores.GET("/", s.GetAll)
	scores.GET("/:id", s.Get)
}

// Get gets the score corresponding the requested ID.
func (s *ScoreService) Get(c *gin.Context) {
	score, err := s.data.GetScore(c.Param("id"), s.ParseQuery(c))
	apiutil.SendSuccessOrError(c, score, err)
}

// GetAll gets all scores.
func (s *ScoreService) GetAll(c *gin.Context) {
	scores, err := s.data.GetScores(s.ParseQuery(c))
	apiutil.SendSuccessOrError(c, scores, err)
}

// ParseQuery builds the conditional Mongo query
func (s *ScoreService) ParseQuery(c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(persistence.Query)
	// Custom query here
	if m := c.Query("movie"); m != "" {
		query.AddCondition("movieId", m)
	}
	return query
}
