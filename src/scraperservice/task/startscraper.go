package task

import (
	"errors"
	"log"
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/scraperutil"
	"github.com/dsbezerra/amenic/src/scraperservice/extractors"
	"github.com/dsbezerra/amenic/src/scraperservice/provider"
)

type ScraperOptions struct {
	ScraperID     string `json:"scraperId"`
	TheaterID     string `json:"theaterId"`
	Type          string `json:"type"`
	Provider      string `json:"provider"`
	IgnoreLastRun bool   `json:"ignore_last_run"`
}

// StartScraper ...
func StartScraper(data persistence.DataAccessLayer, opts ScraperOptions) (*models.ScraperRun, error) {
	run, err := InitScraper(data, opts)
	if err != nil {
		return nil, err
	}

	scraper := run.Scraper
	p := provider.NewProvider(data, scraper.Provider, scraper.Theater.InternalID)
	if p == nil {
		return nil, errors.New("couldn't run scraper for the specified provider")
	}
	e := extractors.NewExtractor(data, p, run)
	err = e.Execute()
	if err != nil {
		// Update scraper run with error
		t := time.Now().UTC()
		run.CompleteTime = &t
		run.Error = err.Error()
		// TODO: Parse error to determine result code
	} else {
		scraper := run.Scraper
		run.ExtractedHash = e.ExtractedHash()
		run.ExtractedCount = e.ExtractedCount()
		t := time.Now().UTC()
		run.CompleteTime = &t
		if scraper.LastRunDoc != nil && scraper.LastRunDoc.ExtractedHash == run.ExtractedHash {
			run.ResultCode = scraperutil.RunResultNotModified
		} else if run.ExtractedCount == 0 {
			run.ResultCode = scraperutil.RunResultNotFound
			// TODO: Notify via email that we found nothing?
		} else {
			run.ResultCode = scraperutil.RunResultSuccess
		}
		e.Complete()
	}

	return run, data.InsertScraperRun(*run)
}

// InitScraper ...
func InitScraper(data persistence.DataAccessLayer, options ScraperOptions) (*models.ScraperRun, error) {

	var scraper *models.Scraper
	var theater *models.Theater
	var err error
	var theaterID string

	if options.ScraperID != "" {
		scraper, err = data.GetScraper(options.ScraperID, data.DefaultQuery())
	}

	if err != nil {
		return nil, err
	}

	if scraper != nil {
		theaterID = scraper.TheaterID.Hex()
	} else {
		if options.TheaterID == "" || options.Type == "" || options.Provider == "" {
			return nil, errors.New("theater, operation and provider must be valid")
		}
		theaterID = options.TheaterID
	}

	theater, err = data.GetTheater(theaterID, data.DefaultQuery())
	if err != nil {
		return nil, err
	}

	if scraper == nil {
		query := data.DefaultQuery().
			AddCondition("theaterId", theater.ID).
			AddCondition("type", options.Type).
			AddCondition("provider", options.Provider)
		scraper, err = data.FindScraper(query)
		if err != nil {
			return nil, errors.New("couldn't retrieve scraper with theses options")
		}
	}

	if !options.IgnoreLastRun && !scraper.LastRun.IsZero() {
		l, err := data.GetScraperRun(scraper.LastRun.Hex(), data.DefaultQuery())
		if err != nil {
			log.Printf("couldn't find scraper last run with ID: %s", scraper.LastRun.Hex())
		}
		scraper.LastRunDoc = l
	}

	r := scraperutil.NewScraperRun(scraper.Type)
	r.Scraper = scraper
	r.ScraperID = scraper.ID
	r.Scraper.Theater = theater

	return r, err
}
