package tags

import "github.com/pkg/errors"

// SingletonTags maps tags to single values.
type SingletonTags map[Key]string

func (ts SingletonTags) AsTags() Tags {
	tags := make(Tags, len(ts))
	for t, v := range ts {
		tags.SetSingleton(t, v)
	}
	return tags
}

// FromPairsSingleton returns a map from parsing a list of "key=value" pairs.
// Produces an error if any element of the list is improperly formatted,
// or if any key is given more than once.
// The caller must emit an appropriate warning if any keys are reserved.
func FromPairs(pairs []string) (SingletonTags, error) {
	tags, err := FromPairsMultivalue(pairs)
	if err != nil {
		return nil, err
	}

	results := make(SingletonTags, len(tags))
	for k, vs := range tags {
		vSlice := vs.AsSlice()
		if len(vSlice) == 0 {
			continue
		} else if len(vSlice) > 1 {
			return nil, errors.Errorf("tag with key %s specified more than once", k)
		}
		results[k] = vSlice[0]
	}

	return results, nil
}
