package tags

import (
	"encoding/json"
	"sort"
)

type ValueSet map[Value]struct{}

func NewValueSet(vs ...Value) ValueSet {
	rv := make(ValueSet, len(vs))
	for _, v := range vs {
		rv[v] = struct{}{}
	}
	return rv
}

// Marshals ValueSet as a sorted list.
func (vs ValueSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(vs.AsSlice())
}

// Parses ValueSet from a list, removing duplicates.
func (vs *ValueSet) UnmarshalJSON(b []byte) error {
	var slice []Value
	if err := json.Unmarshal(b, &slice); err != nil {
		return err
	}

	*vs = NewValueSet(slice...)

	return nil
}

// Adds a value to this set.
func (vs ValueSet) Add(v Value) {
	vs[v] = struct{}{}
}

// Adds all values from other to v.
func (vs ValueSet) AddAll(other ValueSet) {
	for v, _ := range other {
		vs.Add(v)
	}
}

// Removes values in vs that are not in other.
func (vs ValueSet) Intersect(other ValueSet) {
	for v, _ := range vs {
		if _, exists := other[v]; !exists {
			delete(vs, v)
		}
	}
}

// Adds values to vs that are in other.
func (vs ValueSet) Union(other ValueSet) {
	for v, _ := range other {
		vs.Add(v)
	}
}

// Returns the smallest value in the set.  If the set is empty, returns
// exists == false.
func (vs ValueSet) GetFirst() (v Value, exists bool) {
	if len(vs) == 0 {
		return "", false
	}
	return vs.AsSlice()[0], true
}

// Returns a sorted slice of values.
func (vs ValueSet) AsSlice() []Value {
	slice := make([]Value, 0, len(vs))
	for v, _ := range vs {
		slice = append(slice, v)
	}
	sort.Strings(slice)
	return slice
}

func (vs ValueSet) Clone() ValueSet {
	rv := make(ValueSet, len(vs))
	for v, _ := range vs {
		rv.Add(v)
	}
	return rv
}
