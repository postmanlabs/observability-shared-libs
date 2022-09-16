package analytics

import "time"

// Holds the name and properties of an analytics event.
type Event struct {
	// The name of the event.
	name string

	// The timestamp of the event. Set to nil to use the current time.
	timestamp time.Time

	// Custom properties of the event.
	properties map[string]any
}

// Returns a new event with the given name, properties, and timestamp. If no timestamp is provided, the current time is used.
func NewEvent(name string, properties map[string]any, timestamp ...time.Time) *Event {
	var eventTime time.Time
	if len(timestamp) > 0 {
		eventTime = timestamp[0]
	} else {
		eventTime = time.Now()
	}

	return &Event{
		name:       name,
		properties: properties,
		timestamp:  eventTime,
	}
}
