package spec_util

import (
	"regexp"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/visitors"
	"github.com/akitasoftware/akita-libs/visitors/http_rest"
)

// Heuristically determines whether the given pb.Struct (assumed to not
// represent a map) should be a map.
func StructShouldBeMap(struc *pb.Struct) bool {
	// A struct should be a map if its total number of fields exceeds
	// maxFieldsPerStruct.
	if len(struc.Fields) > maxFieldsPerStruct {
		return true
	}

	// A struct should be a map if its number of optional fields exceeds
	// maxOptionalFieldsPerStruct.
	numOptionalFields := 0
	allFieldsStartWithNumbers := true
	for fieldName, field := range struc.Fields {
		if field.GetOptional() != nil {
			numOptionalFields++
			if numOptionalFields > maxOptionalFieldsPerStruct {
				return true
			}
		}
		if !startsWithNumber(fieldName) {
			allFieldsStartWithNumbers = false
		}
	}

	// A struct should be a map if all its fields start with numbers and there
	// are a sufficient number of fields.  Because many programming languages
	// disallow struct names starting with numbers, the number of fields is
	// lower.
	if allFieldsStartWithNumbers && len(struc.Fields) >= minNumberedFields {
		return true
	}

	return false
}

// Check each non-map struct in method, and convert structs to maps
// if StructShouldBeMap is true.
func InferMapsInMethod(method *pb.Method) {
	http_rest.Apply(newStructToMapVisitor(), method)
}

// Check each non-map struct in model, and convert structs to maps
// if StructShouldBeMap is true.
func InferMapsInModel(model *pb.APISpec) {
	http_rest.Apply(newStructToMapVisitor(), model)
}

var startsWithNumberRegexp = regexp.MustCompile(`^\d`)

func startsWithNumber(s string) bool {
	return startsWithNumberRegexp.MatchString(s)
}

type structToMapVisitor struct {
	http_rest.DefaultSpecVisitorImpl
	melder melder
}

var _ http_rest.DefaultSpecVisitor = (*structToMapVisitor)(nil)

func newStructToMapVisitor() *structToMapVisitor {
	return &structToMapVisitor{melder: melder{mergeTracking: true}}
}

func (v *structToMapVisitor) EnterStruct(self interface{}, c http_rest.SpecVisitorContext, s *pb.Struct) visitors.Cont {
	if !isMap(s) && StructShouldBeMap(s) {
		v.melder.structToMap(s)
	}
	return visitors.Continue
}
