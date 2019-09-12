package mongolayer

import (
	"context"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertPrice ...
func (m *MongoDAL) InsertPrice(price models.Price) error {
	if price.ID.IsZero() {
		price.ID = primitive.NewObjectID()
	}
	_, err := m.C(CollectionPrices).InsertOne(context.Background(), price)
	return err
}

// InsertPrices ...
func (m *MongoDAL) InsertPrices(prices ...models.Price) error {
	arr := make([]interface{}, len(prices))
	for i, p := range prices {
		if p.ID.IsZero() {
			p.ID = primitive.NewObjectID()
		}
		arr[i] = p
	}
	_, err := m.C(CollectionPrices).InsertMany(context.Background(), arr)
	return err
}

// FindPrice ...
func (m *MongoDAL) FindPrice(query persistence.Query) (*models.Price, error) {
	var result models.Price
	err := m.C(CollectionPrices).FindOne(context.Background(), query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GetPrice ...
func (m *MongoDAL) GetPrice(id string, query persistence.Query) (*models.Price, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.FindPrice(query.AddCondition("_id", ID))
}

// GetPrices ...
func (m *MongoDAL) GetPrices(query persistence.Query) ([]models.Price, error) {
	// TODO: Implement aggregate if we have any include queries
	var result []models.Price
	var ctx = context.Background()
	cursor, err := m.C(CollectionPrices).Find(ctx, query.GetConditions(), getFindOptions(query))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// DeletePrice ...
func (m *MongoDAL) DeletePrice(id string) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.C(CollectionPrices).DeleteOne(context.Background(), bson.M{"_id": ID})
	return err
}

// DeletePrices ...
func (m *MongoDAL) DeletePrices(query persistence.Query) (int64, error) {
	result, err := m.C(CollectionPrices).DeleteMany(context.Background(), query.GetConditions())
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, err
}
