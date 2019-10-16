package v2

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
	client := rg.Group("/notifications", rest.JWTAuth(nil))
	client.GET("/notification/:id", s.Get)

	// Apply AdminAuth only to /notifications
	admin := rg.Group("/notifications", rest.JWTAuth(&rest.Endpoint{AdminOnly: true}))
	admin.GET("", s.GetAll)
}

// Get gets the notification corresponding the requested ID.
func (s *NotificationService) Get(c *gin.Context) {
	notification, err := s.data.GetNotification(c.Param("id"), BuildNotificationQuery(s.data, c))
	apiutil.SendSuccessOrError(c, notification, err)
}

// GetAll gets all notifications.
func (s *NotificationService) GetAll(c *gin.Context) {
	notifications, err := s.data.GetNotifications(BuildNotificationQuery(s.data, c))
	apiutil.SendSuccessOrError(c, notifications, err)
}

// BuildNotificationQuery builds notification persistence.Query from request query string
func BuildNotificationQuery(data persistence.DataAccessLayer, c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(map[string]string)
	return data.BuildNotificationQuery(query)
}
