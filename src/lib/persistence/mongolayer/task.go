package mongolayer

import (
	"context"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// InsertTask ...
func (m *MongoDAL) InsertTask(task models.Task) error {
	_, err := m.C(CollectionTasks).InsertOne(context.Background(), task)
	return err
}

// FindTask ...
func (m *MongoDAL) FindTask(query persistence.Query) (*models.Task, error) {
	var result models.Task
	err := m.C(CollectionTasks).FindOne(context.Background(), query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GetTask ...
func (m *MongoDAL) GetTask(id string, query persistence.Query) (*models.Task, error) {
	return m.FindTask(query.AddCondition("_id", id))
}

// GetTasks ...
func (m *MongoDAL) GetTasks(query persistence.Query) ([]models.Task, error) {
	// TODO: Implement aggregate if we have any include queries
	var result []models.Task
	var ctx = context.Background()
	cursor, err := m.C(CollectionTasks).Find(ctx, query.GetConditions(), getFindOptions(query))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// UpdateTask ...
func (m *MongoDAL) UpdateTask(id string, mt models.Task) (int64, error) {
	result, err := m.C(CollectionTheaters).UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": mt})
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

// DeleteTask ...
func (m *MongoDAL) DeleteTask(id string) error {
	_, err := m.C(CollectionTasks).DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}

// DeleteTasks ...
func (m *MongoDAL) DeleteTasks(query persistence.Query) (int64, error) {
	result, err := m.C(CollectionTasks).DeleteMany(context.Background(), query.GetConditions())
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, err
}

// EnsureTasksExists ...
func (m *MongoDAL) EnsureTasksExists(mp map[string]models.Task) {
	if m == nil {
		return
	}

	for _, v := range mp {
		ID := v.ID
		if ID == "" {
			_, ID = v.GenerateID()
		}

		if ID != "" {
			_, err := m.GetTask(ID, m.DefaultQuery())
			if err != mongo.ErrNoDocuments {
				m.InsertTask(v)
			}
		}
	}
}
