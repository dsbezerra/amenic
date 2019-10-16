package mongolayer

import (
	"context"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertNotification ...
func (m *MongoDAL) InsertNotification(notification models.Notification) error {
	_, err := m.C(CollectionNotifications).InsertOne(context.Background(), notification)
	return err
}

// FindNotification ...
func (m *MongoDAL) FindNotification(query persistence.Query) (*models.Notification, error) {
	var result models.Notification
	err := m.C(CollectionNotifications).FindOne(context.Background(), query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GetNotification ...
func (m *MongoDAL) GetNotification(id string, query persistence.Query) (*models.Notification, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.FindNotification(query.AddCondition("_id", ID))
}

// GetNotifications ...
func (m *MongoDAL) GetNotifications(query persistence.Query) ([]models.Notification, error) {
	// TODO: Implement aggregate if we have any include queries
	var result []models.Notification
	var ctx = context.Background()
	cursor, err := m.C(CollectionNotifications).Find(ctx, query.GetConditions(), getFindOptions(query))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// DeleteNotification ...
func (m *MongoDAL) DeleteNotification(id string) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.C(CollectionNotifications).DeleteOne(context.Background(), bson.M{"_id": ID})
	return err
}

// DeleteNotifications ...
func (m *MongoDAL) DeleteNotifications(query persistence.Query) (int64, error) {
	result, err := m.C(CollectionNotifications).DeleteMany(context.Background(), query.GetConditions())
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, err
}

// BuildNotificationQuery ...
func (m *MongoDAL) BuildNotificationQuery(q map[string]string) persistence.Query {
	query := BuildQuery("", q)
	if len(q) > 0 {

	}
	return query
}
