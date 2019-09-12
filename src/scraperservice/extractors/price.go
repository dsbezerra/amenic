package extractors

import (
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/scraperservice/provider"
	"github.com/sirupsen/logrus"
)

type (
	PriceExtractor struct {
		Data     persistence.DataAccessLayer
		Logger   *logrus.Entry
		Provider provider.Provider
		Run      *models.ScraperRun
		Prices   []models.Price
	}
)

// NewPriceExtractor ...
func NewPriceExtractor(data persistence.DataAccessLayer, p provider.Provider, s *models.ScraperRun) *PriceExtractor {
	result := &PriceExtractor{
		Data:     data,
		Logger:   logrus.WithFields(logrus.Fields{"extractor": "Price"}),
		Run:      s,
		Provider: p,
	}
	return result
}

// Execute extract prices information for a given provider
func (e *PriceExtractor) Execute() error {
	result, err := e.Provider.GetPrices()
	if err != nil {
		return err
	}
	e.Prices = result
	return nil
}

// Complete TODO
func (e *PriceExtractor) Complete() {
	query := e.Data.DefaultQuery().
		AddCondition("theaterId", e.Run.Scraper.TheaterID)
	_, err := e.Data.DeletePrices(query)
	if err != nil {
		e.Logger.Error(err)
	}
	err = e.Data.InsertPrices(e.Prices...)
	if err != nil {
		e.Logger.Error(err)
	}
}

// ExtractedHash TODO
func (e *PriceExtractor) ExtractedHash() string {
	return GetExtractedHash(e.Prices)
}

// ExtractedCount TODO
func (e *PriceExtractor) ExtractedCount() int {
	return len(e.Prices)
}
