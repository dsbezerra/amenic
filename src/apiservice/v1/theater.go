/** DEPRECATED */

package v1

import (
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/gin-gonic/gin"
)

type Cinema struct {
}

// TheaterService ...
type TheaterService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeTheaters ...
func (r *RESTService) ServeTheaters(rg *gin.RouterGroup) {
	s := &TheaterService{r.data, r.emitter}

	client := rg.Group("/cinemas", rest.ClientAuth(r.data))
	client.GET("/:id/showtimes", s.GetSessions)
}

// GetSessions gets theater sessions.
func (s *TheaterService) GetSessions(c *gin.Context) {
	getSessions(c, c.Param("id"), s.data, "{theater,movie{title\\nposter\\nrating}}")
}
