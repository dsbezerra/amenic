package listener

import (
	"log"
	"strconv"

	"github.com/dsbezerra/amenic/src/contracts"
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/scraperservice/queue"
	"github.com/sirupsen/logrus"
)

// EventProcessor ...
type EventProcessor struct {
	EventListener messagequeue.EventListener
	EventEmitter  messagequeue.EventEmitter
	Data          persistence.DataAccessLayer
	Log           *logrus.Entry
}

// ProcessEvents ...
func (p *EventProcessor) ProcessEvents() error {
	log.Println("Listening to events...")

	var eventsList = []string{
		"commandDispatched",
		"scraperRun",
	}

	received, errors, err := p.EventListener.Listen(eventsList...)
	if err != nil {
		return err
	}

	for {
		select {
		case event := <-received:
			p.handle(event)
		case err = <-errors:
			log.Printf("received error while processing message: %s", err)
		}
	}
}

func (p *EventProcessor) handle(event messagequeue.Event) {
	switch event.(type) {
	case *contracts.EventCommandDispatched:
		p.handleCommandDispatched(event.(*contracts.EventCommandDispatched))
	default:
		log.Printf("unknown event: %t", event)
	}
}

func (p *EventProcessor) handleCommandDispatched(e *contracts.EventCommandDispatched) {
	abort := messagequeue.CheckAbort(e.DispatchTime, e.ExecutionTimeout)
	if abort {
		p.Log.Warnf("event %s aborted. reason: timeout reached", e.Name)
		return
	}

	p.Log.Infof("handling event %s. dispatched at: %s", e.Name, e.DispatchTime)

	switch e.Type {
	case models.TaskStartScraper:
		args := models.ParseArgs(e.Args)

		var ignoreLastRun bool
		if value, ok := args["ignore_last_run"]; ok {
			ignoreLastRun, _ = strconv.ParseBool(value)
		}

		if args["scraper_id"] != "" {
			queue.AddWork(queue.WorkRequest{
				ScraperID:     args["scraper_id"],
				IgnoreLastRun: ignoreLastRun,
			})
			return
		}

		var query persistence.Query
		if value, ok := args["theater"]; ok {
			theater, err := p.Data.GetTheater(value, p.Data.DefaultQuery())
			if err == nil {
				if query == nil {
					query = p.Data.DefaultQuery()
				}
				query.AddCondition("theaterId", theater.ID)
			}
		}

		if value, ok := args["type"]; ok {
			if query == nil {
				query = p.Data.DefaultQuery()
			}
			query.AddCondition("type", value)
		}

		if query != nil {
			scrapers, err := p.Data.GetScrapers(query)
			if err != nil {
				p.Log.Error(err.Error())
				return
			}
			for _, s := range scrapers {
				queue.AddWork(queue.WorkRequest{
					ScraperID:     s.ID.Hex(),
					IgnoreLastRun: ignoreLastRun,
				})
			}
		}

		// Add to queue for each
	default:
		p.Log.Infof("handler for event %s was not found", e.Name)
	}
}
