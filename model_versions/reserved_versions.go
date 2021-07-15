package model_versions

import "strings"

const (
	XAkitaLatestModelVersion = "latest"
)

// Determines whether a version is reserved for Akita internal use.
func IsReservedModelVersion(k string) bool {
	s := strings.ToLower(string(k))
	return strings.EqualFold(s, XAkitaLatestModelVersion)
}
