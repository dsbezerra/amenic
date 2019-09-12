package contracts

// EventImageUploaded is emitted whenever a image is uploaded
type EventImageUploaded struct {
	MovieID string `json:"movie_id"`
	Name    string `json:"url"`
}

// EventName returns the event's name
func (e *EventImageUploaded) EventName() string {
	return "imageUploaded"
}
