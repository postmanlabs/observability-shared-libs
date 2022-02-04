package path_pattern

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// For patterns with trailing slashes, for which serialization/deserialization
// is not idempotent.
func TestParse(t *testing.T) {
	testCases := []struct {
		input    string
		expected Pattern
	}{
		{
			input: "/v1//foobar/{foobar}/",
			expected: Pattern{
				Val(""),
				Val("v1"),
				Val(""),
				Val("foobar"),
				Var("foobar"),
			},
		},
	}

	for _, c := range testCases {
		p := Parse(c.input)
		assert.Equal(t, c.expected, p, c.input)
	}
}

// For patterns without trailing slashes.  Also tests that parsing and then
// serializing the pattern returns the original string.
func TestParseAndString(t *testing.T) {
	testCases := []struct {
		input    string
		expected Pattern
	}{
		{
			input: "/v1/{my_arg_name}",
			expected: Pattern{
				Val(""),
				Val("v1"),
				Var("my_arg_name"),
			},
		},
		{
			input: "/v1/*/foobar",
			expected: Pattern{
				Val(""),
				Val("v1"),
				Wildcard{},
				Val("foobar"),
			},
		},
	}

	for _, c := range testCases {
		p := Parse(c.input)
		assert.Equal(t, c.expected, p, c.input)
		assert.Equal(t, c.input, p.String())
	}
}

func TestMatch(t *testing.T) {
	testCases := []struct {
		pattern     string
		target      string
		expectMatch bool
	}{
		{
			pattern:     "/v1/{my_arg_name}",
			target:      "/v1/foobar",
			expectMatch: true,
		},
		{
			pattern:     "/v1/{my_arg_name}",
			target:      "/v1/foobar/x",
			expectMatch: false,
		},
		{
			pattern:     "/v1/{my_arg_name}",
			target:      "/v1/foobar/",
			expectMatch: true,
		},
		{
			pattern:     "/v1/{my_arg_name}/",
			target:      "/v1/foobar",
			expectMatch: true,
		},
		{
			pattern:     "/v1/{my_arg_name}",
			target:      "/v1/{my_old_arg_name}",
			expectMatch: true,
		},
		{
			pattern:     "/v1/*/{my_arg_name}",
			target:      "/v1/foo/bar",
			expectMatch: true,
		},
		{
			pattern:     "/v1/*/{my_arg_name}",
			target:      "/v1/{foo_param}/bar",
			expectMatch: true,
		},
		{
			pattern:     "/v1/^/{my_arg_name}",
			target:      "/v1/foo/bar",
			expectMatch: false,
		},
		{
			pattern:     "/v1/{my_arg_name}",
			target:      "/{arg1}/foobar/",
			expectMatch: false,
		},
		{
			// Variable cannot match empty component.
			pattern:     "/v1/{my_arg_name}",
			target:      "/v1/",
			expectMatch: false,
		},
		{
			// Target does not have enough components to match.
			pattern:     "/v1/{my_arg_name}",
			target:      "/v1",
			expectMatch: false,
		},
		{
			pattern:     "/v1/**",
			target:      "/v1/foo/bar/baz",
			expectMatch: true,
		},
		{
			pattern:     "/v1/**/baz",
			target:      "/v1/foo/bar/baz",
			expectMatch: true,
		},
		{
			pattern:     "/v1/**/baz/bar",
			target:      "/v1/foo/bar/baz",
			expectMatch: false,
		},
		{
			pattern:     "/v1/foo/;|{}/bar",
			target:      "/v1/foo/bar/baz",
			expectMatch: false,
		},
		{
			pattern:     "/v1/foo/b.r/baz",
			target:      "/v1/foo/bar/baz",
			expectMatch: false,
		},
	}

	for _, c := range testCases {
		p := Parse(c.pattern)
		assert.Equal(t, c.expectMatch, p.Match(c.target), c.pattern+" vs "+c.target)
	}
}
