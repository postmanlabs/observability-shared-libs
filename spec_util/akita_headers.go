package spec_util

// Special header values used by the CLI client.
// Previously this file contained a filter to remove them from a trace,
// unelss x-akita-dogfood is set.
const (
	XAkitaCLIGitVersion = "x-akita-cli-git-version"
	XAkitaRequestID     = "x-akita-request-id"
	XAkitaDogfood       = "x-akita-dogfood"
)
