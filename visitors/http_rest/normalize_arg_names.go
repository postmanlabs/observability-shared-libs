package http_rest

import (
	"fmt"
	"strconv"
	"strings"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	. "github.com/akitasoftware/akita-libs/visitors"
	"github.com/akitasoftware/akita-libs/visitors/go_ast"
	"github.com/pkg/errors"
)

// Represents the "name" of an argument to a method.
type argName interface {
	String() string

	isArgName()
}

// Returns a version of the given method-argument map, wherein arguments are
// indexed by name, rather than by hash.
//
// In IR method objects, arguments to requests and responses are indexed by
// the arguments' hashes. However, because arguments are not normalized,
// equivalent arguments can have different hashes, so these indices are not
// useful for determining whether and how a method has changed.
func getNormalizedArgMap(args map[string]*pb.Data) map[argName]*pb.Data {
	// XXX Ignoring errors from the normalizer regarding non-HTTP metadata.
	normalizer := newArgMapNormalizer()
	Apply(normalizer, args)
	return normalizer.normalizedMap
}

type argMapNormalizer struct {
	DefaultSpecVisitorImpl

	// Metadata for the method whose arg map is being normalized.
	methodMeta *pb.HTTPMethodMeta

	// The arg item whose name is being normalized.
	arg *pb.Data

	// The output of the normalization.
	normalizedMap map[argName]*pb.Data

	// Contains any arguments encountered with non-HTTP metadata.
	nonHTTPArgs []*pb.Data

	err error
}

var _ DefaultSpecVisitor = (*argMapNormalizer)(nil)

func newArgMapNormalizer() *argMapNormalizer {
	return &argMapNormalizer{
		normalizedMap: make(map[argName]*pb.Data),
	}
}

func (v *argMapNormalizer) EnterData(self interface{}, _ SpecVisitorContext, arg *pb.Data) Cont {
	// Set our context.
	v.arg = arg

	// Make sure we have HTTP metadata for the current argument.
	argMeta := arg.GetMeta().GetHttp()
	if argMeta == nil {
		v.nonHTTPArgs = append(v.nonHTTPArgs, arg)
		return SkipChildren
	}

	return Continue
}

func (*argMapNormalizer) VisitDataChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, arg *pb.Data) Cont {
	// Only visit the argument's metadata.
	ctx := c.AppendPath("Meta")
	return go_ast.ApplyWithContext(vm, ctx, arg.GetMeta())
}

func (v *argMapNormalizer) setName(name argName) {
	if _, ok := v.normalizedMap[name]; ok {
		panic(fmt.Sprintf("Unexpected duplicated name for %v", name))
	}
	v.normalizedMap[name] = v.arg
}

func (v *argMapNormalizer) LeaveData(self interface{}, _ SpecVisitorContext, _ *pb.Data, cont Cont) Cont {
	v.arg = nil
	return cont
}

// == Path parameters =========================================================

type pathName struct {
	index int
}

var _ argName = (*pathName)(nil)

func (pathName) isArgName() {}

func (v *argMapNormalizer) EnterHTTPPath(self interface{}, _ SpecVisitorContext, path *pb.HTTPPath) Cont {
	template := v.methodMeta.GetPathTemplate()
	components := strings.Split(template, "/")
	for idx, component := range components {
		if component == path.GetKey() {
			v.setName(pathName{
				index: idx,
			})
			return SkipChildren
		}
	}

	v.err = errors.Errorf("Path parameter %q not found in %q", path.GetKey(), template)
	return SkipChildren
}

func (n pathName) String() string {
	return strconv.Itoa(n.index)
}

// == Query parameters ========================================================

type queryName struct {
	name string
}

var _ argName = (*queryName)(nil)

func (queryName) isArgName() {}

func (v *argMapNormalizer) EnterHTTPQuery(self interface{}, _ SpecVisitorContext, query *pb.HTTPQuery) Cont {
	v.setName(queryName{
		name: query.GetKey(),
	})
	return SkipChildren
}

func (n queryName) String() string {
	return n.name
}

// == Header parameters =======================================================

type headerName struct {
	name string
}

var _ argName = (*headerName)(nil)

func (headerName) isArgName() {}

func (v *argMapNormalizer) EnterHTTPHeader(self interface{}, _ SpecVisitorContext, header *pb.HTTPHeader) Cont {
	v.setName(headerName{
		name: header.GetKey(),
	})
	return SkipChildren
}

func (n headerName) String() string {
	return n.name
}

// == Cookie parameters =======================================================

type cookieName struct {
	name string
}

var _ argName = (*cookieName)(nil)

func (cookieName) isArgName() {}

func (v *argMapNormalizer) EnterHTTPCookie(self interface{}, _ SpecVisitorContext, cookie *pb.HTTPCookie) Cont {
	v.setName(cookieName{
		name: cookie.GetKey(),
	})
	return SkipChildren
}

func (n cookieName) String() string {
	return n.name
}

// == Body parameters =========================================================

type bodyName struct{}

var _ argName = (*bodyName)(nil)

func (bodyName) isArgName() {}

func (v *argMapNormalizer) EnterHTTPBody(self interface{}, _ SpecVisitorContext, body *pb.HTTPBody) Cont {
	// Assumes there is at most one per method.
	v.setName(bodyName{})
	return SkipChildren
}

func (n bodyName) String() string {
	return "(body)"
}

// == Empty parameters ========================================================

type emptyName struct{}

var _ argName = (*bodyName)(nil)

func (emptyName) isArgName() {}

func (v *argMapNormalizer) EnterHTTPEmpty(self interface{}, _ SpecVisitorContext, empty *pb.HTTPEmpty) Cont {
	// Assumes there is at most one per method.
	v.setName(emptyName{})
	return SkipChildren
}

func (n emptyName) String() string {
	return ""
}

// == Auth parameters =========================================================

type authName struct{}

var _ argName = (*authName)(nil)

func (authName) isArgName() {}

func (v *argMapNormalizer) EnterAuth(_ SpecVisitorContext, auth *pb.HTTPAuth) Cont {
	// Assumes there is at most one per method.
	v.setName(authName{})
	return SkipChildren
}

func (n authName) String() string {
	return "Authorization"
}

// == Multipart body parameters ===============================================

type multipartName struct{}

var _ argName = (*multipartName)(nil)

func (multipartName) isArgName() {}

func (v *argMapNormalizer) EnterMultipart(_ SpecVisitorContext, multipart *pb.HTTPMultipart) Cont {
	// Assumes there is at most one per method.
	v.setName(multipartName{})
	return SkipChildren
}

func (n multipartName) String() string {
	return "(multipart)"
}
