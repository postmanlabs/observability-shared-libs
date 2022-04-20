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

func TestSetSingleton(t *testing.T) {
	tagMap := Tags{}

	// Add new tag
	tagMap.SetSingleton("key", "v1")
	assert.Equal(t, Tags{"key": NewValueSet("v1")}, tagMap, "insert new value")

	// Overwrite existing tag
	tagMap.SetSingleton("key", "v2")
	assert.Equal(t, Tags{"key": NewValueSet("v2")}, tagMap, "overwrite value")
}

func TestTagAdd(t *testing.T) {
	tagMap := Tags{}

	// Add new tag
	tagMap.Add("key", "v1")
	assert.Equal(t, Tags{"key": NewValueSet("v1")}, tagMap, "add new tag")

	// Add existing value to existing tag
	tagMap.Add("key", "v1")
	assert.Equal(t, Tags{"key": NewValueSet("v1")}, tagMap, "add existing value")

	// Add new value to existing tag
	tagMap.Add("key", "v2")
	assert.Equal(t, Tags{"key": NewValueSet("v1", "v2")}, tagMap, "add new value")
}

func TestTagSetAll(t *testing.T) {
	testCases := []struct {
		name     string
		left     Tags
		right    Tags
		expected Tags
	}{
		{
			name:     "write to empty",
			left:     Tags{},
			right:    Tags{"k": NewValueSet("v")},
			expected: Tags{"k": NewValueSet("v")},
		},
		{
			name:     "overwrite when left is a subset of right",
			left:     Tags{"k1": NewValueSet("v2")},
			right:    Tags{"k1": NewValueSet("v"), "k2": NewValueSet("v2")},
			expected: Tags{"k1": NewValueSet("v"), "k2": NewValueSet("v2")},
		},
		{
			name:     "overwrite when right is a subset of left",
			left:     Tags{"k1": NewValueSet("v"), "k2": NewValueSet("v2")},
			right:    Tags{"k1": NewValueSet("v2")},
			expected: Tags{"k1": NewValueSet("v2"), "k2": NewValueSet("v2")},
		},
	}

	for _, tc := range testCases {
		tc.left.SetAll(tc.right)
		assert.Equal(t, tc.expected, tc.left, tc.name)
	}
}

func TestAsSingletonTags(t *testing.T) {
	testCases := []struct {
		name     string
		tags     Tags
		expected SingletonTags
	}{
		{
			name:     "tags are singletons",
			tags:     Tags{"k1": NewValueSet("v1"), "k2": NewValueSet("v2")},
			expected: SingletonTags{"k1": "v1", "k2": "v2"},
		},
		{
			name:     "tags have multiple values",
			tags:     Tags{"k1": NewValueSet("v11", "v12"), "k2": NewValueSet("v22", "v21")},
			expected: SingletonTags{"k1": "v11", "k2": "v21"},
		},
		{
			name:     "tags have empty sets",
			tags:     Tags{"k1": NewValueSet("v11", "v12"), "k2": NewValueSet()},
			expected: SingletonTags{"k1": "v11"},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, tc.tags.AsSingletonTags(), tc.name)
	}
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
