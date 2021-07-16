package version_names

import (
	"fmt"
	"strings"

	"github.com/akitasoftware/akita-libs/tags"
)

type VersionName = string

const (
	XAkitaLatestVersionName         VersionName = "latest"
	XAkitaStableVersionName         VersionName = "stable"
	XAkitaReservedVersionNamePrefix string      = "x-akita"
)

// Determines whether a version is reserved for Akita internal use.
func IsReservedVersionName(k VersionName) bool {
	s := strings.ToLower(k)
	isReservedConstant := strings.EqualFold(s, XAkitaLatestVersionName) || strings.EqualFold(s, XAkitaStableVersionName)
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