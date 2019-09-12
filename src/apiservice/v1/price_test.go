package v1

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPrice(t *testing.T) {
	data := NewMockDataAccessLayer()
	r := NewMockRouter(data)
	r.Use(rest.Init())

	testPrice := models.Price{
		ID:        primitive.NewObjectID(),
		TheaterID: primitive.NewObjectID(),
		Amount:    10,
		Weekdays:  []models.Weekday{models.SATURDAY},
	}
	err := data.InsertPrice(testPrice)
	if err != nil {
		log.Fatal(err)
	}

	s := RESTService{data: data}
	s.ServePrices(&r.RouterGroup)

	testCases := []apiTestCase{
		newAPITestCase("Get all prices", "GET", "/prices", "", http.StatusUnauthorized, false, nil),
		newAPITestCase("Get single price with ID 1", "GET", "/prices/price/1", "", http.StatusBadRequest, true, nil),
		newAPITestCase("Get single price with ID 5c353e8cebd54428b4f25447", "GET", "/prices/price/5c353e8cebd54428b4f25447", "", http.StatusNotFound, true, nil),
		newAPITestCase(
			fmt.Sprintf("Get single price with ID %s", testPrice.ID.Hex()),
			"GET", fmt.Sprintf("/prices/%s", testPrice.ID.Hex()), "", http.StatusOK, true, nil),
	}
	r.RunTests(t, testCases)

	err = data.DeletePrice(testPrice.ID.Hex())
	assert.NoError(t, err)
}
