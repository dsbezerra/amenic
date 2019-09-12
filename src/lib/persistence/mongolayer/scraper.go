package mongolayer

import (
	"context"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertScraper ...
func (m *MongoDAL) InsertScraper(scraper models.Scraper) error {
	_, err := m.C(CollectionScrapers).InsertOne(context.Background(), scraper)
	return err
}

// FindScraper ...
func (m *MongoDAL) FindScraper(query persistence.Query) (*models.Scraper, error) {
	var result models.Scraper
	err := m.C(CollectionScrapers).FindOne(context.Background(), query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GetScraper ...
func (m *MongoDAL) GetScraper(id string, query persistence.Query) (*models.Scraper, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.FindScraper(query.AddCondition("_id", ID))
}

// GetScrapers ...
func (m *MongoDAL) GetScrapers(query persistence.Query) ([]models.Scraper, error) {
	// TODO: Implement aggregate if we have any include queries
	var result []models.Scraper
	var ctx = context.Background()
	cursor, err := m.C(CollectionScrapers).Find(ctx, query.GetConditions(), getFindOptions(query))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// UpdateScraper ...
func (m *MongoDAL) UpdateScraper(id string, ms models.Scraper) (int64, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, err
	}
	result, err := m.C(CollectionScrapers).UpdateOne(context.Background(), bson.M{"_id": ID}, bson.M{"$set": ms})
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

// BuildScraperQuery converts a map of query string to mongolayer syntax for Scraper model
func (m *MongoDAL) BuildScraperQuery(q map[string]string) persistence.Query {
	query := BuildQuery("", q)
	if len(q) > 0 {
		// TODO: Implement specific scraper query
	}
	return query
}
