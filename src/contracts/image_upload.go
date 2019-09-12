package contracts

// EventImageUpload is emitted whenever a image should be uploaded
type EventImageUpload struct {
	MovieID   string `json:"movie_id"`
	ImageType string `json:"image_type"`
	URL       string `json:"url"`
}

// EventName returns the event's name
func (e *EventImageUpload) EventName() string {
	return "imageUpload"
}
