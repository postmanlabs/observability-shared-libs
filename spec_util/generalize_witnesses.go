package spec_util

import (
	"fmt"
	"regexp"

	"github.com/golang/protobuf/proto"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/spec_util/ir_hash"
)

func GetPathRegexps(spec *pb.APISpec) map[*regexp.Regexp]string {
	return getPathRegexps(getPaths(spec))
}

// Transforms witnesses that are concrete instances of paths with path parameters
// into witnesses with path parameters.
//
// Witnesses are copied, not mutated.
//
// For example, suppose a spec has a method on path `/v1/services/{arg1}`, and a
// witness has a path `/v1/services/svc_123`.  GeneralizeWitnesses will produce
// a new witness with path `/v1/services/{arg1}` and a new request argument mapping
// `arg1 -> svc_123`.
func GeneralizeWitness(pathMatchers map[*regexp.Regexp]string, witnessIn *pb.Witness) (*pb.Witness, error) {
	witness, ok := proto.Clone(witnessIn).(*pb.Witness)
	if !ok {
		return nil, fmt.Errorf("failed to clone witness")
	}

	// Get the path, if it exists
	if witness.GetMethod().GetMeta().GetHttp() == nil {
		return witness, nil
	}
	path := witness.GetMethod().GetMeta().GetHttp().GetPathTemplate()

	// Remove trailing '/', if present
	if len(path) == 0 {
		return witness, nil
	}
	if path[len(path)-1] == '/' {
		path = path[0 : len(path)-1]
	}

	// Try each generalized path regular expression.  If one matches, generalize
	// and return the witness.
	for pathRegexp, generalizedPath := range pathMatchers {
		match := pathRegexp.FindStringSubmatch(path)
		if len(match) == 0 {
			continue
		}

		// Matches, but no path parameters in spec path, so no changes needed
		// to the witness
		if len(pathRegexp.SubexpNames()) == 0 {
			return witness, nil
		}

		// Add each concrete path argument from the concrete path as arg data
		// in the witness
		for i, arg := range match {
			// Skip 0th element, which is the whole matched string
			if i == 0 {
				continue
			}

			parameterName := pathRegexp.SubexpNames()[i]
			argPrim := &pb.Data_Primitive{Primitive: CategorizeString(arg).Obfuscate().ToProto()}
			argMeta := &pb.DataMeta{Meta: &pb.DataMeta_Http{Http: &pb.HTTPMeta{
				Location: &pb.HTTPMeta_Path{Path: &pb.HTTPPath{
					Key: parameterName,
				}},
			}}}
			argData := &pb.Data{
				Value:    argPrim,
				Meta:     argMeta,
				Nullable: false,
			}
			hash := ir_hash.HashDataToString(argData)
			witness.GetMethod().Args[hash] = argData
		}

		// Update the witness with the generalized path
		witness.GetMethod().GetMeta().GetHttp().PathTemplate = generalizedPath

		return witness, nil
	}

	// No path matched--this shouldn't happen.  Return the witness clone unmodified.
	return witness, nil
}

func getPaths(spec *pb.APISpec) []string {
	paths := make(map[string]struct{}, 0)
	for _, method := range spec.Methods {
		if method.GetMeta() == nil || method.GetMeta().GetHttp() == nil {
			continue
		}
		paths[method.GetMeta().GetHttp().GetPathTemplate()] = struct{}{}
	}

	rv := make([]string, 0, len(paths))
	for path := range paths {
		rv = append(rv, path)
	}

	return rv
}

// For each path, produce a regular expression that captures the arguments
// at each path parameter position.
func getPathRegexps(paths []string) map[*regexp.Regexp]string {
	rv := make(map[*regexp.Regexp]string, len(paths))
	parameterMatcher := regexp.MustCompile("{(.*?)}")
	for _, path := range paths {
		// Remove trailing slash, if any
		if len(path) > 0 && path[len(path)-1] == '/' {
			path = path[0 : len(path)-1]
		}
		pathRegexp := regexp.MustCompile("^" + parameterMatcher.ReplaceAllString(path, "(?P<$1>[^{/}]*)") + "$")
		rv[pathRegexp] = path
	}
	return rv
}
