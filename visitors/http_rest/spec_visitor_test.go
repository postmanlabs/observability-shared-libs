package http_rest

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/test"

	. "github.com/akitasoftware/akita-libs/visitors"
)

/* You can extend DefaultHttpRestSpecVisitor with a custom reader that
 * implements a subset of the visitor methods.  For example, MyPreorderVisitor
 * only visits Primitives in the spec and ignores other terms.
 */
type MyPreorderVisitor struct {
	DefaultSpecVisitorImpl
	actualPaths []string
}

var _ DefaultSpecVisitor = (*MyPreorderVisitor)(nil)

func (v *MyPreorderVisitor) EnterPrimitive(self interface{}, c SpecVisitorContext, p *pb.Primitive) Cont {
	// Prints the path through the REST request/response to this primitive,
	// including the host/operation/path, response code (if present), parameter
	// name, etc.
	if c.IsResponse() && c.GetRestPath()[2] == "/api/0/projects/" {
		pathWithType := append(c.GetRestPath(), GetPrimitiveType(p).String())
		v.actualPaths = append(v.actualPaths, strings.Join(pathWithType, "."))
	}
	return Continue
}

type MyPostorderVisitor struct {
	DefaultSpecVisitorImpl
	actualPaths []string
}

func (v *MyPostorderVisitor) LeavePrimitive(self interface{}, c SpecVisitorContext, p *pb.Primitive, cont Cont) Cont {
	// Prints the path through the REST request/response to this primitive,
	// including the host/operation/path, response code (if present), parameter
	// name, etc.
	if c.IsResponse() && c.GetRestPath()[2] == "/api/0/projects/" {
		pathWithType := append(c.GetRestPath(), GetPrimitiveType(p).String())
		v.actualPaths = append(v.actualPaths, strings.Join(pathWithType, "."))
	}
	return cont
}

var expectedPaths = []string{
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.slug.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.firstEvent.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.name.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.isInternal.Data.api_spec.Bool",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.organization.Data.avatar.Data.avatarType.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.organization.Data.dateCreated.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.organization.Data.features.Data.0.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.organization.Data.status.Data.id.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.organization.Data.status.Data.name.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.organization.Data.id.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.organization.Data.isEarlyAdopter.Data.api_spec.Bool",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.organization.Data.name.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.organization.Data.require2FA.Data.api_spec.Bool",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.organization.Data.slug.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.dateCreated.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.hasAccess.Data.api_spec.Bool",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.status.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.id.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.isBookmarked.Data.api_spec.Bool",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.features.Data.0.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.isMember.Data.api_spec.Bool",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.isPublic.Data.api_spec.Bool",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.avatar.Data.avatarType.Data.api_spec.String",
	"localhost:9000.GET./api/0/projects/.Response.200.Body.JSON.0.color.Data.api_spec.String",
}

func TestPreorderTraversal(t *testing.T) {
	spec := test.LoadAPISpecFromFileOrDie("../testdata/sentry_ir_spec.pb.txt")

	var visitor MyPreorderVisitor
	Apply(&visitor, spec)
	sort.Strings(expectedPaths)
	sort.Strings(visitor.actualPaths)
	assert.Equal(t, expectedPaths, visitor.actualPaths)
}

func TestPostorderTraversal(t *testing.T) {
	spec := test.LoadAPISpecFromFileOrDie("../testdata/sentry_ir_spec.pb.txt")

	var visitor MyPostorderVisitor
	Apply(&visitor, spec)
	sort.Strings(expectedPaths)
	sort.Strings(visitor.actualPaths)
	assert.Equal(t, expectedPaths, visitor.actualPaths)
}

type queryOnlyVisitor struct {
	DefaultSpecVisitorImpl
	actualPaths []string
}

var _ DefaultSpecVisitor = (*queryOnlyVisitor)(nil)

func (v *queryOnlyVisitor) EnterPrimitive(self interface{}, c SpecVisitorContext, p *pb.Primitive) Cont {
	if c.IsArg() && c.GetRestPath()[2] == "/api/1/store/" && c.GetValueType() == QUERY {
		pathWithType := append(c.GetRestPath(), GetPrimitiveType(p).String())
		v.actualPaths = append(v.actualPaths, strings.Join(pathWithType, "."))
	}
	return Continue
}

func TestFilterByValueType(t *testing.T) {
	spec := test.LoadAPISpecFromFileOrDie("../testdata/sentry_ir_spec.pb.txt")

	expectedPaths = []string{
		"localhost:9000.POST./api/1/store/.Arg.Query.sentry_key.api_spec.String",
		"localhost:9000.POST./api/1/store/.Arg.Query.sentry_version.api_spec.Int32",
	}

	var visitor queryOnlyVisitor
	Apply(&visitor, spec)
	sort.Strings(expectedPaths)
	sort.Strings(visitor.actualPaths)
	assert.Equal(t, expectedPaths, visitor.actualPaths)
}

type responsePathVisitor struct {
	DefaultSpecVisitorImpl
	PathOfInterest   string
	actualPaths      []string
	actualFieldPaths []string
}

var _ DefaultSpecVisitor = (*responsePathVisitor)(nil)

func (v *responsePathVisitor) EnterPrimitive(self interface{}, c SpecVisitorContext, p *pb.Primitive) Cont {
	// The path is specifically picked to contain response values with nested Data
	// objects.
	if c.IsResponse() && c.GetRestPath()[2] == v.PathOfInterest {
		pathWithType := append(c.GetResponsePath(), GetPrimitiveType(p).String())
		v.actualPaths = append(v.actualPaths, strings.Join(pathWithType, "."))

		fieldPath := c.GetFieldPath()
		fieldPathAsString := make([]string, len(fieldPath))
		for i, p := range fieldPath {
			fieldPathAsString[i] = p.String()
		}
		v.actualFieldPaths = append(v.actualFieldPaths, strings.Join(fieldPathAsString, "."))
	}
	return Continue
}

