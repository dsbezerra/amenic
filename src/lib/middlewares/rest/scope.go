package rest

import (
	"time"

	"github.com/gin-gonic/gin"
)

// RequestScope contains information that are carried around in a request.
type RequestScope interface {
	// UserCredentials returns the user credentials for request
	UserCredentials() Credentials

	// SetUserCredentials sets the user credentials for request
	SetUserCredentials(cred Credentials)

	// Now returns the timestamp representing the time when the request is being processed
	Now() time.Time
}

type requestScope struct {
	now  time.Time
	cred Credentials
}

// Init returns a middleware that prepares the request scope
func Init() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("RequestScope", NewRequestScope(time.Now()))
	}
}

// GetRequestScope returns the RequestScope of the current request.
func GetRequestScope(c *gin.Context) RequestScope {
	return c.MustGet("RequestScope").(RequestScope)
}

// NewRequestScope TODO
func NewRequestScope(now time.Time) RequestScope {
	return &requestScope{
		now: now,
	}
}

func (rs *requestScope) UserCredentials() Credentials {
	return rs.cred
}

func (rs *requestScope) SetUserCredentials(cred Credentials) {
	rs.cred = cred
}

func (rs *requestScope) Now() time.Time {
	return rs.now
}
