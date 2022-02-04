package path_pattern

import (
	"regexp"
	"strings"
)

type Pattern []Component

func (p Pattern) String() string {
	parts := make([]string, 0, len(p))
	for _, c := range p {
		parts = append(parts, c.String())
	}
	return strings.Join(parts, "/")
}

func (p Pattern) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *Pattern) UnmarshalText(data []byte) error {
	*p = Parse(string(data))
	return nil
}

func (p Pattern) regexp() *regexp.Regexp {
	var pieces []string
	for _, piece := range p {
		pieces = append(pieces, piece.Regexp())
	}
	return regexp.MustCompile("^" + strings.Join(pieces, "/") + "$")
}

// Match happens if the pattern exactly matches the string.
func (p Pattern) Match(v string) bool {
	r := p.regexp()
	return r.MatchString(removeTrailingSlashes(v))
}

func (p Pattern) MatchWithGroup(v string) (bool, []string) {
	r := p.regexp()
	subMatches := r.FindStringSubmatch(removeTrailingSlashes(v))
	return subMatches != nil, subMatches
}

// Removes trailing slashes, except the final slash if v == "/".
func removeTrailingSlashes(v string) string {
	s := v
	for len(s) > 1 && s[len(s) - 1] == '/' {
		s = s[:len(s) - 1]
	}
	return s
}

// Converts a string pattern "/v1/{arg2}" to Pattern.
func Parse(v string) Pattern {
	parts := strings.Split(removeTrailingSlashes(v), "/")
	result := make(Pattern, 0, len(parts))

	for _, p := range parts {
		if p == "*" {
			result = append(result, Wildcard{})
		} else if p == "**" {
			result = append(result, DoubleWildcard{})
		} else if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			result = append(result, Var(p[1:len(p)-1]))
		} else {
			result = append(result, Val(p))
		}
	}
	return result
}
