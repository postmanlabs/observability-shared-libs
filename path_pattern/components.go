package path_pattern

// Represents a path component value, which can be either a concrete string Val
// or a Var.
type Component interface {
	Match(string) bool
	String() string
}

type Val string

func (v Val) Match(c string) bool {
	return string(v) == c
}

func (v Val) String() string {
	return string(v)
}

type Var string

func (Var) Match(c string) bool {
	// Var matches anything other than empty.
	return len(c) > 0
}

func (v Var) String() string {
	return "{" + string(v) + "}"
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

// A component that should retain the original value verbatim, otherwise behaves
// like a wildcard.
type Placeholder struct{}

func (Placeholder) Match(c string) bool {
	return len(c) > 0
}

func (Placeholder) String() string {
	return "^"
}
