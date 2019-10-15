package apiutil

import (
	"net/http"
	"strconv"

	jwt_lib "github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// APIError ...
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (
	apiErrorUnknown            = NewAPIError("unknown", "Unknown error occurred")
	apiErrorNotFound           = NewAPIError("not_found", "Resource not found")
	apiErrorBadRequest         = NewAPIError("bad_request", "Request cannot be fulfilled due to bad syntax")
	apiErrorUnauthorized       = NewAPIError("unauthorized", "This action requires authentication")
	apiInternalServerError     = NewAPIError("internal_server_error", "An internal server error occurred")
	apiErrorProtectedResource  = NewAPIError("protected", "This resource is protected and cannot be deleted or modified")
	apiErrorInvalidCredentials = NewAPIError("invalid_credentials", "Invalid credentials")
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
	case jwt_lib.ErrNoTokenInRequest:
		res.Status = http.StatusUnauthorized
		res.Error = apiErrorUnauthorized
	case ErrInvalidCredentials, bcrypt.ErrMismatchedHashAndPassword, bcrypt.ErrHashTooShort:
		res.Status = http.StatusUnauthorized
		res.Error = apiErrorInvalidCredentials
	default:
		res.Status = 500
		res.Error = NewAPIError("unknown", err.Error()) // @Temporary
	}

	c.SecureJSON(res.Status, res)
	c.Abort()
}

// NewAPIError create and allocates new APIError error instance.
func NewAPIError(code, message string) *APIError {
	return &APIError{Code: code, Message: message}
}
