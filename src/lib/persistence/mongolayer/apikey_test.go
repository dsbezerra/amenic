package mongolayer

import (
	"testing"

	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAPIKey(t *testing.T) {
	data, err := getTestingMongoDAL()
	defer data.Close()
	assert.NoError(t, err)

	doc := models.APIKey{
		ID:       primitive.NewObjectID(),
		Key:      "key-value",
		Owner:    "owner-name",
		Name:     "key-name",
		UserType: "client",
	}

	// Insert
	err = data.InsertAPIKey(doc)
	assert.NoError(t, err)

	// Find one with conditions
	apikey, err := data.FindAPIKey(DefaultOptions("").AddCondition("key", doc.Key))
	assert.NoError(t, err)
	assert.NotEmpty(t, apikey)

	// Find one by id
	apikey, err = data.GetAPIKey(doc.ID.Hex(), DefaultOptions(""))
	assert.NoError(t, err)
	assert.NotEmpty(t, apikey)

	// Find many
	apikeys, err := data.GetAPIKeys(DefaultOptions(""))
	assert.NoError(t, err)
	assert.NotEmpty(t, apikeys)

	// Delete
	err = data.DeleteAPIKey(doc.ID.Hex())
	assert.NoError(t, err)
}
