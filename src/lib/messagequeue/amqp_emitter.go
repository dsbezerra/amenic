package messagequeue

import (
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

// NewAMQPEventEmitter ...
func NewAMQPEventEmitter(conn *amqp.Connection, exchange string) (EventEmitter, error) {
	emitter := amqpEventEmitter{
		conn:     conn,
		exchange: exchange,
	}

	err := emitter.setup()
	if err != nil {
		return nil, err
	}

	return &emitter, nil
}

func (a *amqpEventEmitter) setup() error {
	channel, err := a.conn.Channel()
	if err != nil {
		return err
	}

	defer channel.Close()

	// Normally, all(many) of these options should be configurable.
	// For our example, it'll probably do.
	err = channel.ExchangeDeclare(a.exchange, "topic", true, false, false, false, nil)
	return err
}

func (a *amqpEventEmitter) Emit(event Event) error {
	channel, err := a.conn.Channel()
	if err != nil {
		return err
	}

	defer channel.Close()

	b, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not JSON-serialize event: %s", err)
	}

	msg := amqp.Publishing{
		Headers:     amqp.Table{"x-event-name": event.EventName()},
		ContentType: "application/json",
		Body:        b,
	}

	err = channel.Publish(a.exchange, event.EventName(), false, false, msg)
	return err
}
