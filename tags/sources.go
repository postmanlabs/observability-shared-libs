package tags

type Source = string

// Valid values for the XAkitaSource tag.
const (
	// Designates a trace or spec that was generated from CI.
	CISource Source = "ci"

	// Designates a trace or spec that was derived from a staging or production
	// deployment.
	DeploymentSource Source = "deployment"

	// Designates a trace or spec that whas manually uploaded by the user.
	UploadedSource Source = "uploaded"

	// Designates a trace or spec that was manually created by the user. For
	// example, traces captured from network traffic using the CLI's `apidump`
	// command are tagged with this source.
	UserSource Source = "user"
)
