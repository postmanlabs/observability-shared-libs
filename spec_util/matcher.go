package spec_util

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

type methodRegexp struct {
	Operation string         // HTTP operation
	Template  string         // original method template
	RE        *regexp.Regexp // template converted to regexp on path
}

// MethodMatcher is currently a list of regular expressions to try in order;
// in the future it could be a tree lookup structure (for efficiency and
// to more easily accommodate longest-prefix matching.)
type MethodMatcher struct {
	methods []methodRegexp
}

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

const (
	// Allow % for URL encoding but I'm not bothering to verify that the
	// correct format is followed.  Other valid unreserved characters are
	// - . _ ~
	// according to RFC3986.  I'm not accepting the reserved characters
	// : / ? # [ ] @ ! $ & ' ( ) *  + , ; =
	//
	uriPathCharacters = "[A-Za-z0-9.,_~%]+"
	uriArgument       = "\\{.*?\\}" // non-greedy match
)

var (
	uriArgumentRegexp = regexp.MustCompile(uriArgument)
)

// Convert a string with templates like
// v1/api/get/user/{arg1}/{arg2}
// to a regular expression that matches the entire path like
// ^v1/api/get/user/([^/]+)/([^/]+)$
func templateToRegexp(pathTemplate string) (*regexp.Regexp, error) {
	// If there are special characters, then the easiest way to escape them is
	// to break the string up by arguments, and escape everything in between.
	literals := uriArgumentRegexp.Split(pathTemplate, -1)

	// Insert between every pair of literals, so not after the last.
	// If the path ends with an argument we should get an empty literal at
	// the end.
	var buf strings.Builder
	first := true
	for _, l := range literals {
		if first {
			first = false
		} else {
			buf.WriteString(uriPathCharacters)
		}
		buf.WriteString(regexp.QuoteMeta(l))
	}
	re, err := regexp.Compile(buf.String())
	if err != nil {
		return nil, errors.Wrapf(err, "could not convert template %q to regexp", pathTemplate)
	}
	return re, nil

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
		re, err := templateToRegexp(httpMeta.PathTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "could not extract paths from spec")
		}
		mm.methods = append(mm.methods, methodRegexp{
			Operation: httpMeta.Method,
			Template:  httpMeta.PathTemplate,
			RE:        re,
		})
	}
	return mm, nil
}
