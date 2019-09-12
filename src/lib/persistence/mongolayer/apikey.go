package mongolayer

import (
	"context"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertAPIKey ...
func (m *MongoDAL) InsertAPIKey(apikey models.APIKey) error {
	if apikey.Timestamp == nil {
		apikey.Timestamp = getCurrentTime()
	}
	_, err := m.C(CollectionAPIKeys).InsertOne(context.Background(), apikey)
	return err
}

// FindAPIKey ...
func (m *MongoDAL) FindAPIKey(query persistence.Query) (*models.APIKey, error) {
	var result models.APIKey
	err := m.C(CollectionAPIKeys).FindOne(context.Background(), query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GetAPIKey ...
func (m *MongoDAL) GetAPIKey(id string, query persistence.Query) (*models.APIKey, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.FindAPIKey(query.AddCondition("_id", ID))
}

// GetAPIKeys ...
func (m *MongoDAL) GetAPIKeys(query persistence.Query) ([]models.APIKey, error) {
	var result []models.APIKey
	var ctx = context.Background()
	cursor, err := m.C(CollectionAPIKeys).Find(ctx, query.GetConditions(), getFindOptions(query))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// DeleteAPIKey ...
func (m *MongoDAL) DeleteAPIKey(id string) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.C(CollectionAPIKeys).DeleteOne(context.Background(), bson.M{"_id": ID})
	return err
}
