package analytics

type Config struct {
	App AppInfo `yaml:"app"`

	// Toggle for logging sent events
	IsLoggingEnabled bool `yaml:"logging_enabled"`

	// Disable batching (used for CLI) by setting this parameter to 1
	BatchSize int `yaml:"batch_size"`

	// Toggle for sending events to Amplitude
	IsAmplitudeEnabled bool `yaml:"amplitude_enabled"`

	// Toggle for whether client is used in agent or internal service
	IsInternalService bool `yaml:"is_internal_service"`

	// Separate config for amplitude client
	AmplitudeConfig AmplitudeConfig `yaml:"amplitude"`
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

	// Amplitude endpoint
	AmplitudeEndpoint string `yaml:"amplitude_endpoint"`

	// Toggle for logging sent events
	IsLoggingEnabled bool `yaml:"logging_enabled"`

	// Toggle for batching events. Make sure to set batch amplitude endpoint
	IsBatchingEnabled bool `yaml:"batching_enabled"`

	// The maximum number of events to send in a single batch
	FlushQueueSize int `yaml:"flush_queue_size"`
}
