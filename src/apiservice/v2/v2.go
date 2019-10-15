package v2

import (
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/gin-gonic/gin"
)

type (
	// RESTService TODO
	RESTService struct {
		data    persistence.DataAccessLayer
		emitter messagequeue.EventEmitter
	}
)

// AddRoutes add V2 routes to main router in group v2
func AddRoutes(r *gin.Engine, data persistence.DataAccessLayer, emitter messagequeue.EventEmitter) {
	r.Use(middlewares.BaseParseQuery())
	v2 := r.Group("v2")
	s := RESTService{data, emitter}

	s.ServeAuth(v2)
	s.ServeSchedules(v2)
	s.ServeCities(v2)
	s.ServeStates(v2)
	s.ServeTheaters(v2)
	s.ServeScores(v2)
	s.ServePrices(v2)
	s.ServeMovies(v2)
	s.ServeNotifications(v2)
}
