package extractors

import (
	"crypto/md5"
	"fmt"
	"io"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/scraperutil"
	"github.com/dsbezerra/amenic/src/scraperservice/provider"
)

// TODO: DOC
type Extractor interface {
	ExtractedHash() string
	ExtractedCount() int

	// TODO: DOC
	Execute() error

	// TODO: DOC
	Complete()
}

// NewExtractor creates a brand new extractor instance.
func NewExtractor(data persistence.DataAccessLayer, p provider.Provider, s *models.ScraperRun) Extractor {
	var result Extractor

	t := s.Scraper.Type
	switch t {
	case scraperutil.TypeNowPlaying, scraperutil.TypeUpcoming:
		result = NewMovieExtractor(data, p, s)
	case scraperutil.TypeSchedule:
		result = NewScheduleExtractor(data, p, s)
	case scraperutil.TypePrices:
		result = NewPriceExtractor(data, p, s)
	}

	return result
}

// GetExtractedHash simple
func GetExtractedHash(data interface{}) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%v", data))
	return fmt.Sprintf("%x", h.Sum(nil))
}
