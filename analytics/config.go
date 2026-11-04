package analytics

import "time"

type Config struct {
	App AppInfo `yaml:"app"`

	// Toggle for whether client is used in agent or internal service
	IsInternalService bool `yaml:"is_internal_service"`

	// Toggle for sending events to Amplitude
	IsAmplitudeEnabled bool `yaml:"amplitude_enabled"`

	// Separate config for amplitude client
	AmplitudeConfig AmplitudeConfig `yaml:"amplitude"`

	// Toggle for sending events to Segment
	IsSegmentEnabled bool `yaml:"segment_enabled"`

	// Separate config for segment client
	SegmentConfig SegmentConfig `yaml:"segment"`
}

// Data pertaining to the application such as name, version, and build
// If set, the specified values will be added globally to each event context
type AppInfo struct {
	Name      string `yaml:"name"`
	Version   string `yaml:"version"`
	Build     string `yaml:"build"`
	Namespace string `yaml:"namespace"`
}

type AmplitudeConfig struct {
	// Amplitude API Key
	AmplitudeAPIKey string `yaml:"amplitude_api_key"`

	// Amplitude endpoint. Empty means Amplitude's default US endpoint for the selected mode.
	// A URL with no path (such as "https://api.eu.amplitude.com") gets the mode's API path:
	// "/batch" with batching, "/2/httpapi" without. A URL with a path is used as is.
	AmplitudeEndpoint string `yaml:"amplitude_endpoint"`

	// Toggle for logging sent events
	IsLoggingEnabled bool `yaml:"logging_enabled"`

	// Toggle for batching events (Amplitude's Batch Event Upload API instead of the HTTP V2 API).
	// An endpoint with a path must then be a batch endpoint.
	IsBatchingEnabled bool `yaml:"batching_enabled"`

	// The maximum number of events to send in a single batch
	FlushQueueSize int `yaml:"flush_queue_size"`
}

type SegmentConfig struct {
	// Segment Write Key
	SegmentAPIKey string `yaml:"segment_api_key"`

	// Segment endpoint
	SegmentEndpoint string `yaml:"segment_endpoint"`

	// Toggle for logging sent events
	IsLoggingEnabled bool `yaml:"logging_enabled"`

	// The interval at which to flush the queue
	FlushInterval time.Duration `yaml:"flush_interval"`

	// The maximum number of events to send in a single batch
	BatchSize int `yaml:"batch_size"`

	// Toggle for verbose logging
	IsVerboseLoggingEnabled bool `yaml:"verbose_logging_enabled"`
}
