package version_names

import (
	"fmt"
	"strings"

	"github.com/akitasoftware/akita-libs/tags"
)

type VersionName = string

const (
	// Unreserved names.  Users are also allowed to use these version names.
	XAkitaStableVersionName VersionName = "stable"

	// Reserved names.  Users are not allowed to use these version names.
	XAkitaLatestVersionName VersionName = "latest"

	// Reserved prefixes.  Users are not allowed to use version names that
	// start with these.
	XAkitaReservedVersionNamePrefix string = "x-akita"
)

// Determines whether a version is reserved for Akita internal use.
func IsReservedVersionName(k VersionName) bool {
	s := strings.ToLower(k)
	isReservedConstant := strings.EqualFold(s, XAkitaLatestVersionName)
	hasReservedPrefix := strings.HasPrefix(k, XAkitaReservedVersionNamePrefix)
	return isReservedConstant || hasReservedPrefix
}

// Produces the version name for the latest model that aggregates all models for
// a deployment and a source.
func GetBigSpecVersionName(source tags.Source, deployment string) VersionName {
	// XXX If source or deployment contain colons, this can result in collisions.
	// For example, ("foo:bar", "baz") and ("foo", "bar:baz") will both result in
	// "x-akita-big-model:foo:bar:baz".
	builder := new(strings.Builder)
	builder.WriteString(fmt.Sprintf("%s-big-model:", XAkitaReservedVersionNamePrefix))
	builder.WriteString(source)
	if deployment != "" {
		builder.WriteString(":")
		builder.WriteString(deployment)
	}
	return builder.String()
}

// Produces the version name for the latest "large" model that was precomputed
// for a deployment.
func GetLargeModelVersionName(deployment string) VersionName {
	return fmt.Sprintf("%s-large-model:%s", XAkitaReservedVersionNamePrefix, deployment)
}

// Produces the version name for the latest "large diffing model" to be used for
// diffing against the "small diffing model" in a deployment.
func GetLargeDiffingModelVersionName(deployment string) VersionName {
	return fmt.Sprintf("%s-large-diffing-model:%s", XAkitaReservedVersionNamePrefix, deployment)
}

// Produces the version name for the latest "small diffing model" to be used for
// diffing against the "large diffing model" in a deployment.
func GetSmallDiffingModelVersionName(deployment string) VersionName {
	return fmt.Sprintf("%s-small-diffing-model:%s", XAkitaReservedVersionNamePrefix, deployment)
}
