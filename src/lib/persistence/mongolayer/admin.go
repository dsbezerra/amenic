package mongolayer

import (
	"context"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertAdmin ...
func (m *MongoDAL) InsertAdmin(admin models.Admin) error {
	if admin.CreatedAt == nil {
		admin.CreatedAt = getCurrentTime()
	}
	_, err := m.C(CollectionAdmins).InsertOne(context.Background(), admin)
	return err
}

// FindAdmin ...
func (m *MongoDAL) FindAdmin(query persistence.Query) (*models.Admin, error) {
	var result models.Admin
	err := m.C(CollectionAdmins).FindOne(context.Background(), query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GetAdmin ...
func (m *MongoDAL) GetAdmin(id string, query persistence.Query) (*models.Admin, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.FindAdmin(query.AddCondition("_id", ID))
}

// GetAdmins ...
func (m *MongoDAL) GetAdmins(query persistence.Query) ([]models.Admin, error) {
	var result []models.Admin
	var ctx = context.Background()
	cursor, err := m.C(CollectionAdmins).Find(ctx, query.GetConditions(), getFindOptions(query))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// DeleteAdmin ...
func (m *MongoDAL) DeleteAdmin(id string) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.C(CollectionAdmins).DeleteOne(context.Background(), bson.M{"_id": ID})
	return err
}
