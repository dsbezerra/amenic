/** DEPRECATED */

package v1

import (
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
)

// NotificationService ...
type NotificationService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeNotifications ...
func (r *RESTService) ServeNotifications(rg *gin.RouterGroup) {
	s := &NotificationService{r.data, r.emitter}

	// Apply ClientAuth only to /notifications/:id path
	client := rg.Group("/notifications", rest.ClientAuth(r.data))
	client.GET("/notification/:id", s.Get)
	// Apply AdminAuth only to /notifications
	admin := rg.Group("/notifications", rest.AdminAuth(r.data))
	admin.GET("", s.GetAll)
}

// Get gets the notification corresponding the requested ID.
func (s *NotificationService) Get(c *gin.Context) {
	notification, err := s.data.GetNotification(c.Param("id"), s.ParseQuery(c))
	apiutil.SendSuccessOrError(c, notification, err)
}

// GetAll gets all notifications.
func (s *NotificationService) GetAll(c *gin.Context) {
	notifications, err := s.data.GetNotifications(s.ParseQuery(c))
	apiutil.SendSuccessOrError(c, notifications, err)
}

// ParseQuery builds the conditional Mongo query
func (s *NotificationService) ParseQuery(c *gin.Context) persistence.Query {
	return c.MustGet("query_options").(persistence.Query)
}
