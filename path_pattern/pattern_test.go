package path_pattern

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	pattern     string
	target      string
	expectMatch bool
}

var commonTestCases = []testCase{
	{
		pattern:     "/v1/{my_arg_name}",
		target:      "/v1/foobar",
		expectMatch: true,
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
		expectMatch: true,
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
	{
		pattern:     "/^",
		target:      "/v1",
		expectMatch: true,
	},
	{
		pattern:     "/v1/^",
		target:      "/v1/foo",
		expectMatch: true,
	},
	{
		pattern:     "/v1/^/bar",
		target:      "/v1/foo/bar",
		expectMatch: true,
	},
	{
		pattern:     "/v1/^",
		target:      "/v1/{arg}",
		expectMatch: false,
	},
	{
		pattern:     "/v1/^/bar",
		target:      "/v1/{arg}/bar",
		expectMatch: false,
	},
}

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
				components: []Component{
					Val(""),
					Val("v1"),
					Val(""),
					Val("foobar"),
					Var("foobar"),
				},
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
				components: []Component{
					Val(""),
					Val("v1"),
					Var("my_arg_name"),
				},
			},
		},
		{
			input: "/v1/*/foobar",
			expected: Pattern{
				components: []Component{
					Val(""),
					Val("v1"),
					Wildcard{},
					Val("foobar"),
				},
			},
		},
	}

	for _, c := range testCases {
		p := Parse(c.input)
		assert.Equal(t, c.expected, p, c.input)
		assert.Equal(t, c.input, p.String())

		jsonStr := `"` + c.input + `"`
		var unmarshalled Pattern
		if err := json.Unmarshal([]byte(jsonStr), &unmarshalled); err != nil {
			assert.NoError(t, err, "failed to unmarshal %s", c.input)
			continue
		}
		marshalled, err := json.Marshal(unmarshalled)
		if err != nil {
			assert.NoError(t, err, "failed to marshal %s", c.input)
			continue
		}
		assert.Equal(t, jsonStr, string(marshalled), "failed to marshal/unmarshal %s", c.input)
	}
}

func TestMatch(t *testing.T) {
	testCases := []testCase{
		{
			pattern:     "/v1/{my_arg_name}",
			target:      "/v1/foobar/x",
			expectMatch: false,
		},
	}
	testCases = append(testCases, commonTestCases...)

	for _, c := range testCases {
		p := Parse(c.pattern)
		assert.Equal(t, c.expectMatch, p.Match(c.target), c.pattern+" vs "+c.target)
	}
}

func TestPrefixMatch(t *testing.T) {
	testCases := []testCase{
		{
			pattern:     "/v1/foobar",
			target:      "/v1/foobar/x",
			expectMatch: true,
		},
		{
			pattern:     "/v1/{my_arg_name}",
			target:      "/v1/foobar/x",
			expectMatch: true,
		},
		{
			pattern:     "/^",
			target:      "/v1/foobar/x",
			expectMatch: true,
		},
		{
			pattern:     "/*",
			target:      "/v1/foobar/x",
			expectMatch: true,
		},
		{
			pattern:     "/foo/",
			target:      "/foo",
			expectMatch: true,
		},
		{
			pattern:     "/foo",
			target:      "/foo/",
			expectMatch: true,
		},
		{
			pattern:     "/foo",
			target:      "/foobar",
			expectMatch: false,
		},
	}
	testCases = append(testCases, commonTestCases...)

	for _, c := range testCases {
		p := Parse(c.pattern)
		assert.Equal(t, c.expectMatch, p.PrefixMatch(c.target), c.pattern+" vs "+c.target)
	}
}
