package v2

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/mongolayer"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const (
	testConnection = "mongodb://localhost/amenic-test"
)

type apiTestCase struct {
	name       string
	method     string
	url        string
	body       string
	authToken  string
	status     int
	onResponse func(r *httptest.ResponseRecorder)
}

type mockRouter struct {
	*gin.Engine
	data persistence.DataAccessLayer
}

func (r *mockRouter) Call(method, URL, body, authToken string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, URL, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	return res
}

func (r *mockRouter) RunTests(t *testing.T, tests []apiTestCase) {
	for _, test := range tests {
		res := r.Call(test.method, test.url, test.body, test.authToken)
		assert.Equal(t, test.status, res.Code, test.name)
		if res != nil && test.onResponse != nil {
			test.onResponse(res)
		}
	}
}

func newAPITestCase(name, method, url, body string, status int, authToken string, onResponse func(*httptest.ResponseRecorder)) apiTestCase {
	return apiTestCase{
		name,
		method,
		url,
		body,
		authToken,
		status,
		onResponse,
	}
}

func NewMockRouter(data persistence.DataAccessLayer) *mockRouter {
	return &mockRouter{Engine: gin.New(), data: data}
}

func NewMockDataAccessLayer() persistence.DataAccessLayer {
	data, err := mongolayer.NewMongoDAL(testConnection)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func ConvertAPIResponse(r *httptest.ResponseRecorder, result interface{}) {
	var response apiutil.APIResponse
	err := json.Unmarshal(r.Body.Bytes(), &response)
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.Marshal(response.Data)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Fatal(err)
	}
}
