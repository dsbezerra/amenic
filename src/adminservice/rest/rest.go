package rest

import (
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/gin-gonic/gin"
)

// Service TODO
type Service struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeAPI ...
func ServeAPI(r *gin.Engine, data persistence.DataAccessLayer, emitter messagequeue.EventEmitter) {
	s := &Service{data, emitter}

	// Apply default middlewares
	r.Use(rest.Init(), rest.AdminAuth(data))

	// AdminService routes.
	s.ServeCommands(r)
}
