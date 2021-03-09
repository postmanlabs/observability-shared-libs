package http_rest

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/test"

	"github.com/akitasoftware/akita-libs/visitors/go_ast"
)

/* You can extend DefaultHttpRestSpecVisitor with a custom reader that
 * implements a subset of the visitor methods.  For example, MyVisitor
 * only visits Primitives in the spec and ignores other terms.
 */
type MyVisitor struct {
	DefaultHttpRestSpecVisitor
	actualPaths []string
}

func (v *MyVisitor) VisitPrimitive(c HttpRestSpecVisitorContext, p *pb.Primitive) bool {
	// Prints the path through the REST request/response to this primitive,
	// including the host/operation/path, response code (if present), parameter
	// name, etc.
	if c.IsResponse() && c.GetRestPath()[2] == "/api/0/projects/" {
		pathWithType := append(c.GetRestPath(), GetPrimitiveType(p).String())
		v.actualPaths = append(v.actualPaths, strings.Join(pathWithType, "."))
	}
	return true
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

	var visitor MyVisitor
	Apply(go_ast.PREORDER, &visitor, spec)
	sort.Strings(expectedPaths)
	sort.Strings(visitor.actualPaths)
	assert.Equal(t, expectedPaths, visitor.actualPaths)
}

func TestPostorderTraversal(t *testing.T) {
	spec := test.LoadAPISpecFromFileOrDie("../testdata/sentry_ir_spec.pb.txt")

	var visitor MyVisitor
	Apply(go_ast.PREORDER, &visitor, spec)
	sort.Strings(expectedPaths)
	sort.Strings(visitor.actualPaths)
	assert.Equal(t, expectedPaths, visitor.actualPaths)
}

type queryOnlyVisitor struct {
	DefaultHttpRestSpecVisitor
	actualPaths []string
}

func (v *queryOnlyVisitor) VisitPrimitive(c HttpRestSpecVisitorContext, p *pb.Primitive) bool {
	if c.IsArg() && c.GetRestPath()[2] == "/api/1/store/" && c.GetValueType() == QUERY {
		pathWithType := append(c.GetRestPath(), GetPrimitiveType(p).String())
		v.actualPaths = append(v.actualPaths, strings.Join(pathWithType, "."))
	}
	return true
}

func TestFilterByValueType(t *testing.T) {
	spec := test.LoadAPISpecFromFileOrDie("../testdata/sentry_ir_spec.pb.txt")

	expectedPaths = []string{
		"localhost:9000.POST./api/1/store/.Arg.Query.sentry_key.api_spec.String",
		"localhost:9000.POST./api/1/store/.Arg.Query.sentry_version.api_spec.Int32",
	}

	var visitor queryOnlyVisitor
	Apply(go_ast.PREORDER, &visitor, spec)
	sort.Strings(expectedPaths)
	sort.Strings(visitor.actualPaths)
	assert.Equal(t, expectedPaths, visitor.actualPaths)
}

type responsePathVisitor struct {
	DefaultHttpRestSpecVisitor
	actualPaths []string
}

func (v *responsePathVisitor) VisitPrimitive(c HttpRestSpecVisitorContext, p *pb.Primitive) bool {
	// The path is specifically picked to contain response values with nested Data
	// objects.
	if c.IsResponse() && c.GetRestPath()[2] == "/api/0/projects/{organization_slug}/{project_slug}/users/" {
		pathWithType := append(c.GetResponsePath(), GetPrimitiveType(p).String())
		v.actualPaths = append(v.actualPaths, strings.Join(pathWithType, "."))
	}
	return true
}

func TestGetDataPath(t *testing.T) {
	spec := test.LoadAPISpecFromFileOrDie("../testdata/sentry_ir_spec.pb.txt")

	expectedPaths = []string{
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
	}

	var visitor responsePathVisitor
	Apply(go_ast.PREORDER, &visitor, spec)
	sort.Strings(expectedPaths)
	sort.Strings(visitor.actualPaths)
	assert.Equal(t, expectedPaths, visitor.actualPaths)
}
