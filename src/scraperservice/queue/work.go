package queue

import (
	"errors"
)

var WorkQueue = make(chan WorkRequest, 100)

type WorkRequest struct {
	ScraperID     string
	IgnoreLastRun bool
}

func AddWork(wr WorkRequest) error {
	if wr.ScraperID == "" {
		// TODO: Improve
		return errors.New("invalid scraper id")
	}

	if WorkQueue != nil {
		WorkQueue <- wr
	}

	return nil
}
