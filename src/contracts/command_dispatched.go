package contracts

import (
	"time"
)

// EventCommandDispatched is emitted whenever a command is dispatched by admin or other service
type EventCommandDispatched struct {
	ID               string        `json:"id"`
	TaskID           string        `json:"task_id"`
	Name             string        `json:"name"`
	Type             string        `json:"type"`
	Args             []string      `json:"args"`
	DispatchTime     time.Time     `json:"dispatch_time"`
	ExecutionTimeout time.Duration `json:"execution_timeout"`
}

// EventName returns the event's name
func (e *EventCommandDispatched) EventName() string {
	return "commandDispatched"
}
