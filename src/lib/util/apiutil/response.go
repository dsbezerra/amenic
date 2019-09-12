package apiutil

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse TODO
type APIResponse struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  *APIError   `json:"error,omitempty"`
}

// SendSuccess is a helper for sending success responses.
func SendSuccess(c *gin.Context, data interface{}) {
	response := &APIResponse{Status: 200, Data: data}
	c.SecureJSON(response.Status, response)
}

// SendBadRequest is a helper for sending bad_request error responses.
func SendBadRequest(c *gin.Context) {
	c.SecureJSON(http.StatusBadRequest, &APIResponse{
		Status: http.StatusBadRequest,
		Error:  apiErrorBadRequest,
	})
}

// SendUnauthorized is a helper for sending unauthorized error response.
func SendUnauthorized(c *gin.Context) {
	c.SecureJSON(http.StatusUnauthorized, &APIResponse{
		Status: http.StatusUnauthorized,
		Error:  apiErrorUnauthorized,
	})
}

// SendSuccessOrError is a shortcut function for sending success/error responses.
func SendSuccessOrError(c *gin.Context, data interface{}, err error) {
	if err != nil {
		HandleError(c, err)
		return
	}
	SendSuccess(c, data)
}

// SendNotFound is a helper for sending not_found error responses.
func SendNotFound(c *gin.Context) {
	c.SecureJSON(http.StatusNotFound, &APIResponse{
		Status: http.StatusNotFound,
		Error:  apiErrorNotFound,
	})
}

// SendProtectedResource is a helper for sending protected resource error response.
func SendProtectedResource(c *gin.Context) {
	c.SecureJSON(http.StatusUnauthorized, &APIResponse{
		Status: http.StatusUnauthorized,
		Error:  apiErrorProtectedResource,
	})
}
