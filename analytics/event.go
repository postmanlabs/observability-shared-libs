package analytics

import "time"

// Holds the name and properties of an analytics event.
type Event struct {
	// The name of the event.
	name string

	// The timestamp of the event.
	timestamp time.Time

	// Custom properties of the event.
	properties map[string]any
}

// Returns a new event with the given name and properties. The timestamp is set to the current time.
func NewEvent(name string, properties map[string]any) *Event {
	return &Event{
		name:       name,
		properties: properties,
		timestamp:  time.Now(),
	}
}

// Sets the event timestamp to the input time and returns the event.
func (e *Event) SetTime(t time.Time) *Event {
	e.timestamp = t
	return e
}
