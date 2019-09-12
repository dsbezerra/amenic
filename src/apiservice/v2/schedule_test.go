package v2

import (
	"net/http"
	"testing"

	"github.com/dsbezerra/amenic/src/lib/middlewares"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
)

func TestSchedule(t *testing.T) {
	data := NewMockDataAccessLayer()
	r := NewMockRouter(data)
	r.Use(rest.Init(), middlewares.ValidObjectIDHex(), middlewares.BaseParseQuery())

	s := RESTService{data: data}
	s.ServeSchedules(&r.RouterGroup)

	cases := []apiTestCase{
		apiTestCase{
			name:         "It should return schedules",
			method:       "GET",
			url:          "/schedules?date=2019-08-22&theaterId=5d4e00db1b3e2d231434d147",
			status:       http.StatusOK,
			appendAPIKey: true,
		},
	}

	r.RunTests(t, cases)
}
