package analytics

import (
	"net/url"

	"github.com/amplitude/analytics-go/amplitude"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

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

	amplitudeConfig.UseBatch = rawAmplitudeConfig.IsBatchingEnabled

	if rawAmplitudeConfig.AmplitudeEndpoint != "" {
		serverURL, err := amplitudeServerURL(rawAmplitudeConfig.AmplitudeEndpoint, amplitudeConfig.UseBatch)
		if err != nil {
			return nil, amplitude.EventOptions{}, errors.Wrap(err, "unable to construct new amplitude analytics client")
		}
		amplitudeConfig.ServerURL = serverURL
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

// The API paths the Amplitude SDK posts to, for an endpoint given without a path.
const (
	amplitudeHTTPAPIPath = "/2/httpapi"
	amplitudeBatchPath   = "/batch"
)

// Returns the URL the Amplitude SDK posts events to, for the configured endpoint.
//
// An endpoint with no path (such as "https://api.eu.amplitude.com") gets the API path of the
// selected mode: "/batch" with batching, "/2/httpapi" without. An endpoint with a path is used
// as is, so a full URL (such as "https://api.eu.amplitude.com/2/httpapi") or a relay with its
// own path keeps working.
func amplitudeServerURL(endpoint string, useBatch bool) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", errors.Wrapf(err, "invalid amplitude endpoint %q", endpoint)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.Errorf("invalid amplitude endpoint %q: the scheme must be http or https", endpoint)
	}
	if u.Host == "" {
		return "", errors.Errorf("invalid amplitude endpoint %q: no host", endpoint)
	}

	if u.Path == "" || u.Path == "/" {
		if useBatch {
			u.Path = amplitudeBatchPath
		} else {
			u.Path = amplitudeHTTPAPIPath
		}
		u.RawPath = ""
	}

	return u.String(), nil
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

var _ amplitude.Logger = &amplitudeLogger{}

func (d amplitudeLogger) Debugf(format string, args ...any) {
	glog.V(1).Infof(format, args...)
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

var _ amplitude.Logger = &disabledAmplitudeLogger{}

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
