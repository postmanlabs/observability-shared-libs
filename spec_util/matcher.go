package spec_util

import (
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

type methodRegexp struct {
	Operation         string         // HTTP operation
	Host              string         // HTTP host
	Template          string         // original method template
	RE                *regexp.Regexp // template converted to regexp on path
	VariablePositions []int          // positions of path variables in templates, in sorted order
}

func (r methodRegexp) LessThan(other methodRegexp) bool {
	for i, p := range r.VariablePositions {
		// Other template has more specific path if it has fewer
		// variables, or the variable in it comes later.
		if i >= len(other.VariablePositions) {
			return false
		}
		if p < other.VariablePositions[i] {
			return false
		}
		if p > other.VariablePositions[i] {
			return true
		}
	}
	// Fall back to string comparison
	if r.Template < other.Template {
		return true
	}
	return false
}

// MethodMatcher is currently a list of regular expressions to try in order;
// in the future it could be a tree lookup structure (for efficiency and
// to more easily accommodate longest-prefix matching.)
//
// During creation, ensure that /abc/def is sorted before /abc/{var1} so that the former is preferred.
type MethodMatcher struct {
	methods []methodRegexp
}

// Original behavior-- deprecated but still in use.
// Lookup returns either a matching template, or the original path if no match is found.
func (m *MethodMatcher) Lookup(operation string, path string) (template string) {
	for _, candidate := range m.methods {
		if candidate.Operation != operation {
			continue
		}
		if candidate.RE.MatchString(path) {
			return candidate.Template
		}
	}
	return path
}

// Lookup returns either a matching template, or the original path if no match is found.
// This version matches on host as well.
// If there is no exact match on (operation, host,string) accept a partial match on (host,string) instead.
// This handles things calls like OPTION that we do not include in our API model, which currently does
// path parameter inference without considering operations to be distinct.
func (m *MethodMatcher) LookupWithHost(operation string, host string, path string) (template string) {
	for _, candidate := range m.methods {
		if candidate.Operation != operation {
			continue
		}
		if candidate.Host != host {
			continue
		}
		if candidate.RE.MatchString(path) {
			return candidate.Template
		}
	}
	// If we failed, try again without Operation filter
	for _, candidate := range m.methods {
		if candidate.Host != host {
			continue
		}
		if candidate.RE.MatchString(path) {
			return candidate.Template
		}
	}
	return path
}

const (
	// Allow % for URL encoding but I'm not bothering to verify that the
	// correct format is followed.  Other valid unreserved characters are
	// - . _ ~
	// according to RFC3986.  I'm not accepting the reserved characters
	// : / ? # [ ] @ ! $ & ' ( ) *  + , ; =
	//
	uriPathCharacters = "[A-Za-z0-9-._~%]+"
	uriArgument       = "\\{.*?\\}" // non-greedy match
)

var (
	uriArgumentRegexp = regexp.MustCompile(uriArgument)
)

// Convert a string with templates like
// v1/api/get/user/{arg1}/{arg2}
// to a regular expression that matches the entire path like
// ^v1/api/get/user/([^/]+)/([^/]+)$
//
// Return the position of each argument within the original template, in sorted order,
// counting all variables as length 1.
func templateToRegexp(pathTemplate string) (*regexp.Regexp, []int, error) {
	// If there are special characters, then the easiest way to escape them is
	// to break the string up by arguments, and escape everything in between.
	literals := uriArgumentRegexp.Split(pathTemplate, -1)

	// Insert between every pair of literals, so not after the last.
	// If the path ends with an argument we should get an empty literal at
	// the end.
	var buf strings.Builder
	buf.WriteString("^")
	positions := make([]int, 0, len(literals)-1)
	first := true
	currentPosition := 0

	for _, l := range literals {
		if first {
			// No variable before the first literal
			first = false
		} else {
			buf.WriteString(uriPathCharacters)
			positions = append(positions, currentPosition)
			currentPosition += 1
		}
		buf.WriteString(regexp.QuoteMeta(l))
		currentPosition += len(l)
	}
	buf.WriteString("$")
	re, err := regexp.Compile(buf.String())
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not convert template %q to regexp", pathTemplate)
	}
	return re, positions, nil

}

// NewMethodMatcher takes an API spec and returns a dictionary that
// converts witness methods into the matching templatized path in the spec.
func NewMethodMatcher(spec *pb.APISpec) (*MethodMatcher, error) {
	// Convert each method in the spec to a regular expression
	mm := &MethodMatcher{
		methods: make([]methodRegexp, 0, len(spec.Methods)),
	}

	for _, specMethod := range spec.Methods {
		httpMeta := HTTPMetaFromMethod(specMethod)
		if httpMeta == nil {
			continue // just ignore non-http methods
		}
		re, positions, err := templateToRegexp(httpMeta.PathTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "could not extract paths from spec")
		}
		mm.methods = append(mm.methods, methodRegexp{
			Operation:         httpMeta.Method,
			Host:              httpMeta.Host,
			Template:          httpMeta.PathTemplate,
			RE:                re,
			VariablePositions: positions,
		})
	}

	// Order by most-specific path first
	sort.Slice(mm.methods, func(i, j int) bool {
		return mm.methods[i].LessThan(mm.methods[j])
	})

	return mm, nil
}
