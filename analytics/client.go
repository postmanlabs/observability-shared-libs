package analytics

import (
	"github.com/dukex/mixpanel"
	"github.com/golang/glog"
	multierror "github.com/hashicorp/go-multierror"
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
	config        Config
	segmentClient segment.Client

	// TODO: Remove Mixpanel once we've confirmed that Segment is working.
	mixpanelClient mixpanel.Mixpanel
}

func NewClient(config Config) (Client, error) {
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
		return nil, errors.New("unable to construct new analytics client. write key cannot be empty")
	}

	segmentClient, err := segment.NewWithConfig(config.WriteKey, analyticsConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create segment client")
	}

	mixpanelClient, err := newMixpanelClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create mixpanel client")
	}

	return &clientImpl{
		config:         config,
		segmentClient:  segmentClient,
		mixpanelClient: mixpanelClient,
	}, nil
}

func (c clientImpl) TrackEvent(event *Event) error {
	var err error

	segmentErr := c.segmentClient.Enqueue(
		segment.Track{
			UserId:     event.distinctID,
			Event:      event.name,
			Properties: event.properties,
			Timestamp:  event.timestamp,
		},
	)

	if segmentErr != nil {
		err = multierror.Append(err, segmentErr)
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
	return c.segmentClient.Close()
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

// Returns the logger to use for the segment client if logging is enabled. Otherwise, returns nil.
func provideLogger(isLoggingEnabled bool) *analyticsLogger {
	if isLoggingEnabled {
		return &analyticsLogger{}
	}

	return nil
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
