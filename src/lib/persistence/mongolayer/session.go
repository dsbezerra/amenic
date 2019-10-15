package mongolayer

import (
	"context"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertSession ...
func (m *MongoDAL) InsertSession(session models.Session) error {
	_, err := m.C(CollectionSessions).InsertOne(context.Background(), session)
	return err
}

// InsertSessions ...
func (m *MongoDAL) InsertSessions(sessions ...models.Session) error {
	arr := make([]interface{}, len(sessions))
	for i, p := range sessions {
		arr[i] = p
	}
	_, err := m.C(CollectionSessions).InsertMany(context.Background(), arr)
	return err
}

// FindSession ...
func (m *MongoDAL) FindSession(query persistence.Query) (*models.Session, error) {
	var result models.Session
	err := m.C(CollectionSessions).FindOne(context.Background(), query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GetSession ...
func (m *MongoDAL) GetSession(id string, query persistence.Query) (*models.Session, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.FindSession(query.AddCondition("_id", ID))
}

// GetSessions ...
func (m *MongoDAL) GetSessions(query persistence.Query) ([]models.Session, error) {
	var result []models.Session
	var ctx = context.Background()
	var cursor *mongo.Cursor
	var err error

	var C = m.C(CollectionSessions)
	if query.HasInclude() {
		opts := options.Aggregate()
		cursor, err = C.Aggregate(ctx, buildPipeline(CollectionSessions, query.(*QueryOptions)), opts)
	} else {
		cursor, err = C.Find(ctx, query.GetConditions(), getFindOptions(query))
	}

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// DeleteSession ...
func (m *MongoDAL) DeleteSession(id string) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.C(CollectionSessions).DeleteOne(context.Background(), bson.M{"_id": ID})
	return err
}

// DeleteSessions ...
func (m *MongoDAL) DeleteSessions(query persistence.Query) (int64, error) {
	result, err := m.C(CollectionSessions).DeleteMany(context.Background(), query.GetConditions())
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, err
}

// BuildSessionQuery ...
func (m *MongoDAL) BuildSessionQuery(q map[string]string) persistence.Query {
	query := BuildQuery("", q)
	if len(q) > 0 {

		if theater, ok := q["theaterId"]; ok {
			value, err := primitive.ObjectIDFromHex(theater)
			if err == nil {
				query.AddCondition("theaterId", value)
				query.SetLimit(-1)
			}
		}

		if movie, ok := q["movieId"]; ok {
			value, err := primitive.ObjectIDFromHex(movie)
			if err == nil {
				query.AddCondition("movieId", value)
			}
		}

	}

	return query
}
