package mongolayer

import (
	"context"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertScore ...
func (m *MongoDAL) InsertScore(score models.Score) error {
	_, err := m.C(CollectionScores).InsertOne(context.Background(), score)
	return err
}

// FindScore ...
func (m *MongoDAL) FindScore(query persistence.Query) (*models.Score, error) {
	var result models.Score
	err := m.C(CollectionScores).FindOne(context.Background(), query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GetScore ...
func (m *MongoDAL) GetScore(id string, query persistence.Query) (*models.Score, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.FindScore(query.AddCondition("_id", ID))
}

// GetScores ...
func (m *MongoDAL) GetScores(query persistence.Query) ([]models.Score, error) {
	// TODO: Implement aggregate if we have any include queries
	var result []models.Score
	var ctx = context.Background()
	cursor, err := m.C(CollectionScores).Find(ctx, query.GetConditions(), getFindOptions(query))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// UpdateScore ...
func (m *MongoDAL) UpdateScore(id string, ms models.Score) (int64, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, err
	}
	ms.UpdatedAt = getCurrentTime()
	result, err := m.C(CollectionTheaters).UpdateOne(context.Background(), bson.M{"_id": ID}, bson.M{"$set": ms})
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

// DeleteScore ...
func (m *MongoDAL) DeleteScore(id string) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.C(CollectionScores).DeleteOne(context.Background(), bson.M{"_id": ID})
	return err
}

// DeleteScores ...
func (m *MongoDAL) DeleteScores(query persistence.Query) (int64, error) {
	result, err := m.C(CollectionScores).DeleteMany(context.Background(), query.GetConditions())
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, err
}

// BuildScoreQuery ...
func (m *MongoDAL) BuildScoreQuery(q map[string]string) persistence.Query {
	query := BuildQuery("", q)
	if len(q) > 0 {
		if movie, ok := q["movieId"]; ok {
			value, err := primitive.ObjectIDFromHex(movie)
			if err == nil {
				query.AddCondition("movieId", value)
			}
		}
	}
	return query
}
