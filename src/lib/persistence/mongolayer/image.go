package mongolayer

import (
	"context"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertImage ...
func (m *MongoDAL) InsertImage(image models.Image) error {
	_, err := m.C(CollectionImages).InsertOne(context.Background(), image)
	return err
}

// FindImage ...
func (m *MongoDAL) FindImage(query persistence.Query) (*models.Image, error) {
	var result models.Image
	err := m.C(CollectionImages).FindOne(context.Background(), query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GetImage ...
func (m *MongoDAL) GetImage(id string, query persistence.Query) (*models.Image, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.FindImage(query.AddCondition("_id", ID))
}

// GetImages ...
func (m *MongoDAL) GetImages(query persistence.Query) ([]models.Image, error) {
	var result []models.Image
	var ctx = context.Background()
	cursor, err := m.C(CollectionImages).Find(ctx, query.GetConditions(), getFindOptions(query))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// GetMovieImages ...
func (m *MongoDAL) GetMovieImages(id string, query persistence.Query) ([]models.Image, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.GetImages(query.AddCondition("_id", ID))
}

// DeleteImage ...
func (m *MongoDAL) DeleteImage(id string) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.C(CollectionImages).DeleteOne(context.Background(), bson.M{"_id": ID})
	return err
}

// DeleteImages ...
func (m *MongoDAL) DeleteImages(query persistence.Query) (int64, error) {
	result, err := m.C(CollectionImages).DeleteMany(context.Background(), query.GetConditions())
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, err
}

// DeleteImagesByIDs ...
func (m *MongoDAL) DeleteImagesByIDs(ids []string) (int64, error) {
	return m.DeleteImages(m.DefaultQuery().AddCondition("_id", bson.M{"$in": ids}))
}

// UpdateImage ...
func (m *MongoDAL) UpdateImage(id string, mi models.Image) (int64, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, err
	}
	result, err := m.C(CollectionImages).UpdateOne(context.Background(), bson.M{"_id": ID}, bson.M{"$set": mi})
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

// BuildImageQuery converts a map of query string to mongolayer syntax for Image model
func (m *MongoDAL) BuildImageQuery(q map[string]string) persistence.Query {
	query := BuildQuery("", q)
	if len(q) > 0 {
		// TODO
	}
	return query
}
