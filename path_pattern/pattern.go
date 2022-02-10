package path_pattern

import (
	"regexp"
	"strings"
)

type Pattern interface {
	// Returns this pattern, represented as a list of path components.
	Components() []Component

	// Returns true if the pattern matches the path.  Patterns and paths are
	// compared after removing any trailing slashes.
	Match(string) bool

	// Returns true if the pattern matches the path.  Patterns and paths are
	// compared after removing any trailing slashes.
	//
	// Also returns a list of submatches, where each component is interpreted as a
	// match group.  The first element is the entire matched string, and the
	// remaining elements are per-component matches.
	//
	// For example, "/v1/{arg}/**".MatchWithGroup("/v1/foo/bar/baz") would return
	// ["/v1/foo/bar/baz", "v1", "foo", "bar/baz"].
	//
	// See documentation for regexp.FindStringSubmatch for more details.
	MatchWithGroup(string) (bool, []string)

	// Returns a string that parses into an equivalent pattern.
	String() string

	MarshalText() ([]byte, error)
	UnmarshalText(data []byte) error
}

type patternImpl struct {
	components []Component
	regexp     *regexp.Regexp
}

func (p *patternImpl) Components() []Component {
	return p.components
}

func (p *patternImpl) String() string {
	parts := make([]string, 0, len(p.components))
	for _, c := range p.components {
		parts = append(parts, c.String())
	}
	return strings.Join(parts, "/")
}

func (p *patternImpl) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *patternImpl) UnmarshalText(data []byte) error {
	*p = *Parse(string(data)).(*patternImpl)
	return nil
}

func (p *patternImpl) getOrCreateRegexp() *regexp.Regexp {
	if p.regexp != nil {
		return p.regexp
	}
	var pieces []string
	for _, piece := range p.components {
		pieces = append(pieces, piece.Regexp())
	}
	p.regexp = regexp.MustCompile("^" + strings.Join(pieces, "/") + "$")
	return p.regexp
}

func (p *patternImpl) Match(v string) bool {
	return p.getOrCreateRegexp().MatchString(removeTrailingSlashes(v))
}

func (p *patternImpl) MatchWithGroup(v string) (bool, []string) {
	subMatches := p.getOrCreateRegexp().FindStringSubmatch(removeTrailingSlashes(v))
	return subMatches != nil, subMatches
}

// Removes trailing slashes, except the final slash if v == "/".
func removeTrailingSlashes(v string) string {
	s := v
	for len(s) > 1 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

// Converts a string pattern "/v1/{arg2}" to Pattern.
func Parse(v string) Pattern {
	parts := strings.Split(removeTrailingSlashes(v), "/")
	result := &patternImpl{
		components: make([]Component, 0, len(parts)),
	}

	for _, p := range parts {
		if p == "*" {
			result.components = append(result.components, Wildcard{})
		} else if p == "**" {
			result.components = append(result.components, DoubleWildcard{})
		} else if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			result.components = append(result.components, Var(p[1:len(p)-1]))
		} else {
			result.components = append(result.components, Val(p))
		}
	}
	return result
}
