package v2

import (
	"net/http"
	"testing"

	"github.com/dsbezerra/amenic/src/lib/middlewares"

	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
)

func TestState(t *testing.T) {
	data := NewMockDataAccessLayer()

	r := NewMockRouter(data)
	r.Use(rest.Init(), middlewares.BaseParseQuery())

	s := RESTService{data: data}
	s.ServeStates(&r.RouterGroup)

	clientAuthToken := getClientAuthToken(t)
	adminAuthToken := getAdminAuthToken(t)

	// State routes can be accessed only by admins
	cases := []apiTestCase{
		apiTestCase{
			name:   "It should return Unauthorized",
			method: "GET",
			url:    "/states",
			status: http.StatusUnauthorized,
		},
		apiTestCase{
			name:      "It should return Unauthorized because client token is not allowed to access",
			method:    "GET",
			url:       "/states",
			status:    http.StatusUnauthorized,
			authToken: clientAuthToken,
		},
		apiTestCase{
			name:      "It should return NotFound because state doesn't exist",
			method:    "GET",
			url:       "/states/invalid-state-id",
			status:    http.StatusNotFound,
			authToken: adminAuthToken,
		},
		apiTestCase{
			name:      "It should return OK",
			method:    "GET",
			url:       "/states/state/MG",
			status:    http.StatusOK,
			authToken: adminAuthToken,
		},
		apiTestCase{
			name:      "It should return OK",
			method:    "GET",
			url:       "/states/state/MG/cities",
			status:    http.StatusOK,
			authToken: adminAuthToken,
		},
	}

	r.RunTests(t, cases)
}
