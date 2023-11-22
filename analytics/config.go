package analytics

type Config struct {
	// Toggle for additionally sending all events to Mixpanel
	IsSegmentEnabled bool `yaml:"segment_enabled"`

	// The key used to identify the Segment source to use
	WriteKey string `yaml:"segment_write_key"`

	// The endpoint to which the segment client connects to send events
	SegmentEndpoint string `yaml:"segment_endpoint"`

	App AppInfo `yaml:"app"`

	// Toggle for logging sent events
	IsLoggingEnabled bool `yaml:"logging_enabled"`

	// TODO: This should be removed once we have fully migrated over to Segment
	// Toggle for additionally sending all events to Mixpanel
	IsMixpanelEnabled bool `yaml:"mixpanel_enabled"`

	// The Mixpanel token.
	MixpanelToken string `yaml:"mixpanel_token"`

	// The Mixpanel endpoint
	MixpanelEndpoint string `yaml:"mixpanel_endpoint"`

	// If present, adds a Basic authentication header with secret as the
	// username and the empty string as the password.
	MixpanelSecret string `yaml:"mixpanel_secret"`

	// Disable batching (used for CLI) by setting this parameter to 1
	BatchSize int `yaml:"batch_size"`

	// Sets the default destination types that events will be sent to.
	// If not set, events will be sent to all destinations.
	DefaultIntegrations map[string]bool `yaml:"default_integrations"`

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
