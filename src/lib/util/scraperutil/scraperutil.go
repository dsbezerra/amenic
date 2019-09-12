package scraperutil

import (
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// TypeNowPlaying indicates the scraper is getting Now Playing movies data
	TypeNowPlaying = "now_playing"

	// TypeUpcoming indicates the scraper is getting Upcoming movies data
	TypeUpcoming = "upcoming"

	// TypeSchedule indicates the scraper is getting current now playing movies schedule
	TypeSchedule = "schedule"

	// TypePrices indicates the scraper is getting theater prices
	TypePrices = "prices"

	// RunResultSuccess indicates the run's execution was successfull.
	RunResultSuccess = "success"

	// RunResultNotModified indicates the data was not modified since last run.
	RunResultNotModified = "not_modified"

	// RunResultNotFound indicates the run encountered zero items in its execution.
	RunResultNotFound = "not_found"

	// RunResultTimeout indicates the run encountered a timeout error during its execution.
	// TODO: detect this!
	RunResultTimeout = "server_timeout"
)

// NewScraperRun creates an instance of ScraperRun for a given theater and op
func NewScraperRun(operation string) *models.ScraperRun {
	if operation == "" {
		return nil
	}
	var result *models.ScraperRun
	start := time.Now().UTC()
	result = &models.ScraperRun{
		ID:        primitive.NewObjectID(),
		StartTime: &start,
	}
	return result
}
