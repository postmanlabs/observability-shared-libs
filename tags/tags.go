package tags

import (
	"strings"

	"github.com/pkg/errors"
)

type Key string
type Value = string

// Maps tags to sets of values.
type Tags map[Key]ValueSet

// Sets the given key to the given values.  Values are copied.
func (t Tags) Set(key Key, values []string) {
	t[key] = NewValueSet(values...)
}

// Sets the given key to the given value.
func (t Tags) SetSingleton(key Key, value string) {
	t[key] = NewValueSet(value)
}

// Adds a value to the given key.
func (t Tags) Add(key Key, value string) {
	if values, exists := t[key]; !exists {
		t[key] = NewValueSet(value)
	} else {
		values.Add(value)
	}
}

// SetAll copies all values from t2 into t. If a key exists in both sets of tags, the
// value in t is overwritten with that in t2.
func (t Tags) SetAll(t2 Tags) {
	for key, values := range t2 {
		t.Set(key, values.AsSlice())
	}
}

// Removes any tag from t that doesn't exist in t2.  For any key k in both t
// and t2, t[k] is remapped to the intersection of their values.  If the
// intersection is empty, then k is removed.
func (t Tags) Intersect(t2 Tags) {
	for key, values := range t {
		otherValues, ok := t2[key]

		// Remove keys not in t2.
		if !ok {
			delete(t, key)
		}

		// Remove values not in t2.
		values.Intersect(otherValues)

		// If there are no values remaining, remove the tag.
		if len(values) == 0 {
			delete(t, key)
		}
	}
}

// Copies tags from t2 that don't exist in t.  For any key k in both t and t2,
// t[k] is remapped to the union of their values.  Does not preserve duplicates
// or maintain list order.
func (t Tags) Union(t2 Tags) {
	for otherTag, otherValues := range t2 {
		if values, exists := t[otherTag]; !exists {
			t[otherTag] = otherValues.Clone()
		} else {
			values.Union(otherValues)
		}
	}
}

func (t Tags) Clone() Tags {
	rv := make(Tags, len(t))
	for t, vs := range t {
		rv[t] = vs.Clone()
	}
	return rv
}

// Returns a new tags map with a single value for each tag.  If more than one
// value was present for a given tag, returns the first value in the list.
// If there are no values in the list, the tag is removed.
func (t Tags) AsSingletonTags() SingletonTags {
	rv := make(SingletonTags, len(t))
	for tag, values := range t {
		if v, exists := values.GetFirst(); exists {
			rv[tag] = v
		}
	}
	return rv
}

// Returns a Tags from parsing a list of "key=value" pairs.
// Produces an error if any element of the list is improperly formatted.
// The caller must emit an appropriate warning if any keys are reserved.
func FromPairsMultivalue(pairs []string) (Tags, error) {
	results := make(Tags, len(pairs))
	for _, p := range pairs {
		parts := strings.Split(p, "=")
		if len(parts) != 2 {
			return nil, errors.Errorf("%s is not a valid key=value format", p)
		}

		k, v := Key(parts[0]), parts[1]
		results.Add(k, v)
	}
	return results, nil
}
