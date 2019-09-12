package mongolayer

import (
	"testing"

	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTheater(t *testing.T) {
	data, err := getTestingMongoDAL()
	defer data.Close()
	assert.NoError(t, err)

	// Clear all
	opts := DefaultOptions(CollectionCities)
	_, err = data.DeleteCities(opts)
	assert.NoError(t, err)

	opts = DefaultOptions(CollectionTheaters)
	_, err = data.DeleteTheaters(opts)
	assert.NoError(t, err)

	// Count
	count, err := data.CountTheaters(opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Insert
	ID, _ := primitive.ObjectIDFromHex("5d545238dde6cd12e23afaaa")
	cityDoc := models.City{
		ID:   ID,
		Name: "some-city-name",
	}
	ID, _ = primitive.ObjectIDFromHex("5d5442bdc8c6dc53016b2a6d")
	doc := models.Theater{
		ID:        ID,
		CityID:    cityDoc.ID,
		Name:      "theater-name",
		ShortName: "theater-short-name",
	}
	err = data.InsertCity(cityDoc)
	assert.NoError(t, err)
	err = data.InsertTheater(doc)
	assert.NoError(t, err)

	// Find one by id
	theater, err := data.GetTheater(doc.ID.Hex(), opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, theater)

	// Find many
	theaters, err := data.GetTheaters(opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, theaters)

	// Find incuding city
	includingOpts := opts
	includingOpts.Includes = append(includingOpts.Includes, QueryInclude{
		Field:  "city",
		Fields: []string{"name"},
	})
	theater, err = data.GetTheater(doc.ID.Hex(), includingOpts)
	assert.NoError(t, err)
	assert.NotEmpty(t, theater)

	// Count
	count, err = data.CountTheaters(opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete
	err = data.DeleteTheater(doc.ID.Hex())
	assert.NoError(t, err)

	err = data.DeleteCity(cityDoc.ID.Hex())
	assert.NoError(t, err)

	// Count
	count, err = data.CountTheaters(opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
