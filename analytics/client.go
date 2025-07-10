package analytics

import (
	"github.com/amplitude/analytics-go/amplitude"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/segmentio/analytics-go/v3"
)

type AnalyticsProvider string

const (
	Amplitude AnalyticsProvider = "amplitude"
	Segment   AnalyticsProvider = "segment"
)

type Client interface {
	// Sends the given tracking event to Amplitude (if enabled).
	TrackEvent(event *Event)

	// A shorthand wrapper method for TrackEvent that sends a tracking event with the given distinct id, name and properties.
	Track(distinctID string, name string, properties map[string]any)

	// Sends the given tracking event to Segment (if enabled).
	TrackSegmentEvent(event *Event)

	Close() error
}

type NullClient struct{}

var _ Client = &NullClient{}

func (NullClient) TrackEvent(*Event) {
	// Do nothing.
}

func (NullClient) Track(string, string, map[string]any) {
	// Do nothing.
}

func (NullClient) TrackSegmentEvent(*Event) {
	// Do nothing.
}

func (NullClient) Close() error {
	// Do nothing.
	return nil
}

type clientImpl struct {
	// The analytics client configuration.
	config Config

	// The internal client used to send events to Amplitude.
	amplitudeClient amplitude.Client

	// This is included in all tracking events reported to Amplitude.
	amplitudeAppInfo amplitude.EventOptions

	// The internal client used to send events to Segment.
	segmentClient analytics.Client
}

func NewClient(config Config) (Client, error) {
	amplitudeClient, amplitudeAppInfo, err := newAmplitudeClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create amplitude client")
	}

	segmentClient, err := newSegmentClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create segment client")
	}

	return &clientImpl{
		config:           config,
		amplitudeClient:  amplitudeClient,
		amplitudeAppInfo: amplitudeAppInfo,
		segmentClient:    segmentClient,
	}, nil
}

func (c clientImpl) prepareEvent(event *Event) {
	// Added prefix to follow naming convention and differentiate between agent and internal service
	if c.config.IsInternalService {
		event.name = "Insights - " + event.name
	} else {
		event.name = "Insights - Agent - " + event.name
	}

	// Postman's property naming convention in Amplitude is snake case. So convert event.properties keys to snake case.
	properties := map[string]any{}
	for k, v := range event.properties {
		properties[strcase.ToSnake(k)] = v
	}

	event.properties = properties
}

func (c clientImpl) TrackEvent(event *Event) {
	c.prepareEvent(event)

	if c.config.IsAmplitudeEnabled && c.amplitudeClient != nil {
		c.amplitudeClient.Track(amplitude.Event{
			UserID:          event.distinctID,
			EventType:       event.name,
			EventProperties: event.properties,
			EventOptions:    c.amplitudeAppInfo,
		})
	}
}

func (c clientImpl) Track(distinctID string, name string, properties map[string]any) {
	c.TrackEvent(NewEvent(distinctID, name, properties))
}

func (c clientImpl) TrackSegmentEvent(event *Event) {
	c.prepareEvent(event)

	if c.config.IsSegmentEnabled && c.segmentClient != nil {
		c.segmentClient.Enqueue(analytics.Track{
			UserId:     event.distinctID,
			Event:      event.name,
			Properties: event.properties,
		})
	}

}

func (c clientImpl) Close() error {
	var err error

	if c.amplitudeClient != nil {
		c.amplitudeClient.Shutdown()
	}

	if c.segmentClient != nil {
		err = c.segmentClient.Close()
	}

	return err
}
