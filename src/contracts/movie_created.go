package contracts

// EventMovieCreated is emitted whenever a new movie is created
type EventMovieCreated struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EventName returns the event's name
func (e *EventMovieCreated) EventName() string {
	return "movieCreated"
}
