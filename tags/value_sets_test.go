package tags

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	vs := NewValueSet()
	vs.Add("v")
	assert.Equal(t, NewValueSet("v"), vs, "single value")

	vs.Add("v2")
	assert.Equal(t, NewValueSet("v", "v2"), vs, "two values")
}

func TestValueSetUnion(t *testing.T) {
	testCases := []struct {
		name     string
		left     ValueSet
		right    ValueSet
		expected ValueSet
	}{
		{
			name:     "union with duplicates",
			left:     NewValueSet("v1", "v2", "v3"),
			right:    NewValueSet("v2", "v3", "v4"),
			expected: NewValueSet("v1", "v2", "v3", "v4"),
		},
		{
			name:     "empty sets remain empty",
			left:     NewValueSet(),
			right:    NewValueSet(),
			expected: NewValueSet(),
		},
		{
			name:     "empty set is idempotent",
			left:     NewValueSet("v1", "v2"),
			right:    NewValueSet(),
			expected: NewValueSet("v1", "v2"),
		},
		{
			name:     "union no duplicates",
			left:     NewValueSet("v1"),
			right:    NewValueSet("v1"),
			expected: NewValueSet("v1"),
		},
	}

	for _, tc := range testCases {
		left := tc.left.Clone()
		left.Union(tc.right)
		assert.Equal(t, tc.expected, left, "[left]"+tc.name)

		right := tc.right.Clone()
		right.Union(tc.left)
		assert.Equal(t, tc.expected, right, "[right] "+tc.name)
	}
}

func TestValueSetIntersect(t *testing.T) {
	testCases := []struct {
		name     string
		left     ValueSet
		right    ValueSet
		expected ValueSet
	}{
		{
			name:     "intersect with duplicates",
			left:     NewValueSet("v1", "v1", "v2"),
			right:    NewValueSet("v1", "v1", "v1"),
			expected: NewValueSet("v1"),
		},
		{
			name:     "empty tags remain empty",
			left:     NewValueSet(),
			right:    NewValueSet(),
			expected: NewValueSet(),
		},
		{
			name:     "empty list is an annihilator",
			left:     NewValueSet("v1", "v1", "v2"),
			right:    NewValueSet(),
			expected: NewValueSet(),
		},
	}

	for _, tc := range testCases {
		left := tc.left.Clone()
		left.Intersect(tc.right)
		assert.Equal(t, tc.expected, left, "[left]"+tc.name)

		right := tc.right.Clone()
		right.Intersect(tc.left)
		assert.Equal(t, tc.expected, right, "[right] "+tc.name)
	}
}

func TestGetFirst(t *testing.T) {
	f, exists := NewValueSet("v2", "v1").GetFirst()
	assert.True(t, exists, "get first exists")
	assert.Equal(t, "v1", f, "get first")

	_, exists = NewValueSet().GetFirst()
	assert.False(t, exists, "get first doesn't exist")
}

func TestAsSlice(t *testing.T) {
	assert.Equal(t, []Value{}, NewValueSet().AsSlice(), "as empty slice")
	assert.Equal(t, []Value{"v1", "v2"}, NewValueSet("v2", "v1").AsSlice(), "as slice")
}

func TestValueSetJSON(t *testing.T) {
	testCases := []struct {
		name string
		vs   ValueSet
	}{
		{
			name: "value set with values",
			vs:   NewValueSet("v1", "v2"),
		},
		{
			name: "empty value set",
			vs:   NewValueSet(),
		},
	}

	for _, tc := range testCases {
		// Marshal to bytes
		bs, err := json.Marshal(tc.vs)
		assert.NoError(t, err, tc.name+": marshal")

		// Parse from bytes
		var parsed ValueSet
		err = json.Unmarshal(bs, &parsed)
		assert.NoError(t, err, tc.name+": unmarshal")

		// Check that we got what we started with
		assert.Equal(t, tc.vs, parsed, tc.name+": round trip")
	}
}
