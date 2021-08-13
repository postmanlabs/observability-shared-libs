package path_pattern

import (
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

// Match happens if the prefix of the input matches the pattern.
func (p Pattern) Match(v string) bool {
	parts := strings.Split(v, "/")

	compLen := len(p)
	if compLen > len(parts) {
		return false
	}

	for i := 0; i < compLen; i++ {
		if !p[i].Match(parts[i]) {
			return false
		}
	}
	return true
}

// Converts a string pattern "/v1/{arg2}" to Pattern.
func Parse(v string) Pattern {
	parts := strings.Split(v, "/")
	result := make(Pattern, 0, len(parts))

	for _, p := range parts {
		if p == "*" {
			result = append(result, Wildcard{})
		} else if p == "^" {
			result = append(result, Placeholder{})
		} else if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			result = append(result, Var(p[1:len(p)-1]))
		} else {
			result = append(result, Val(p))
		}
	}
	return result
}
