package version_names

import (
	"fmt"
	"strings"

	"github.com/akitasoftware/akita-libs/tags"
)

type VersionName = string

const (
	// Unreserved names.  Users are also allowed to use these version names.
	XAkitaStableVersionName         VersionName = "stable"

	// Reserved names.  Users are not allowed to use these version names.
	XAkitaLatestVersionName         VersionName = "latest"

	// Reserved prefixes.  Users are not allowed to use version names that
	// start with these.
	XAkitaReservedVersionNamePrefix string      = "x-akita"
)

// Determines whether a version is reserved for Akita internal use.
func IsReservedVersionName(k VersionName) bool {
	s := strings.ToLower(k)
	isReservedConstant := strings.EqualFold(s, XAkitaLatestVersionName)
	hasReservedPrefix := strings.HasPrefix(k, XAkitaReservedVersionNamePrefix)
	return isReservedConstant || hasReservedPrefix
}

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