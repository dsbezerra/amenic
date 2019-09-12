package messagequeue

import (
	"fmt"

	"github.com/streadway/amqp"
)

const eventNameHeader = "x-event-name"

type amqpEventListener struct {
	conn     *amqp.Connection
	exchange string
	queue    string
	mapper   EventMapper
}

// NewAMQPEventListener ...
func NewAMQPEventListener(conn *amqp.Connection, exchange string, queue string) (EventListener, error) {
	listener := amqpEventListener{
		conn:     conn,
		exchange: exchange,
		queue:    queue,
		mapper:   NewEventMapper(),
	}

	err := listener.setup()
	if err != nil {
		return nil, err
	}

	return &listener, nil
}

func (l *amqpEventListener) setup() error {
	channel, err := l.conn.Channel()
	if err != nil {
		return err
	}

	defer channel.Close()

	err = channel.ExchangeDeclare(l.exchange, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	_, err = channel.QueueDeclare(l.queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not declare queue %s: %s", l.queue, err)
	}

	return nil
}

// Listen configures the event listener to listen for a set of events that are
// specified by name as parameter.
// This method will return two channels: One will contain successfully decoded
// events, the other will contain errors for messages that could not be
// successfully decoded.
func (l *amqpEventListener) Listen(eventNames ...string) (<-chan Event, <-chan error, error) {
	channel, err := l.conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	// Create binding between queue and exchange for each listened event type
	for _, event := range eventNames {
		if err := channel.QueueBind(l.queue, event, l.exchange, false, nil); err != nil {
			return nil, nil, fmt.Errorf("could not bind event %s to queue %s: %s", event, l.queue, err)
		}
	}

	msgs, err := channel.Consume(l.queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("could not consume queue: %s", err)
	}

	events := make(chan Event)
	errors := make(chan error)

	go func() {
		for msg := range msgs {
			rawEventName, ok := msg.Headers[eventNameHeader]
			if !ok {
				errors <- fmt.Errorf("message did not contain %s header", eventNameHeader)
				msg.Nack(false, false)
				continue
			}

			eventName, ok := rawEventName.(string)
			if !ok {
				errors <- fmt.Errorf("header %s did not contain string", eventNameHeader)
				msg.Nack(false, false)
				continue
			}

			event, err := l.mapper.MapEvent(eventName, msg.Body)
			if err != nil {
				errors <- fmt.Errorf("could not unmarshal event %s: %s", eventName, err)
				msg.Nack(false, false)
				continue
			}

			events <- event
			msg.Ack(false)
		}
	}()

	return events, errors, nil
}

func (l *amqpEventListener) Mapper() EventMapper {
	return l.mapper
}
