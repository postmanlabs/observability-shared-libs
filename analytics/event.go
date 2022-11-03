package analytics

import (
	"github.com/akitasoftware/go-utils/optionals"
	segment "github.com/segmentio/analytics-go/v3"
	"time"
)

// Holds the name and properties of an analytics event.
type Event struct {
	// The value used to uniquely identify the user who triggered the event.
	distinctID string

	// The name of the event.
	name string

	// The timestamp of the event.
	timestamp time.Time

	// Custom properties of the event.
	properties map[string]any

	// The integrations that the event should be sent to.
	//This will override any default integrations passed to the Analytics client.
	integrationsOverride optionals.Optional[segment.Integrations]
}

// Returns a new event with the given name and properties.
// The event is initialized with the current time as the timestamp.
func NewEvent(distinctID string, name string, properties map[string]any) *Event {
	return &Event{
		distinctID:           distinctID,
		name:                 name,
		properties:           properties,
		timestamp:            time.Now(),
		integrationsOverride: optionals.None[segment.Integrations](),
	}
}

// Sets the event timestamp to the input time and returns the event.
func (e *Event) SetTime(t time.Time) *Event {
	e.timestamp = t
	return e
}

// Sets the integrations that the event should be sent to.
// This will override any default integrations set by the Analytics client.
func (e *Event) SetIntegrations(integrations segment.Integrations) *Event {
	e.integrationsOverride = optionals.Some(integrations)
	return e
}

func (e *Event) DistinctID() string {
	return e.distinctID
}

func (e *Event) Name() string {
	return e.name
}
