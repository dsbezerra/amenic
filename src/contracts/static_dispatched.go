package contracts

import "time"

// EventStaticDispatched is emitted whenever we want to begin a static operation
type EventStaticDispatched struct {
	Name             string        `json:"name"`
	Type             string        `json:"type"`
	CinemaID         string        `json:"cinema_id"`
	TaskID           string        `json:"task_id"`
	DispatchTime     time.Time     `json:"dispatch_time"`
	ExecutionTimeout time.Duration `json:"execution_timeout"`
}

// EventName returns the event's name
func (e *EventStaticDispatched) EventName() string {
	return "staticDispatched"
}
