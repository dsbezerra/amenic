package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (

	// Scraper is used to track scraper's runs, aswell check for changes in the page content
	// to minimize its execution and server load.
	Scraper struct {
		// ID is a MD5 hash of the combination of fields Theater and Type.
		ID         primitive.ObjectID `json:"_id" bson:"_id"`
		TheaterID  primitive.ObjectID `json:"theater_id" bson:"theaterId"`                  // TheaterID indicates the theater of this scraper corresponds.
		Type       string             `json:"type" bson:"type"`                             // Type of operation of scraper (now_playing/upcoming/prices/schedule)
		Provider   string             `json:"provider" bson:"provider"`                     // Provider ...
		LastRun    primitive.ObjectID `json:"last_run,omitempty" bson:"last_run,omitempty"` // LastRun contains informations about last run
		LastRunDoc *ScraperRun        `json:"-" bson:"-"`                                   // LastRunDoc contains the last run document
		Theater    *Theater           `json:"theater,omitempty" bson:"theater,omitempty"`   // Theater document
	}

	// ScraperRun is the result of any scraper operation.
	ScraperRun struct {
		ID             primitive.ObjectID `json:"_id" bson:"_id"`                         // ID is the document identifier
		ScraperID      primitive.ObjectID `json:"scraper_id" bson:"scraper_id"`           // ScraperID indicates from which scraper this belongs
		ResultCode     string             `json:"result_code" bson:"result_code"`         // ResultCode is a code in string format to make easier to know if run was successful or not
		Error          string             `json:"error" bson:"error"`                     // Error is the possible error message or stack trace encounter in run
		StartTime      *time.Time         `json:"start_time" bson:"start_time"`           // StartTime is the time the run started
		CompleteTime   *time.Time         `json:"complete_time" bson:"complete_time"`     // CompleteTime is the time the run finished
		ExtractedHash  string             `json:"extracted_hash" bson:"extracted_hash"`   // ExtractedHash is used to store the hash data so we can easily determine if it changed or not
		ExtractedCount int                `json:"extracted_count" bson:"extracted_count"` // ExtractedCount indicates how many items were extracted
		Scraper        *Scraper           `json:"-" bson:"-"`                             // Scraper scraper from which this run belongs
		Movies         []Movie            `json:"-" bson:"-"`                             // Movies retrieved from a scraper's execution. (now_playing/upcoming)
		Sessions       []Session          `json:"-" bson:"-"`                             // Sessions retrieved from a scraper's execution. (schedule)
		Prices         []Price            `json:"-" bson:"-"`                             // Prices retrieved from a scraper's execution. (prices)
	}
)

// Finish just adds the complete time.
func (r *ScraperRun) Finish() {
	end := time.Now().UTC()
	r.CompleteTime = &end
}
