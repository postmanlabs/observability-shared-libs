package tags

import (
	"strings"

	"github.com/google/martian/v3/tags"
	"github.com/pkg/errors"
)

type Key = tags.Key

// Returns a map from parsing a list of "key=value" pairs.
// Produces an error if any element of the list is improperly formatted,
// or if any key is given more than once.
// The caller must emit an appropriate warning if any keys are reserved.
func FromPairs(pairs []string) (map[Key]string, error) {
	results := make(map[Key]string, len(pairs))
	for _, p := range pairs {
		parts := strings.Split(p, "=")
		if len(parts) != 2 {
			return nil, errors.Errorf("%s is not a valid key=value format", p)
		}

		k, v := Key(parts[0]), parts[1]
		if _, ok := results[k]; ok {
			return nil, errors.Errorf("tag with key %s specified more than once", k)
		}

		results[k] = v
	}
	return results, nil
}
