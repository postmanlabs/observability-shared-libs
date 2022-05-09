package sets

import (
	"encoding/json"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type Set[T constraints.Ordered] map[T]struct{}

func NewSet[T constraints.Ordered](vs ...T) Set[T] {
	s := make(Set[T], len(vs))
	for _, v := range vs {
		s.Insert(v)
	}
	return s
}

func (s Set[T]) Contains(v T) bool {
	return s.ContainsAny(v)
}

func (s Set[T]) ContainsAny(vs ...T) bool {
	for _, v := range vs {
		_, exists := s[v]
		if exists {
			return true
		}
	}
	return false
}

func (s Set[T]) ContainsAll(vs ...T) bool {
	for _, v := range vs {
		_, exists := s[v]
		if !exists {
			return false
		}
	}
	return true
}

func (s Set[T]) Insert(vs ...T) {
	for _, v := range vs {
		s[v] = struct{}{}
	}
}

func (s Set[T]) Delete(vs ...T) {
	for _, v := range vs {
		delete(s, v)
	}
}

func (s Set[T]) Union(other Set[T]) {
	for k := range other {
		s.Insert(k)
	}
}

func (s Set[T]) Intersect(other Set[T]) {
	var toDelete []T
	for k := range s {
		if _, exists := other[k]; !exists {
			toDelete = append(toDelete, k)
		}
	}
	for _, k := range toDelete {
		delete(s, k)
	}
}

func (s Set[T]) MarshalJSON() ([]byte, error) {
	slice := make([]T, 0, len(s))
	for k := range s {
		slice = append(slice, k)
	}
	slices.Sort(slice)
	return json.Marshal(slice)
}

func (s *Set[T]) UnmarshalJSON(text []byte) error {
	var slice []T
	if err := json.Unmarshal(text, &slice); err != nil {
		return errors.Wrapf(err, "failed to unmarshal stringset")
	}
	*s = make(Set[T], len(slice))
	for _, elt := range slice {
		(*s)[elt] = struct{}{}
	}
	return nil
}

func (s Set[T]) Clone() Set[T] {
	return maps.Clone(s)
}

// AsSlice returns the set as a sorted slice.
func (s Set[T]) AsSlice() []T {
	rv := make([]T, 0, len(s))
	for x, _ := range s {
		rv = append(rv, x)
	}
	slices.Sort(rv)
	return rv
}
