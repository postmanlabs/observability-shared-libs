package path_pattern

import (
	"regexp"
)

// Represents a path component value, which can be either a concrete string Val
// or a Var.
type Component interface {
	Match(string) bool
	String() string
	Regexp() string
}

type Val string

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

func (DoubleWildcard) Match(c string) bool {
	return true
}

func (DoubleWildcard) String() string {
	return "**"
}

func (v DoubleWildcard) Regexp() string {
	return "(.*)"
}
