package tags

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	tagMap := Tags{}
	assert.Equal(t, Tags{}, tagMap, "empty tags maps are equivalent")

	tagMap.Set("key", []string{"v1", "v2"})
	assert.Equal(t, Tags{"key": NewValueSet("v1", "v2")}, tagMap, "insert two values")
}

func TestTagsUnion(t *testing.T) {
	testCases := []struct {
		name     string
		left     Tags
		right    Tags
		expected Tags
	}{
		{
			name: "union lists with duplicates",
			left: Tags{
				XAkitaServiceVersion: NewValueSet("v1", "v1"),
			},
			right: Tags{
				XAkitaServiceVersion: NewValueSet("v1", "v2", "v2"),
			},
			expected: Tags{
				XAkitaServiceVersion: NewValueSet("v1", "v2"),
			},
		},
		{
			name:     "empty tags remain empty",
			left:     Tags{},
			right:    Tags{},
			expected: Tags{},
		},
		{
			name: "empty list is idempotent",
			left: Tags{
				XAkitaServiceVersion: NewValueSet("v1", "v2"),
			},
			right: Tags{
				XAkitaServiceVersion: NewValueSet(),
			},
			expected: Tags{
				XAkitaServiceVersion: NewValueSet("v1", "v2"),
			},
		},
		{
			name: "missing tag is idempotent",
			left: Tags{
				XAkitaServiceVersion: NewValueSet("v1", "v2"),
			},
			right: Tags{},
			expected: Tags{
				XAkitaServiceVersion: NewValueSet("v1", "v2"),
			},
		},
		{
			name: "union no duplicates",
			left: Tags{
				XAkitaServiceVersion: NewValueSet("v1"),
			},
			right: Tags{
				XAkitaServiceVersion: NewValueSet("v1"),
			},
			expected: Tags{
				XAkitaServiceVersion: NewValueSet("v1"),
			},
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

func TestTagsIntersect(t *testing.T) {
	testCases := []struct {
		name     string
		left     Tags
		right    Tags
		expected Tags
	}{
		{
			name: "intersect lists with duplicates",
			left: Tags{
				XAkitaServiceVersion: NewValueSet("v1", "v1", "v2"),
			},
			right: Tags{
				XAkitaServiceVersion: NewValueSet("v1", "v1", "v1"),
			},
			expected: Tags{
				XAkitaServiceVersion: NewValueSet("v1"),
			},
		},
		{
			name:     "empty tags remain empty",
			left:     Tags{},
			right:    Tags{},
			expected: Tags{},
		},
		{
			name: "empty list is an annihilator",
			left: Tags{
				XAkitaServiceVersion: NewValueSet("v1", "v1", "v2"),
			},
			right: Tags{
				XAkitaServiceVersion: NewValueSet(),
			},
			expected: Tags{},
		},
		{
			name: "missing tag is an annihilator",
			left: Tags{
				XAkitaServiceVersion: NewValueSet("v1", "v1", "v2"),
			},
			right:    Tags{},
			expected: Tags{},
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

func TestTagsJSON(t *testing.T) {
	testCases := []struct {
		name string
		tags Tags
	}{
		{
			name: "tags with values",
			tags: Tags{
				"t1": NewValueSet("v1", "v2"),
				"t2": NewValueSet("v"),
			},
		},
	}

	for _, tc := range testCases {
		// Marshal to bytes
		bs, err := json.Marshal(tc.tags)
		assert.NoError(t, err, tc.name+": marshal")

		// Parse from bytes
		var parsed Tags
		err = json.Unmarshal(bs, &parsed)
		assert.NoError(t, err, tc.name+": unmarshal")

		// Check that we got what we started with
		assert.Equal(t, tc.tags, parsed, tc.name+": round trip")
	}
}
