package analytics

import (
	"github.com/dukex/mixpanel"
	"github.com/golang/glog"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/segmentio/analytics-go/v3"
)

type Client interface {
	// Sends the given tracking event to Segment and Mixpanel (if enabled).
	Track(distinctID string, event *Event) error
}

type clientImpl struct {
	config        Config
	segmentClient analytics.Client

	// TODO: Remove Mixpanel once we've confirmed that Segment is working.
	mixpanelClient mixpanel.Mixpanel
}

func NewClient(config Config, mixpanelClient mixpanel.Mixpanel) (Client, error) {
	analyticsConfig := analytics.Config{
		DefaultContext: &analytics.Context{
			App: config.AppInfo,
		},
		Endpoint: provideSegmentEndpoint(config.SegmentEndpoint),
		Logger:   provideLogger(config.IsLoggingEnabled),
	}

	if config.WriteKey == "" {
		return nil, errors.New("unable to construct new analytics client. write key cannot be empty")
	}

	segmentClient, err := analytics.NewWithConfig(config.WriteKey, analyticsConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create segment client")
	}

	return &clientImpl{
		config:         config,
		segmentClient:  segmentClient,
		mixpanelClient: mixpanelClient,
	}, nil
}

func (c clientImpl) Track(distinctID string, event *Event) error {
	var err error

	segmentErr := c.segmentClient.Enqueue(
		analytics.Track{
			UserId:     distinctID,
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
			distinctID, event.name, &mixpanel.Event{
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
		distinctID,
	)
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
		return analytics.DefaultEndpoint
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
