package v2

import (
	"net/http"
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	jwt_lib "github.com/dgrijalva/jwt-go"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtSecret = "somesecret"

func TestAuth(t *testing.T) {
	data := NewMockDataAccessLayer()

	r := NewMockRouter(data)
	r.Use(rest.Init())

	testAdmin := models.Admin{
		ID:       primitive.NewObjectID(),
		Username: "test",
	}

	adm, err := data.FindAdmin(data.DefaultQuery().AddCondition("username", testAdmin.Username))
	if adm != nil {
		data.DeleteAdmin(adm.ID.Hex())
	}

	pwd, err := bcrypt.GenerateFromPassword([]byte("my password"), 0)
	assert.NoError(t, err)

	testAdmin.Password = string(pwd)
	err = data.InsertAdmin(testAdmin)
	assert.NoError(t, err)

	// Create test APIKey
	testAPIKey := models.APIKey{
		Key:      "test-api-key",
		UserType: "admin",
		Owner:    adm.Username,
	}
	apikey, err := data.FindAPIKey(data.DefaultQuery().AddCondition("key", testAPIKey.Key))
	if apikey != nil {
		data.DeleteAPIKey(apikey.ID.Hex())
	}

	err = data.InsertAPIKey(testAPIKey)
	assert.NoError(t, err)

	os.Setenv("JWT_SECRET", jwtSecret)

	s := RESTService{data: data}
	s.ServeAuth(&r.RouterGroup)

	cases := []apiTestCase{
		apiTestCase{
			name:   "It should return Bad Request",
			method: "POST",
			url:    "/auth",
			status: http.StatusBadRequest,
			body:   "invalid body",
		},
		apiTestCase{
			name:   "It should return Unauthorized",
			method: "POST",
			url:    "/auth",
			status: http.StatusUnauthorized,
			body: `
				{
					"username": "test",
					"password": "wrong password"
				}
			`,
		},
		apiTestCase{
			name:   "It should return OK",
			method: "POST",
			url:    "/auth",
			status: http.StatusOK,
			body: `
				{
					"username": "test",
					"password": "my password"
				}
			`,
		},
	}

	r.RunTests(t, cases)
}

func getAuthToken(t *testing.T, tt string, scopes []string) string {
	claims := rest.NewClaims(&models.APIKey{
		UserType: tt,
		Platform: "some_platform",
	}, scopes)
	assert.NotNil(t, claims)

	claims.ExpiresAt = time.Now().Add(time.Hour * 1).Unix()
	// claims.Issuer ?

	token := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))
	token.Claims = claims

	ts, err := token.SignedString([]byte(rest.JWTSecret))
	assert.NoError(t, err)
	return ts
}

func getClientAuthToken(t *testing.T) string {
	return getAuthToken(t, "client", []string{"api_read"})
}

func getAdminAuthToken(t *testing.T) string {
	return getAuthToken(t, "admin", []string{"api_read", "api_write"})
}
