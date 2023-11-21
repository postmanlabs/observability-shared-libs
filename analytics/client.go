package analytics

import (
	"github.com/amplitude/analytics-go/amplitude"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Client interface {
	// Sends the given tracking event to Segment and Mixpanel (if enabled).
	TrackEvent(event *Event) error

	// A shorthand wrapper method for TrackEvent that sends a tracking event with the given distinct id, name and properties.
	Track(distinctID string, name string, properties map[string]any) error

	Close() error
}

type clientImpl struct {
	// The analytics client configuration.
	config Config

	// The internal client used to send events to Amplitude.
	amplitudeClient amplitude.Client

	// App info from which event is sent
	amplitudeAppInfo amplitude.EventOptions
}

func NewClient(config Config) (Client, error) {
	amplitudeAppInfo := amplitude.EventOptions{
		AppVersion:  config.App.Version,
		VersionName: config.App.Name,
	}

	if config.AmplitudeAPIKey == "" {
		return nil, errors.New("unable to construct new amplitude analytics client. API key cannot be empty")
	}

	amplitudeConfig := amplitude.NewConfig(config.AmplitudeAPIKey)

	amplitudeConfig.Logger = provideLogger(config.IsLoggingEnabled)

	if config.IsBatchingEnabled {
		amplitudeConfig.UseBatch = config.IsBatchingEnabled
		amplitudeConfig.ServerURL = config.AmplitudeEndpoint
	}

	if config.FlushQueueSize > 0 {
		amplitudeConfig.FlushQueueSize = config.FlushQueueSize
	}

	amplitudeClient := amplitude.NewClient(amplitudeConfig)

	return &clientImpl{
		config:           config,
		amplitudeClient:  amplitudeClient,
		amplitudeAppInfo: amplitudeAppInfo,
	}, nil
}

func (c clientImpl) TrackEvent(event *Event) error {
	// Added prefix to follow naming convention and differentiate between agent and internal service
	if c.config.IsAgent {
		event.name = "Live Insights Agent - " + event.name
	} else {
		event.name = "Live Insights - " + event.name
	}

	c.amplitudeClient.Track(amplitude.Event{
		UserID:          event.distinctID,
		EventType:       event.name,
		EventProperties: event.properties,
		EventOptions:    c.amplitudeAppInfo,
	})

	return nil
}

func (c clientImpl) Track(distinctID string, name string, properties map[string]any) error {
	return c.TrackEvent(NewEvent(distinctID, name, properties))
}

func (c clientImpl) Close() error {
	c.amplitudeClient.Shutdown()
	return nil
}

// Returns the logger to use for the Amplitude client if logging is enabled.
func provideLogger(isLoggingEnabled bool) amplitude.Logger {
	if isLoggingEnabled {
		return &analyticsLogger{}
	}

	return &disabledLogger{}
}

// A custom segment logger that logs to glog.
type analyticsLogger struct{}

func (d analyticsLogger) Debugf(format string, args ...any) {
	glog.Infof(format, args...)
}

func (d analyticsLogger) Infof(format string, args ...any) {
	glog.Infof(format, args...)
}

func (d analyticsLogger) Warnf(format string, args ...any) {
	glog.Warningf(format, args...)
}

func (d analyticsLogger) Errorf(format string, args ...interface{}) {
	glog.Errorf(format, args...)
}

// A custom segment logger that does nothing.
// This is used when logging is disabled as the segment client requires a logger (the client uses its own default logger even when none is specified).
type disabledLogger struct{}

func (d disabledLogger) Debugf(format string, args ...any) {
	// Do nothing.
}

func (d disabledLogger) Infof(format string, args ...any) {
	// Do nothing.
}

func (d disabledLogger) Warnf(format string, args ...any) {
	// Do nothing.
}

func (d disabledLogger) Errorf(format string, args ...interface{}) {
	// Do nothing.
}
