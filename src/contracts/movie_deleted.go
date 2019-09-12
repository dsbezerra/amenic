package contracts

// EventMovieDeleted is emitted whenever a movie is deleted
type EventMovieDeleted struct {
	MovieID string `json:"movie_id"`
	Name    string `json:"name"`
}

// EventName returns the event's name
func (e *EventMovieDeleted) EventName() string {
	e.Name = "movieDeleted"
	return e.Name
}
