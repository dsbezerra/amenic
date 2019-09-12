package mongolayer

import (
	"testing"

	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNotification(t *testing.T) {
	data, err := getTestingMongoDAL()
	defer data.Close()
	assert.NoError(t, err)

	opts := DefaultOptions("")
	_, err = data.DeleteNotifications(opts)
	assert.NoError(t, err)

	// Insert
	doc := models.Notification{
		ID:   primitive.NewObjectID(),
		Type: "some-type",
	}
	err = data.InsertNotification(doc)
	assert.NoError(t, err)

	// Find one with conditions
	notification, err := data.FindNotification(opts.AddCondition("type", doc.Type))
	assert.NoError(t, err)
	assert.NotEmpty(t, notification)

	// Find one by id
	notification, err = data.GetNotification(doc.ID.Hex(), opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, notification)

	// Find many
	notifications, err := data.GetNotifications(opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, notifications)

	// Delete
	err = data.DeleteNotification(doc.ID.Hex())
	assert.NoError(t, err)
}
