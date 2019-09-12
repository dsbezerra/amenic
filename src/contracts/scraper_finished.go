package contracts

// EventScraperFinished is emitted whenever a scraper has finished to run
type EventScraperFinished struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ScraperID string `json:"scraper_id"`
	Type      string `json:"type"`
}

// EventName returns the event's name
func (e *EventScraperFinished) EventName() string {
	return "scraperFinished"
}
