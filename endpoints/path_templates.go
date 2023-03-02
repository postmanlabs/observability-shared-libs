package endpoints

import "strings"

// A path template whose parameter names have not been normalized.
type PathTemplate string

// Splits the path template into its path components. For example,
// "/v1/{foo}/bar".GetComponents() returns {"", "v1", "{foo}", "bar"}.
func (p PathTemplate) GetComponents() []string {
	return strings.Split(string(p), "/")
}

func (p PathTemplate) String() string {
	return string(p)
}
