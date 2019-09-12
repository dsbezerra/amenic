package listener

import (
	"time"

	"github.com/dsbezerra/amenic/src/apiservice/v1"
	"github.com/dsbezerra/amenic/src/contracts"
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/shared"
	"github.com/dsbezerra/amenic/src/lib/util/scraperutil"
	"github.com/sirupsen/logrus"
)

// EventProcessor ...
type EventProcessor struct {
	EventListener messagequeue.EventListener
	Data          persistence.DataAccessLayer
	Log           *logrus.Entry
}

// ProcessEvents ...
func (p *EventProcessor) ProcessEvents() error {
	p.Log.Println("Listening to events...")

	var eventsList = []string{
		"staticDispatched", // Used to handle manual static stuff
		"scraperFinished",  // We need to recreate static files whenever a scraper runs to ensure it's updated
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
			p.Log.Printf("received error while processing message: %s", err)
		}
	}
}

func (p *EventProcessor) handle(event messagequeue.Event) {
	switch event.(type) {
	case *contracts.EventStaticDispatched:
		p.handleStaticDispatched(event.(*contracts.EventStaticDispatched))

	case *contracts.EventScraperFinished:
		e := event.(*contracts.EventScraperFinished)
		if e.Type != scraperutil.TypePrices {
			// Temporaly create and handle a fake staticDispatched event
			name := models.CommandCreateStatic
			_, ID := models.GenerateTaskID(name, []string{"-type", "home"})
			p.handleStaticDispatched(&contracts.EventStaticDispatched{
				TaskID:           ID,
				Name:             name,
				Type:             e.Type,
				DispatchTime:     time.Now().UTC(),
				ExecutionTimeout: time.Second * 5,
			})
		}

	default:
		p.Log.Infof("unknown event: %t", event)
	}
}

func (p *EventProcessor) handleStaticDispatched(e *contracts.EventStaticDispatched) {
	abort := messagequeue.CheckAbort(e.DispatchTime, e.ExecutionTimeout)
	if abort {
		p.Log.Infof("event %s aborted. reason: timeout reached", e.Name)
		return
	}

	p.Log.Infof("handling event %s. dispatched at: %s", e.Name, e.DispatchTime)

	var runError error
	var ran = true

	// NOTE: If we change our API we should update this
	t := v1.ToStaticType(e.Type)
	switch e.Name {
	case models.CommandCreateStatic:
		_, runError = v1.CreateStatic(p.Data, t)
	case models.CommandClearStatic:
		_, runError = v1.ClearStatic(t)
	default:
		p.Log.Infof("handler for event %s was not found", e.Name)
		ran = false
	}

	if ran {
		shared.UpdateTaskStatus(e.TaskID, p.Data, p.Log, runError)
	}
}
