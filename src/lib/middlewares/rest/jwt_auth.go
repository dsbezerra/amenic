package rest

import (
	"os"

	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"

	"github.com/dgrijalva/jwt-go"
	jwt_lib "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
)

// JWTSecret ...
var JWTSecret string

// Claims ...
type Claims struct {
	Key      string   `json:"key"`      // Not using, forever?
	Platform string   `json:"platform"` // Not using for now.
	Type     string   `json:"type"`
	Scopes   []string `json:"scopes"`
	jwt.StandardClaims
}

// Endpoint ?
type Endpoint struct {
	AdminOnly bool
}

// NewClaims ...
func NewClaims(apikey *models.APIKey, scopes []string) *Claims {
	if apikey == nil || len(scopes) == 0 {
		return nil
	}

	// scopes must be either api_read or api_write or both
	valid := true
	for _, s := range scopes {
		if s != "api_read" && s != "api_write" {
			valid = false
			break
		}
	}

	if !valid {
		return nil
	}

	return &Claims{
		Key:      apikey.Key,
		Platform: apikey.Platform,
		Type:     apikey.UserType,
		Scopes:   scopes,
	}
}

// HasScope ...
func (c *Claims) HasScope(scope string) bool {
	result := false
	for _, s := range c.Scopes {
		if s == scope {
			result = true
			break
		}
	}
	return result
}

// EnsureJWTSecret ...
func EnsureJWTSecret(c *gin.Context) {
	if JWTSecret == "" {
		JWTSecret = os.Getenv("JWT_SECRET")
		if JWTSecret == "" {
			apiutil.SendInternalServerError(c)
			return
		}
	}
	c.Next()
}

// JWTAuth ...
func JWTAuth(endpoint *Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := request.ParseFromRequestWithClaims(c.Request, request.OAuth2Extractor, &Claims{}, func(token *jwt_lib.Token) (interface{}, error) {
			return []byte(JWTSecret), nil
		})

		authorized := true
		if err != nil {
			// We don't care about the error.
			authorized = false
		} else {
			if claims, ok := token.Claims.(*Claims); ok && token.Valid {
				if endpoint != nil && endpoint.AdminOnly {
					authorized = claims.Type == models.UserTypeAdmin && claims.HasScope("api_write")
				} else {
					switch c.Request.Method {
					case "GET":
						authorized = claims.HasScope("api_read")
					case "POST", "PUT", "DELETE":
						authorized = claims.HasScope("api_write")
					}
				}
			}
		}

		if !authorized {
			apiutil.SendUnauthorized(c)
		}
	}
}
