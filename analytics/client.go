package analytics

import (
	"github.com/amplitude/analytics-go/amplitude"
	"github.com/golang/glog"
	"github.com/iancoleman/strcase"
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

	// This is included in all tracking events reported to Amplitude.
	amplitudeAppInfo amplitude.EventOptions
}

func NewClient(config Config) (Client, error) {
	amplitudeClient, amplitudeAppInfo, err := newAmplitudeClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create amplitude client")
	}

	return &clientImpl{
		config:           config,
		amplitudeClient:  amplitudeClient,
		amplitudeAppInfo: amplitudeAppInfo,
	}, nil
}

func (c clientImpl) TrackEvent(event *Event) error {
	var err error

	if c.config.IsAmplitudeEnabled && c.amplitudeClient != nil {
		// Added prefix to follow naming convention and differentiate between agent and internal service
		if c.config.IsInternalService {
			event.Name = "Insights - " + event.Name
		} else {
			event.Name = "Insights - Agent - " + event.Name
		}

		// Postman's property naming convention in Amplitude is snake case. So convert event.properties keys to snake case.
		properties := map[string]any{}
		for k, v := range event.Properties {
			properties[strcase.ToSnake(k)] = v
		}

		c.amplitudeClient.Track(amplitude.Event{
			UserID:          event.DistinctID,
			EventType:       event.Name,
			EventProperties: properties,
			EventOptions:    c.amplitudeAppInfo,
		})
	}

	return errors.Wrapf(
		err,
		"failed to send analytics tracking event '%s' for distinct id %s",
		event.Name,
		event.DistinctID,
	)
}

func (c clientImpl) Track(distinctID string, name string, properties map[string]any) error {
	return c.TrackEvent(NewEvent(distinctID, name, properties))
}

func (c clientImpl) Close() error {
	var err error

	if c.amplitudeClient != nil {
		c.amplitudeClient.Shutdown()
	}

	return err
}

func newAmplitudeClient(config Config) (amplitude.Client, amplitude.EventOptions, error) {
	if !config.IsAmplitudeEnabled {
		return nil, amplitude.EventOptions{}, nil
	}

	rawAmplitudeConfig := config.AmplitudeConfig

	if rawAmplitudeConfig == (AmplitudeConfig{}) {
		return nil, amplitude.EventOptions{}, errors.New("unable to construct new amplitude analytics client. amplitude config cannot be empty")
	}

	if rawAmplitudeConfig.AmplitudeAPIKey == "" {
		return nil, amplitude.EventOptions{}, errors.New("unable to construct new amplitude analytics client. API key cannot be empty")
	}

	amplitudeConfig := amplitude.NewConfig(rawAmplitudeConfig.AmplitudeAPIKey)

	amplitudeConfig.Logger = provideAmplitudeLogger(rawAmplitudeConfig.IsLoggingEnabled)

	if rawAmplitudeConfig.IsBatchingEnabled {
		amplitudeConfig.UseBatch = rawAmplitudeConfig.IsBatchingEnabled
		amplitudeConfig.ServerURL = rawAmplitudeConfig.AmplitudeEndpoint
	}

	if rawAmplitudeConfig.FlushQueueSize > 0 {
		amplitudeConfig.FlushQueueSize = rawAmplitudeConfig.FlushQueueSize
	}

	amplitudeAppInfo := amplitude.EventOptions{
		AppVersion:  config.App.Version,
		VersionName: config.App.Name,
	}

	return amplitude.NewClient(amplitudeConfig), amplitudeAppInfo, nil
}

// Returns the logger to use for the Amplitude client if logging is enabled.
func provideAmplitudeLogger(isLoggingEnabled bool) amplitude.Logger {
	if isLoggingEnabled {
		return &amplitudeLogger{}
	}

	return &disabledAmplitudeLogger{}
}

// A custom segment logger that logs to glog.
type amplitudeLogger struct{}

func (d amplitudeLogger) Debugf(format string, args ...any) {
	glog.Infof(format, args...)
}

func (d amplitudeLogger) Infof(format string, args ...any) {
	glog.Infof(format, args...)
}

func (d amplitudeLogger) Warnf(format string, args ...any) {
	glog.Warningf(format, args...)
}

func (d amplitudeLogger) Errorf(format string, args ...interface{}) {
	glog.Errorf(format, args...)
}

// A custom segment logger that does nothing.
// This is used when logging is disabled as the segment client requires a logger (the client uses its own default logger even when none is specified).
type disabledAmplitudeLogger struct{}

func (d disabledAmplitudeLogger) Debugf(format string, args ...any) {
	// Do nothing.
}

func (d disabledAmplitudeLogger) Infof(format string, args ...any) {
	// Do nothing.
}

func (d disabledAmplitudeLogger) Warnf(format string, args ...any) {
	// Do nothing.
}

func (d disabledAmplitudeLogger) Errorf(format string, args ...interface{}) {
	// Do nothing.
}
