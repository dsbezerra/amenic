package messagequeue

import "time"

// Event ...
type Event interface {
	EventName() string
}

// CheckAbort returns true if we need to abort and false otherwise
func CheckAbort(timestamp time.Time, timeout time.Duration) bool {
	return time.Now().UTC().After(timestamp.Add(timeout))
}
