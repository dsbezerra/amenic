package messagequeue

// EventMapper ...
type EventMapper interface {
	MapEvent(string, interface{}) (Event, error)
}

// NewEventMapper ...
func NewEventMapper() EventMapper {
	return &StaticEventMapper{}
}
