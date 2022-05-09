package maps

import (
	"testing"

	"akitasoftware.com/superstar/lib/stdlib/math"
	"github.com/stretchr/testify/assert"
)

func TestBasicOps(t *testing.T) {
	m := Map[string, int]{}

	m.Upsert("foo", 1, math.Add[int])
	assert.Equal(t, Map[string, int]{"foo": 1}, m)

	m.Add(Map[string, int]{"foo": 2, "bar": 1}, math.Add[int])
	assert.Equal(t, Map[string, int]{"foo": 3, "bar": 1}, m)
}