func TestGetDataPath(t *testing.T) {
	// Maps test files to expected paths.
	tests := []struct {
		TestFile      string
		Endpoint      string
		ExpectedPaths []string
	}{
		{"../testdata/sentry_ir_spec.pb.txt",
			"/api/0/projects/{organization_slug}/{project_slug}/users/",
			[]string{
				"Response.200.Body.JSON.0.avatarUrl.Data.api_spec.String",
				"Response.200.Body.JSON.0.dateCreated.Data.api_spec.String",
				"Response.200.Body.JSON.0.email.Data.api_spec.String",
				"Response.200.Body.JSON.0.hash.Data.api_spec.String",
				"Response.200.Body.JSON.0.id.Data.api_spec.String",
				"Response.200.Body.JSON.0.identifier.Data.api_spec.String",
				"Response.200.Body.JSON.0.ipAddress.Data.api_spec.String",
				"Response.200.Body.JSON.0.name.Data.api_spec.String",
				"Response.200.Body.JSON.0.tagValue.Data.api_spec.String",
				"Response.200.Body.JSON.0.username.Data.api_spec.String",
			},
		},
		{"../testdata/sentry_ir_map_spec.pb.txt",
			"/api/0/projects/{organization_slug}/{project_slug}/users/",
			[]string{
				"Response.200.Body.JSON.0.Key.api_spec.String",
				"Response.200.Body.JSON.0.Value.Data.api_spec.String",
				"Response.200.Body.JSON.1.avatarUrl.Data.api_spec.String",
				"Response.200.Body.JSON.1.dateCreated.Data.api_spec.String",
				"Response.200.Body.JSON.1.email.Data.api_spec.String",
				"Response.200.Body.JSON.1.hash.Data.api_spec.String",
				"Response.200.Body.JSON.1.id.Data.api_spec.String",
				"Response.200.Body.JSON.1.identifier.Data.api_spec.String",
				"Response.200.Body.JSON.1.ipAddress.Data.api_spec.String",
				"Response.200.Body.JSON.1.name.Data.api_spec.String",
				"Response.200.Body.JSON.1.tagValue.Data.api_spec.String",
				"Response.200.Body.JSON.1.username.Data.api_spec.String",
			},
		},
		{"../testdata/contains_oneof.pb.txt",
			"/api/example",
			[]string{
				// The RestPath does not pretty print the options; it includes the hash values
				"Response.200.Body.JSON.result.0ChzXURDSRY=.api_spec.String",
				"Response.200.Body.JSON.result.va5tP-fnZF8=.api_spec.Int64",
				"Response.200.Header.X-Request-Id.api_spec.String",
			},
		},
		{"../testdata/multipart_body.pb.txt",
			"/api/pets",
			[]string{
				"Response.200.Body.Multi-Part.Body.field1.TEXT_PLAIN.api_spec.String",
				"Response.200.Body.Multi-Part.Body.field2.JSON.baz.api_spec.Int64",
				"Response.200.Body.Multi-Part.Body.field2.JSON.foo.api_spec.String",
			},
		},
	}

	for _, tc := range tests {
		spec := test.LoadAPISpecFromFileOrDie(tc.TestFile)

		visitor := responsePathVisitor{
			// Pick out just one matching method in the file
			PathOfInterest: tc.Endpoint,
		}
		Apply(&visitor, spec)
		sort.Strings(tc.ExpectedPaths)
		sort.Strings(visitor.actualPaths)
		assert.Equal(t, tc.ExpectedPaths, visitor.actualPaths)
	}
}

func TestGetFieldPath(t *testing.T) {
	// Maps test files to expected paths.
	tests := []struct {
		TestFile      string
		Endpoint      string
		ExpectedPaths []string
	}{
		{"../testdata/contains_oneof.pb.txt",
			"/api/example",
			[]string{
				"result.(format 1 of 2)",
				"result.(format 2 of 2)",
				"X-Request-Id",
			},
		},
		{"../testdata/multipart_body.pb.txt",
			"/api/pets",
			[]string{
				"field1",
				"field2.foo",
				"field2.baz",
			},
		},
	}

	for _, tc := range tests {
		spec := test.LoadAPISpecFromFileOrDie(tc.TestFile)

		visitor := responsePathVisitor{
			// Pick out just one matching method in the file
			PathOfInterest: tc.Endpoint,
		}
		Apply(&visitor, spec)
		sort.Strings(tc.ExpectedPaths)
		sort.Strings(visitor.actualFieldPaths)
		assert.Equal(t, tc.ExpectedPaths, visitor.actualFieldPaths)
	}
}

type primitiveCounter struct {
	DefaultSpecVisitorImpl
	numPrimitives int
}

var _ DefaultSpecVisitor = (*primitiveCounter)(nil)

func (v *primitiveCounter) EnterPrimitive(self interface{}, c SpecVisitorContext, p *pb.Primitive) Cont {
	v.numPrimitives++
	return Continue
}

// Test for when the visitor starts at a Data node that is not a top-level
// method argument.
func TestVisitDataOnly(t *testing.T) {
	spec := test.LoadAPISpecFromFileOrDie("../testdata/sentry_ir_spec.pb.txt")

	var visitor primitiveCounter
	response := spec.Methods[0].Responses["200-body-0"]
	data := response.GetList().GetElems()[0]
	Apply(&visitor, data)
	assert.Equal(t, visitor.numPrimitives, 17)
}
