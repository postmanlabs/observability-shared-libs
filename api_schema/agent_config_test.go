package api_schema

import (
	"regexp"
	"testing"

	"github.com/akitasoftware/go-utils/slices"
	"github.com/stretchr/testify/assert"
)

func TestFieldRedactionConfigEquals(t *testing.T) {
	c1 := &FieldRedactionConfig{
		FieldNames: []string{"foo", "bar", "baz"},
		FieldNameRegexps: slices.Map(
			[]string{
				"foo",
				"[ba]r",
				"^baz$",
				"qu+x",
			},
			regexp.MustCompile,
		),
	}

	c1Clone := c1.Clone()

	c1Copy := &FieldRedactionConfig{
		FieldNames: []string{"foo", "bar", "baz"},
		FieldNameRegexps: slices.Map(
			[]string{
				"foo",
				"[ba]r",
				"^baz$",
				"qu+x",
			},
			regexp.MustCompile,
		),
	}

	// Different from c1: has a different field name.
	c2 := &FieldRedactionConfig{
		FieldNames: []string{"foo", "bar", "qux"},
		FieldNameRegexps: slices.Map(
			[]string{
				"foo",
				"[ba]r",
				"^baz$",
				"qu+x",
			},
			regexp.MustCompile,
		),
	}

	// Different from c2: has a different field name.
	//
	// Different from c3: has a different field name regexp (although functionally
	// the same).
	c3 := &FieldRedactionConfig{
		FieldNames: []string{"foo", "bar", "baz"},
		FieldNameRegexps: slices.Map(
			[]string{
				"foo",
				"[ba]r",
				"^baz$",
				"quu*x",
			},
			regexp.MustCompile,
		),
	}

	assert.True(t, c1.Equals(c1))

	assert.True(t, c1.Equals(c1Clone))
	assert.True(t, c1Clone.Equals(c1))

	assert.True(t, c1.Equals(c1Copy))
	assert.True(t, c1Copy.Equals(c1))

	assert.False(t, c1.Equals(c2))
	assert.False(t, c2.Equals(c1))

	assert.False(t, c1.Equals(c3))
	assert.False(t, c3.Equals(c1))

	assert.False(t, c2.Equals(c3))
	assert.False(t, c3.Equals(c2))
}
