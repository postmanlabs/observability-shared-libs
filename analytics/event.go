package analytics

import (
	"time"
)

// Holds the name and properties of an analytics event.
type Event struct {
	// The value used to uniquely identify the user who triggered the event.
	DistinctID string

	// The name of the event.
	Name string

	// The timestamp of the event.
	Timestamp time.Time

	// Custom properties of the event.
	Properties map[string]any
}

// Returns a new event with the given name and properties.
// The event is initialized with the current time as the timestamp.
func NewEvent(distinctID string, name string, properties map[string]any) *Event {
	return &Event{
		DistinctID: distinctID,
		Name:       name,
		Properties: properties,
		Timestamp:  time.Now(),
	}
}
