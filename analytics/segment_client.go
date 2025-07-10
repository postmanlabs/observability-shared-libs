package analytics

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/segmentio/analytics-go/v3"
)

func newSegmentClient(config Config) (analytics.Client, error) {
	if !config.IsSegmentEnabled {
		return nil, nil
	}

	rawSegmentConfig := config.SegmentConfig

	if rawSegmentConfig == (SegmentConfig{}) {
		return nil, errors.New("unable to construct new segment analytics client. segment config cannot be empty")
	}

	if rawSegmentConfig.SegmentAPIKey == "" {
		return nil, errors.New("unable to construct new segment analytics client. API key cannot be empty")
	}

	segmentConfig := analytics.Config{
		DefaultContext: &analytics.Context{
			App: analytics.AppInfo{
				Name:      config.App.Name,
				Version:   config.App.Version,
				Build:     config.App.Build,
				Namespace: config.App.Namespace,
			},
		},
	}

	segmentConfig.Logger = provideSegmentLogger(rawSegmentConfig.IsLoggingEnabled)

	if rawSegmentConfig.FlushInterval > 0 {
		segmentConfig.Interval = rawSegmentConfig.FlushInterval
	}

	if rawSegmentConfig.SegmentEndpoint != "" {
		segmentConfig.Endpoint = rawSegmentConfig.SegmentEndpoint
	}

	if rawSegmentConfig.IsVerboseLoggingEnabled {
		segmentConfig.Verbose = true
	}

	if rawSegmentConfig.BatchSize > 0 {
		segmentConfig.BatchSize = rawSegmentConfig.BatchSize
	}

	segmentClient, err := analytics.NewWithConfig(rawSegmentConfig.SegmentAPIKey, segmentConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create segment analytics client")
	}

	return segmentClient, nil
}

func provideSegmentLogger(isLoggingEnabled bool) analytics.Logger {
	if isLoggingEnabled {
		return &segmentLogger{}
	}

	return &disabledSegmentLogger{}
}

type segmentLogger struct{}

var _ analytics.Logger = &segmentLogger{}

func (d segmentLogger) Errorf(format string, args ...any) {
	glog.Errorf(format, args...)
}

func (d segmentLogger) Logf(format string, args ...any) {
	glog.Infof(format, args...)
}

type disabledSegmentLogger struct{}

var _ analytics.Logger = &disabledSegmentLogger{}

func (d disabledSegmentLogger) Errorf(format string, args ...any) {
	// Do nothing.
}

func (d disabledSegmentLogger) Logf(format string, args ...any) {
	// Do nothing.
}
