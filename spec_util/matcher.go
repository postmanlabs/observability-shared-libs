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
	ConcretePathRE    *regexp.Regexp // template converted to regexp for matching concrete paths
	TemplatePathRE    *regexp.Regexp // template converted to regexp for matching path templates
	VariablePositions []int          // positions of path variables in templates, in sorted order
}

func (r methodRegexp) LessThan(other methodRegexp) bool {
	for i, p := range r.VariablePositions {
		// Other template has more specific path if it has fewer variables, or the
		// variable in it comes later.
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

// MethodMatcher is currently a list of regular expressions to try in order; in
// the future it could be a tree lookup structure (for efficiency and to more
// easily accommodate longest-prefix matching.)
//
// During creation, ensure that /abc/def is sorted before /abc/{var1} so that
// the former is preferred.
type MethodMatcher struct {
	methods []methodRegexp
}

type MethodMatchOptions struct {
	// If true, the operation (GET, PUT, etc.) must match.
	MatchOperation bool

	// If true, the host must match.
	MatchHost bool

	// If true, the given path must be a concrete path that matches against a
	// path template in the matcher.
	MatchConcretePaths bool

	// If true, the given path can be a path template. To match, the given
	// template must be equivalent to, or a refinement of, a template in the
	// matcher (modulo alpha renaming).
	MatchPathTemplates bool
}

// Returns either a matching template, or the original path if no match is
// found.
func (m *MethodMatcher) Lookup(operation string, host string, path string, opts MethodMatchOptions) (template string, found bool) {
	for _, candidate := range m.methods {
		if opts.MatchOperation && candidate.Operation != operation {
			continue
		}
		if opts.MatchHost && candidate.Host != host {
			continue
		}
		if opts.MatchConcretePaths && !candidate.ConcretePathRE.MatchString(path) {
			continue
		}
		if opts.MatchPathTemplates && !candidate.TemplatePathRE.MatchString(path) {
			continue
		}
		return candidate.Template, true
	}

	return path, false
}

// Returns either a matching template, or the original path if no match is
// found. If there is no exact match on (operation, host, string) a partial
// match on (host, string) is attempted instead. This handles things calls like
// OPTION that we do not include in our API model, which currently does path
// parameter inference without considering operations to be distinct.
func (m *MethodMatcher) LookupWithHost(operation string, host string, path string) (template string, found bool) {
	opts := MethodMatchOptions{
		MatchOperation:     true,
		MatchHost:          true,
		MatchConcretePaths: true,
		MatchPathTemplates: false,
	}
	template, found = m.Lookup(operation, host, path, opts)

	// If we failed, try again without Operation filter
	if !found {
		opts.MatchOperation = false
		template, found = m.Lookup(operation, host, path, opts)
	}

	return template, found
}

const (
	// Allow % for URL encoding but I'm not bothering to verify that the correct
	// format is followed.  Other valid unreserved characters are
	//   - . _ ~
	// according to RFC3986.  I'm not accepting the reserved characters
	//   : / ? # [ ] @ ! $ & ' ( ) *  + , ; =
	//
	uriPathCharacters            = "[A-Za-z0-9-._~%]+"
	uriPathCharactersOrParameter = "(" + uriPathCharacters + "|{" + uriPathCharacters + "})"
	uriArgument                  = "\\{.*?\\}" // non-greedy match
)

var (
	uriArgumentRegexp = regexp.MustCompile(uriArgument)
)

// Convert a string with templates like
//   v1/api/get/user/{arg1}/{arg2}
// to a pair of regular expressions. The first matches the entire path like
//   ^v1/api/get/user/([^/{}]+)/([^/{}]+)$
// the second matches a refinement of the template like
//   ^v1/api/get/user/([^/{}]+|{[^/{}]+})/([^/{}]+|{[^/{}]+})$
//
// Return the position of each argument within the original template, in sorted
// order, counting all variables as length 1.
func templateToRegexp(pathTemplate string) (concreteRE *regexp.Regexp, templateRE *regexp.Regexp, argPositions []int, err error) {
	// If there are special characters, then the easiest way to escape them is to
	// break the string up by arguments, and escape everything in between.
	literals := uriArgumentRegexp.Split(pathTemplate, -1)

	// Insert between every pair of literals, so not after the last. If the path
	// ends with an argument we should get an empty literal at the end.
	var concreteBuf strings.Builder
	var templateBuf strings.Builder
	concreteBuf.WriteString("^")
	templateBuf.WriteString("^")
	argPositions = make([]int, 0, len(literals)-1)
	first := true
	currentPosition := 0

	for _, l := range literals {
		if first {
			// No variable before the first literal
			first = false
		} else {
			concreteBuf.WriteString(uriPathCharacters)
			templateBuf.WriteString(uriPathCharactersOrParameter)
			argPositions = append(argPositions, currentPosition)
			currentPosition += 1
		}
		concreteBuf.WriteString(regexp.QuoteMeta(l))
		templateBuf.WriteString(regexp.QuoteMeta(l))
		currentPosition += len(l)
	}

	concreteBuf.WriteString("$")
	concreteRE, err = regexp.Compile(concreteBuf.String())
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "could not convert template %q to concrete regexp", pathTemplate)
	}

	templateBuf.WriteString("$")
	templateRE, err = regexp.Compile(templateBuf.String())
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "could not convert template %q to template regexp", pathTemplate)
	}

	return concreteRE, templateRE, argPositions, nil
}

// NewMethodMatcher takes an API spec and returns a dictionary that converts
// witness methods into the matching templatized path in the spec.
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
		concreteRE, templateRE, positions, err := templateToRegexp(httpMeta.PathTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "could not extract paths from spec")
		}
		mm.methods = append(mm.methods, methodRegexp{
			Operation:         httpMeta.Method,
			Host:              httpMeta.Host,
			Template:          httpMeta.PathTemplate,
			ConcretePathRE:    concreteRE,
			TemplatePathRE:    templateRE,
			VariablePositions: positions,
		})
	}

	// Order by most-specific path first
	sort.Slice(mm.methods, func(i, j int) bool {
		return mm.methods[i].LessThan(mm.methods[j])
	})

	return mm, nil
}
