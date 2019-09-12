package listener

import (
	"log"

	"github.com/dsbezerra/amenic/src/contracts"
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/shared"
	"github.com/dsbezerra/amenic/src/notificationservice/fcm"
	"github.com/dsbezerra/amenic/src/notificationservice/task"
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
	log.Println("Listening to events...")

	var eventsList = []string{
		"commandDispatched",
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

	var runError error
	var ran = true

	switch e.Type {
	case models.TaskCheckOpeningMovies:
		err := task.CheckOpeningMovies(p.Data)
		switch err {
		case fcm.ErrAuthKeyUndefined, fcm.ErrInvalidNotification:
			runError = err
		default:
			if err != nil {
				p.Log.Errorf("unknown error %s", err.Error())
			}
		}
	default:
		p.Log.Infof("handler for event %s was not found", e.Name)
		ran = false
	}

	if ran {
		shared.UpdateTaskStatus(e.TaskID, p.Data, p.Log, runError)
	}
}
