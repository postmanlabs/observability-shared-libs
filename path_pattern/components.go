package path_pattern

import (
	"regexp"
)

// Represents a path component value, which can be a concrete string Val, a Var,
// a Wildcard, or a DoubleWildcard.
type Component interface {
	Match(string) bool
	String() string

	// Returns a parenthesized regular expression for this component.
	Regexp() string
}

type Val string

var _ Component = (*Val)(nil)

func (v Val) Match(c string) bool {
	return string(v) == c
}

func (v Val) String() string {
	return string(v)
}

func (v Val) Regexp() string {
	return "(" + regexp.QuoteMeta(v.String()) + ")"
}

type Var string

var _ Component = (*Var)(nil)

func (Var) Match(c string) bool {
	// Var matches anything other than empty.
	return len(c) > 0
}

func (v Var) String() string {
	return "{" + string(v) + "}"
}

func (v Var) Regexp() string {
	return "([^/]+)"
}

// A component that matches any path argument, either a concrete value or a
// parameter.
type Wildcard struct{}

var _ Component = (*Wildcard)(nil)

func (Wildcard) Match(c string) bool {
	return true
}

func (Wildcard) String() string {
	return "*"
}

func (v Wildcard) Regexp() string {
	return "([^/]+)"
}

// A component that matches any number of path arguments.
type DoubleWildcard struct{}

var _ Component = (*DoubleWildcard)(nil)

func (DoubleWildcard) Match(c string) bool {
	return true
}

func (DoubleWildcard) String() string {
	return "**"
}

func (v DoubleWildcard) Regexp() string {
	return "(.*)"
}

// A component that should retain the original value verbatim, otherwise behaves
// like a wildcard.  Matches everything except path parameters.
type Placeholder struct{}

var _ Component = (*Placeholder)(nil)

func (Placeholder) Match(c string) bool {
	return true
}

func (Placeholder) String() string {
	return "^"
}

func (Placeholder) Regexp() string {
	return "([^{}/]+)"
}
