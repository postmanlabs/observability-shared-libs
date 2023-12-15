package analytics

import (
	"github.com/amplitude/analytics-go/amplitude"
	"github.com/dukex/mixpanel"
	"github.com/golang/glog"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	segment "github.com/segmentio/analytics-go/v3"
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
	// The internal client used to send events to Segment.
	segmentClient segment.Client
	// The default integrations to use for all events sent to Segment.
	// This controls which destinations the events are sent to.
	defaultIntegrations segment.Integrations

	// TODO: Remove Mixpanel once we've confirmed that Segment is working.
	mixpanelClient mixpanel.Mixpanel

	// The internal client used to send events to Amplitude.
	amplitudeClient amplitude.Client
	// This is included in all tracking events reported to Amplitude.
	amplitudeAppInfo amplitude.EventOptions
}

func NewClient(config Config) (Client, error) {
	segmentClient, err := newSegmentClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create segment client")
	}

	mixpanelClient, err := newMixpanelClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create mixpanel client")
	}

	amplitudeClient, amplitudeAppInfo, err := newAmplitudeClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create amplitude client")
	}

	return &clientImpl{
		config:              config,
		segmentClient:       segmentClient,
		defaultIntegrations: provideDefaultIntegrations(config),
		mixpanelClient:      mixpanelClient,
		amplitudeClient:     amplitudeClient,
		amplitudeAppInfo:    amplitudeAppInfo,
	}, nil
}

func (c clientImpl) TrackEvent(event *Event) error {
	var err error

	if c.config.IsSegmentEnabled && c.segmentClient != nil {
		integrations := c.defaultIntegrations
		if integrationsOverride, ok := event.integrationsOverride.Get(); ok {
			integrations = integrationsOverride
		}

		segmentErr := c.segmentClient.Enqueue(
			segment.Track{
				UserId:       event.distinctID,
				Event:        event.name,
				Properties:   event.properties,
				Timestamp:    event.timestamp,
				Integrations: integrations,
			},
		)

		if segmentErr != nil {
			err = multierror.Append(err, segmentErr)
		}
	}

	// TODO: Remove Mixpanel once we've fully migrated to Segment.
	if c.config.IsMixpanelEnabled && c.mixpanelClient != nil {
		mixpanelErr := c.mixpanelClient.Track(
			event.distinctID, event.name, &mixpanel.Event{
				Properties: event.properties,
				Timestamp:  &event.timestamp,
			},
		)
		if mixpanelErr != nil {
			err = multierror.Append(err, mixpanelErr)
		}
	}

	if c.config.IsAmplitudeEnabled && c.amplitudeClient != nil {
		// Added prefix to follow naming convention and differentiate between agent and internal service
		if c.config.IsInternalService {
			event.name = "Live Insights - " + event.name
		} else {
			event.name = "Live Insights Agent - " + event.name
		}

		// Postman's property naming convention in Amplitude is snake case. So convert event.properties keys to snake case.
		properties := map[string]any{}
		for k, v := range event.properties {
			properties[strcase.ToSnake(k)] = v
		}

		c.amplitudeClient.Track(amplitude.Event{
			UserID:          event.distinctID,
			EventType:       event.name,
			EventProperties: properties,
			EventOptions:    c.amplitudeAppInfo,
		})
	}

	return errors.Wrapf(
		err,
		"failed to send analytics tracking event '%s' for distinct id %s",
		event.name,
		event.distinctID,
	)
}

func (c clientImpl) Track(distinctID string, name string, properties map[string]any) error {
	return c.TrackEvent(NewEvent(distinctID, name, properties))
}

func (c clientImpl) Close() error {
	var err error

	if c.segmentClient != nil {
		err = c.segmentClient.Close()
	}

	if c.amplitudeClient != nil {
		c.amplitudeClient.Shutdown()
	}

	return err
}

func newMixpanelClient(config Config) (mixpanel.Mixpanel, error) {
	if !config.IsMixpanelEnabled {
		return nil, nil
	}

	const (
		defaultUrl = "https://api.mixpanel.com"
	)

	if config.MixpanelToken == "" {
		return nil, errors.New("unable to construct new mixpanel client. token cannot be empty")
	}

	mixpanelURL := config.MixpanelEndpoint
	if mixpanelURL == "" {
		mixpanelURL = defaultUrl
	}

	if config.MixpanelSecret != "" {
		return mixpanel.NewWithSecret(config.MixpanelToken, config.MixpanelSecret, mixpanelURL), nil
	}

	return mixpanel.New(config.MixpanelToken, mixpanelURL), nil
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

func newSegmentClient(config Config) (segment.Client, error) {
	if !config.IsSegmentEnabled {
		return nil, nil
	}

	appInfo := segment.AppInfo{
		Name:      config.App.Name,
		Version:   config.App.Version,
		Build:     config.App.Build,
		Namespace: config.App.Namespace,
	}

	analyticsConfig := segment.Config{
		DefaultContext: &segment.Context{
			App: appInfo,
		},
		Endpoint: provideSegmentEndpoint(config.SegmentEndpoint),
		Logger:   provideLogger(config.IsLoggingEnabled),
	}

	if config.BatchSize > 0 {
		analyticsConfig.BatchSize = config.BatchSize
	}

	if config.WriteKey == "" {
		return nil, errors.New("unable to construct new segment client. write key cannot be empty")
	}

	return segment.NewWithConfig(config.WriteKey, analyticsConfig)
}

// Returns the default integrations to use for the analytics client based on the default integrations set in the input config.
// If the config does not specify any default integrations, then all integrations are enabled by default.
func provideDefaultIntegrations(config Config) segment.Integrations {
	if len(config.DefaultIntegrations) == 0 {
		return segment.NewIntegrations().EnableAll()
	}

	integrations := segment.NewIntegrations()

	for integrationName, enabled := range config.DefaultIntegrations {
		integrations = integrations.Set(integrationName, enabled)
	}

	return integrations
}

// Returns the logger to use for the segment client if logging is enabled. Otherwise, returns nil.
func provideLogger(isLoggingEnabled bool) segment.Logger {
	if isLoggingEnabled {
		return &analyticsLogger{}
	}

	return &disabledLogger{}
}

// Returns the logger to use for the Amplitude client if logging is enabled.
func provideAmplitudeLogger(isLoggingEnabled bool) amplitude.Logger {
	if isLoggingEnabled {
		return &amplitudeLogger{}
	}

	return &disabledAmplitudeLogger{}
}

// Returns the input endpoint given it is not empty. Otherwise, returns the default endpoint for segment.
func provideSegmentEndpoint(endpoint string) string {
	if endpoint == "" {
		return segment.DefaultEndpoint
	}

	return endpoint
}

// A custom segment logger that logs to glog.
type analyticsLogger struct{}

func (d analyticsLogger) Logf(format string, args ...any) {
	glog.Infof(format, args...)
}

func (d analyticsLogger) Errorf(format string, args ...interface{}) {
	glog.Errorf(format, args...)
}

// A custom segment logger that does nothing.
// This is used when logging is disabled as the segment client requires a logger (the client uses its own default logger even when none is specified).
type disabledLogger struct{}

func (d disabledLogger) Logf(format string, args ...any) {
	// Do nothing.
}

func (d disabledLogger) Errorf(format string, args ...interface{}) {
	// Do nothing.
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
