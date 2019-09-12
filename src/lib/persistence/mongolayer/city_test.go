package mongolayer

import (
	"testing"

	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCity(t *testing.T) {
	data, err := getTestingMongoDAL()
	defer data.Close()
	assert.NoError(t, err)

	opts := DefaultOptions(CollectionCities)
	// Clear
	_, err = data.DeleteCities(opts)
	assert.NoError(t, err)

	// Insert
	cityDoc := models.City{
		ID:    primitive.NewObjectID(),
		Name:  "some-city-name",
		State: models.MG,
	}
	err = data.InsertCity(cityDoc)
	assert.NoError(t, err)

	// Get cities including state
	cities, err := data.GetCities(opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, cities)
}
