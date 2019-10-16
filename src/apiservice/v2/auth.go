package v2

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"golang.org/x/crypto/bcrypt"

	jwt_lib "github.com/dgrijalva/jwt-go"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
)

// AuthService ...
type AuthService struct {
	data persistence.DataAccessLayer
}

type LoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

// ServeAuth ...
func (r *RESTService) ServeAuth(rg *gin.RouterGroup) {
	s := &AuthService{r.data}
	auth := rg.Group("/auth")
	auth.Use(rest.EnsureJWTSecret)
	auth.POST("", s.Login)
	auth.GET("/request_token", s.RequestToken)
}

// Login is used to generate a new admin JWT token 7 days expiry time.
func (r *AuthService) Login(c *gin.Context) {
	var body LoginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apiutil.SendBadRequest(c)
		return
	}

	admin, err := r.data.FindAdmin(r.data.DefaultQuery().AddCondition("username", body.Username))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = apiutil.ErrInvalidCredentials
		}
		apiutil.SendSuccessOrError(c, admin, err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(body.Password))
	if err != nil {
		apiutil.SendSuccessOrError(c, nil, err)
		return
	}

	apikey, err := r.data.FindAPIKey(r.data.DefaultQuery().
		AddCondition("owner", body.Username).
		AddCondition("user_type", "admin"))
	if err != nil {
		apiutil.SendUnauthorized(c)
		return
	}

	claims := rest.NewClaims(apikey, []string{"api_read", "api_write"})
	if claims == nil {
		apiutil.SendUnauthorized(c)
		return
	}

	claims.ExpiresAt = time.Now().AddDate(0, 0, 7).Unix()
	// claims.Issuer ?

	token := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))
	token.Claims = claims

	ts, err := token.SignedString([]byte(rest.JWTSecret))
	apiutil.SendSuccessOrError(c, AuthResponse{ts}, nil)
}

// RequestToken generates a new client JWT token with 1 hour expiry time with permission to read.
func (r *AuthService) RequestToken(c *gin.Context) {
	apikey, err := r.data.FindAPIKey(r.data.DefaultQuery().
		AddCondition("platform", retrievePlatform(c)).
		AddCondition("user_type", "client"))
	if apikey == nil || err != nil {
		apiutil.SendUnauthorized(c)
		return
	}

	claims := rest.NewClaims(apikey, []string{"api_read"})
	if claims == nil {
		apiutil.SendUnauthorized(c)
		return
	}

	claims.ExpiresAt = time.Now().Add(time.Hour * 1).Unix()
	// claims.Issuer ?

	token := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))
	token.Claims = claims

	ts, err := token.SignedString([]byte(rest.JWTSecret))
	apiutil.SendSuccessOrError(c, AuthResponse{ts}, err)
}

func retrievePlatform(c *gin.Context) string {
	lower := strings.ToLower(c.GetHeader("user-agent"))
	if strings.Contains(lower, "android") {
		return "android"
	} else if strings.Contains(lower, "ios") {
		return "ios"
	} else if strings.Contains(lower, "web") {
		return "web"
	}
	return ""
}
