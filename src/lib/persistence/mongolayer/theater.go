package mongolayer

import (
	"context"
	"strconv"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CountTheaters ...
func (m *MongoDAL) CountTheaters(query persistence.Query) (int64, error) {
	return m.C(CollectionTheaters).CountDocuments(context.Background(), query.GetConditions())
}

// InsertTheater ...
func (m *MongoDAL) InsertTheater(theater models.Theater) error {
	_, err := m.C(CollectionTheaters).InsertOne(context.Background(), theater)
	return err
}

// FindTheater ...
func (m *MongoDAL) FindTheater(query persistence.Query) (*models.Theater, error) {
	var result models.Theater

	var ctx = context.Background()
	var cursor *mongo.Cursor
	var err error

	var C = m.C(CollectionTheaters)
	if query.HasInclude() {
		cursor, err = C.Aggregate(ctx, buildPipeline(CollectionTheaters, query.(*QueryOptions)))
		if cursor != nil {
			defer cursor.Close(ctx)
			if cursor.Next(ctx) {
				err = cursor.Decode(&result)
			}
		}
	} else {
		err = C.FindOne(ctx, query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	}

	if err != nil {
		return nil, err
	}

	return &result, err
}

// GetTheater ...
func (m *MongoDAL) GetTheater(id string, query persistence.Query) (*models.Theater, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.FindTheater(query.AddCondition("_id", ID))
}

// GetTheaters ...
func (m *MongoDAL) GetTheaters(query persistence.Query) ([]models.Theater, error) {
	var result []models.Theater

	var ctx = context.Background()
	var cursor *mongo.Cursor
	var err error

	var C = m.C(CollectionTheaters)
	if query.HasInclude() {
		opts := options.Aggregate()
		cursor, err = C.Aggregate(ctx, buildPipeline(CollectionTheaters, query.(*QueryOptions)), opts)
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

// DeleteTheater ...
func (m *MongoDAL) DeleteTheater(id string) error {
	// TODO: Delete all prices, movies and images owned by this theater
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.C(CollectionTheaters).DeleteOne(context.Background(), bson.M{"_id": ID})
	return err
}

// DeleteTheaters ...
func (m *MongoDAL) DeleteTheaters(query persistence.Query) (int64, error) {
	result, err := m.C(CollectionTheaters).DeleteMany(context.Background(), query.GetConditions())
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, err
}

// UpdateTheater ...
func (m *MongoDAL) UpdateTheater(id string, mt models.Theater) (int64, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, err
	}
	mt.UpdatedAt = getCurrentTime()
	result, err := m.C(CollectionTheaters).UpdateOne(context.Background(), bson.M{"_id": ID}, bson.M{"$set": mt})
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

// BuildTheaterQuery converts a map of query string to mongolayer syntax for Theater model
func (m *MongoDAL) BuildTheaterQuery(q map[string]string) persistence.Query {
	query := BuildQuery("", q)
	if len(q) > 0 {
		ID, ok := q["id"]
		if ok {
			value, err := primitive.ObjectIDFromHex(ID)
			if err == nil {
				query.AddCondition("_id", value)
			}
		}
		ID, ok = q["internalId"]
		if ok {
			query.AddCondition("internalId", ID)
		}

		// Hidden
		hidden, ok := q["hidden"]
		if ok {
			value, err := strconv.ParseBool(hidden)
			if err == nil {
				query.AddCondition("hidden", value)
			}
		}

		search, ok := q["search"]
		if ok {
			query.AddCondition("$or", []bson.M{
				bson.M{"name": bson.M{"$regex": search, "$options": "ig"}},
				bson.M{"shortName": bson.M{"$regex": search, "$options": "ig"}},
			})
		}
	}
	return query
}
