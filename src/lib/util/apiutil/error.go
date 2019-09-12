package apiutil

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// APIError ...
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (
	apiErrorNotFound          = NewAPIError("not_found", "Resource not found")
	apiErrorBadRequest        = NewAPIError("bad_request", "Request cannot be fulfilled due to bad syntax")
	apiErrorUnknown           = NewAPIError("unknown", "Unknown error occurred")
	apiErrorUnauthorized      = NewAPIError("unauthorized", "This action requires authentication")
	apiErrorProtectedResource = NewAPIError("protected", "This resource is protected and cannot be deleted or modified")
)

// HandleError main handler for errors in the API.
func HandleError(c *gin.Context, err error) {
	res := &APIResponse{Status: 200}

	switch err {
	case mongo.ErrNoDocuments:
		res.Status = http.StatusNotFound
		res.Error = apiErrorNotFound
	case err.(*strconv.NumError):
		res.Status = http.StatusBadRequest
		res.Error = apiErrorBadRequest
	default:
		res.Status = 500
		res.Error = NewAPIError("unknown", err.Error()) // @Temporary
	}

	c.SecureJSON(res.Status, res)
}

// NewAPIError create and allocates new APIError error instance.
func NewAPIError(code, message string) *APIError {
	return &APIError{Code: code, Message: message}
}
