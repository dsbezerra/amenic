package mongolayer

import (
	"context"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsertScraperRun ...
func (m *MongoDAL) InsertScraperRun(scraperRun models.ScraperRun) error {
	err := m.InsertOne(CollectionScraperRuns, scraperRun)
	if err != nil {
		return err
	}
	if scraperRun.Scraper != nil {
		scraperRun.Scraper.Theater = nil // Make sure we don't store theater...
		scraperRun.Scraper.LastRun = scraperRun.ID
		_, err = m.UpdateScraper(scraperRun.ScraperID.Hex(), *scraperRun.Scraper)
	}
	return err
}

// FindScraperRun ...
func (m *MongoDAL) FindScraperRun(query persistence.Query) (*models.ScraperRun, error) {
	var result models.ScraperRun
	err := m.C(CollectionScraperRuns).FindOne(context.Background(), query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GetScraperRun ...
func (m *MongoDAL) GetScraperRun(id string, query persistence.Query) (*models.ScraperRun, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.FindScraperRun(query.AddCondition("_id", ID))
}

// GetScraperRuns ...
func (m *MongoDAL) GetScraperRuns(query persistence.Query) ([]models.ScraperRun, error) {
	// TODO: Implement aggregate if we have any include queries
	var result []models.ScraperRun
	var ctx = context.Background()
	cursor, err := m.C(CollectionScraperRuns).Find(ctx, query.GetConditions(), getFindOptions(query))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}
