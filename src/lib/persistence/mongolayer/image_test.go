package mongolayer

import (
	"testing"

	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestImage(t *testing.T) {
	data, err := getTestingMongoDAL()
	defer data.Close()
	assert.NoError(t, err)

	doc := models.Image{
		ID:   primitive.NewObjectID(),
		Name: "some-image-name",
	}

	// Insert
	err = data.InsertImage(doc)
	assert.NoError(t, err)

	// Find one with conditions
	image, err := data.FindImage(DefaultOptions("").AddCondition("name", doc.Name))
	assert.NoError(t, err)
	assert.NotEmpty(t, image)

	// Find one by id
	image, err = data.GetImage(doc.ID.Hex(), DefaultOptions(""))
	assert.NoError(t, err)
	assert.NotEmpty(t, image)

	// Find many
	images, err := data.GetImages(DefaultOptions(""))
	assert.NoError(t, err)
	assert.NotEmpty(t, images)

	// Update
	update := doc
	update.Name = "updated-name"
	modified, err := data.UpdateImage(doc.ID.Hex(), update)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), modified)

	image, err = data.GetImage(doc.ID.Hex(), DefaultOptions(""))
	assert.NoError(t, err)
	assert.Equal(t, *image, update)

	// Delete
	err = data.DeleteImage(doc.ID.Hex())
	assert.NoError(t, err)
}
