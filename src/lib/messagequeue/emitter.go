package messagequeue

import (
	"github.com/streadway/amqp"
)

// EventEmitter ...
type EventEmitter interface {
	Emit(event Event) error
}

// ampqEventEmitter ...
type amqpEventEmitter struct {
	conn     *amqp.Connection
	exchange string
	events   chan *emittedEvent
}

type emittedEvent struct {
	event     Event
	errorChan chan error
}
