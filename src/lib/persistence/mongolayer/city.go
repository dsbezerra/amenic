package mongolayer

import (
	"context"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertCity ...
func (m *MongoDAL) InsertCity(city models.City) error {
	_, err := m.C(CollectionCities).InsertOne(context.Background(), city)
	return err
}

// InsertCities ...
func (m *MongoDAL) InsertCities(cities ...models.City) error {
	arr := make([]interface{}, len(cities))
	for i, p := range cities {
		arr[i] = p
	}
	_, err := m.C(CollectionCities).InsertMany(context.Background(), arr)
	return err
}

// FindCity ...
func (m *MongoDAL) FindCity(query persistence.Query) (*models.City, error) {
	var result models.City
	err := m.C(CollectionCities).FindOne(context.Background(), query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GetCity ...
func (m *MongoDAL) GetCity(id string, query persistence.Query) (*models.City, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.FindCity(query.AddCondition("_id", ID))
}

// GetCities ...
func (m *MongoDAL) GetCities(query persistence.Query) ([]models.City, error) {
	var result []models.City
	var ctx = context.Background()
	cursor, err := m.C(CollectionCities).Find(ctx, query.GetConditions(), getFindOptions(query))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// UpdateCity ...
func (m *MongoDAL) UpdateCity(id string, mc models.City) (int64, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, err
	}
	mc.UpdatedAt = getCurrentTime()
	result, err := m.C(CollectionCities).UpdateOne(context.Background(), bson.M{"_id": ID}, bson.M{"$set": mc})
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

// DeleteCity ...
func (m *MongoDAL) DeleteCity(id string) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.C(CollectionCities).DeleteOne(context.Background(), bson.M{"_id": ID})
	return err
}

// DeleteCities ...
func (m *MongoDAL) DeleteCities(query persistence.Query) (int64, error) {
	result, err := m.C(CollectionCities).DeleteMany(context.Background(), query.GetConditions())
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, err
}

// BuildCityQuery converts a map of query string to mongolayer syntax for City model
func (m *MongoDAL) BuildCityQuery(q map[string]string) persistence.Query {
	query := BuildQuery("", q)
	if len(q) > 0 {
		ID, ok := q["id"]
		if ok {
			value, err := primitive.ObjectIDFromHex(ID)
			if err == nil {
				query.AddCondition("_id", value)
			}
		}
		name, ok := q["name"]
		if ok {
			query.AddCondition("name", name)
		}
		state, ok := q["state"]
		if ok {
			query.AddCondition("state", state)
		}
	}
	return query
}
