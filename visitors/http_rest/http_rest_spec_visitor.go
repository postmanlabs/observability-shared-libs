package http_rest

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"

	. "github.com/akitasoftware/akita-libs/visitors"
	"github.com/akitasoftware/akita-libs/visitors/go_ast"
)

// VisitorManager that lets you read each message in an APISpec, starting with
// the APISpec message itself.
type SpecVisitor interface {
	EnterAPISpec(self interface{}, ctxt SpecVisitorContext, node *pb.APISpec) Cont
	VisitAPISpecChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.APISpec) Cont
	LeaveAPISpec(self interface{}, ctxt SpecVisitorContext, node *pb.APISpec, cont Cont) Cont

	EnterMethod(self interface{}, ctxt SpecVisitorContext, node *pb.Method) Cont
	VisitMethodChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.Method) Cont
	LeaveMethod(self interface{}, ctxt SpecVisitorContext, node *pb.Method, cont Cont) Cont

	EnterMethodMeta(self interface{}, ctxt SpecVisitorContext, node *pb.MethodMeta) Cont
	VisitMethodMetaChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.MethodMeta) Cont
	LeaveMethodMeta(self interface{}, ctxt SpecVisitorContext, node *pb.MethodMeta, cont Cont) Cont

	EnterHTTPMethodMeta(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPMethodMeta) Cont
	VisitHTTPMethodMetaChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.HTTPMethodMeta) Cont
	LeaveHTTPMethodMeta(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPMethodMeta, cont Cont) Cont

	EnterData(self interface{}, ctxt SpecVisitorContext, node *pb.Data) Cont
	VisitDataChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.Data) Cont
	LeaveData(self interface{}, ctxt SpecVisitorContext, node *pb.Data, cont Cont) Cont

	EnterDataMeta(self interface{}, ctxt SpecVisitorContext, node *pb.DataMeta) Cont
	VisitDataMetaChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.DataMeta) Cont
	LeaveDataMeta(self interface{}, ctxt SpecVisitorContext, node *pb.DataMeta, cont Cont) Cont

	EnterHTTPMeta(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPMeta) Cont
	VisitHTTPMetaChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.HTTPMeta) Cont
	LeaveHTTPMeta(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPMeta, cont Cont) Cont

	EnterHTTPPath(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPPath) Cont
	VisitHTTPPathChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.HTTPPath) Cont
	LeaveHTTPPath(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPPath, cont Cont) Cont

	EnterHTTPQuery(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPQuery) Cont
	VisitHTTPQueryChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.HTTPQuery) Cont
	LeaveHTTPQuery(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPQuery, cont Cont) Cont

	EnterHTTPHeader(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPHeader) Cont
	VisitHTTPHeaderChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.HTTPHeader) Cont
	LeaveHTTPHeader(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPHeader, cont Cont) Cont

	EnterHTTPCookie(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPCookie) Cont
	VisitHTTPCookieChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.HTTPCookie) Cont
	LeaveHTTPCookie(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPCookie, cont Cont) Cont

	EnterHTTPBody(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPBody) Cont
	VisitHTTPBodyChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.HTTPBody) Cont
	LeaveHTTPBody(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPBody, cont Cont) Cont

	EnterHTTPEmpty(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPEmpty) Cont
	VisitHTTPEmptyChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.HTTPEmpty) Cont
	LeaveHTTPEmpty(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPEmpty, cont Cont) Cont

	EnterHTTPAuth(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPAuth) Cont
	VisitHTTPAuthChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.HTTPAuth) Cont
	LeaveHTTPAuth(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPAuth, cont Cont) Cont

	EnterHTTPMultipart(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPMultipart) Cont
	VisitHTTPMultipartChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.HTTPMultipart) Cont
	LeaveHTTPMultipart(self interface{}, ctxt SpecVisitorContext, node *pb.HTTPMultipart, cont Cont) Cont

	EnterPrimitive(self interface{}, ctxt SpecVisitorContext, node *pb.Primitive) Cont
	VisitPrimitiveChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.Primitive) Cont
	LeavePrimitive(self interface{}, ctxt SpecVisitorContext, node *pb.Primitive, cont Cont) Cont

	EnterStruct(self interface{}, ctxt SpecVisitorContext, node *pb.Struct) Cont
	VisitStructChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.Struct) Cont
	LeaveStruct(self interface{}, ctxt SpecVisitorContext, node *pb.Struct, cont Cont) Cont

	EnterList(self interface{}, ctxt SpecVisitorContext, node *pb.List) Cont
	VisitListChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.List) Cont
	LeaveList(self interface{}, ctxt SpecVisitorContext, node *pb.List, cont Cont) Cont

	EnterOptional(self interface{}, ctxt SpecVisitorContext, node *pb.Optional) Cont
	VisitOptionalChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.Optional) Cont
	LeaveOptional(self interface{}, ctxt SpecVisitorContext, node *pb.Optional, cont Cont) Cont

	EnterOneOf(self interface{}, ctxt SpecVisitorContext, node *pb.OneOf) Cont
	VisitOneOfChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node *pb.OneOf) Cont
	LeaveOneOf(self interface{}, ctxt SpecVisitorContext, node *pb.OneOf, cont Cont) Cont

	// Visits the children of an unknown type.
	DefaultVisitChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node interface{}) Cont
}

// A SpecVisitor with methods for providing default visiting behaviour.
type DefaultSpecVisitor interface {
	SpecVisitor

	EnterNode(self interface{}, ctxt SpecVisitorContext, node interface{}) Cont
	VisitNodeChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node interface{}) Cont
	LeaveNode(self interface{}, ctxt SpecVisitorContext, node interface{}, cont Cont) Cont
}

// Defines nops for all visitor methods in SpecVisitor.
type DefaultSpecVisitorImpl struct{}

var _ SpecVisitor = (*DefaultSpecVisitorImpl)(nil)

func (*DefaultSpecVisitorImpl) EnterNode(self interface{}, ctxt SpecVisitorContext, node interface{}) Cont {
	return Continue
}

func (*DefaultSpecVisitorImpl) VisitNodeChildren(self interface{}, ctxt SpecVisitorContext, vm VisitorManager, node interface{}) Cont {
	return go_ast.DefaultVisitChildren(ctxt, vm, node)
}

func (*DefaultSpecVisitorImpl) LeaveNode(self interface{}, ctxt SpecVisitorContext, node interface{}, cont Cont) Cont {
	return cont
}

func (*DefaultSpecVisitorImpl) DefaultVisitChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, node interface{}) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, node)
}

// == APISpec =================================================================

func (*DefaultSpecVisitorImpl) EnterAPISpec(self interface{}, c SpecVisitorContext, spec *pb.APISpec) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, spec)
}

func (*DefaultSpecVisitorImpl) VisitAPISpecChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, spec *pb.APISpec) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, spec)
}

func (*DefaultSpecVisitorImpl) LeaveAPISpec(self interface{}, c SpecVisitorContext, spec *pb.APISpec, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, spec, cont)
}

// == Method ==================================================================

func (*DefaultSpecVisitorImpl) EnterMethod(self interface{}, c SpecVisitorContext, m *pb.Method) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, m)
}

func (*DefaultSpecVisitorImpl) VisitMethodChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, m *pb.Method) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, m)
}

func (*DefaultSpecVisitorImpl) LeaveMethod(self interface{}, c SpecVisitorContext, m *pb.Method, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, m, cont)
}

// == MethodMeta ==============================================================

func (*DefaultSpecVisitorImpl) EnterMethodMeta(self interface{}, c SpecVisitorContext, m *pb.MethodMeta) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, m)
}

func (*DefaultSpecVisitorImpl) VisitMethodMetaChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, m *pb.MethodMeta) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, m)
}

func (*DefaultSpecVisitorImpl) LeaveMethodMeta(self interface{}, c SpecVisitorContext, m *pb.MethodMeta, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, m, cont)
}

// == HTTPMethodMeta ==========================================================

func (*DefaultSpecVisitorImpl) EnterHTTPMethodMeta(self interface{}, c SpecVisitorContext, m *pb.HTTPMethodMeta) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, m)
}

func (*DefaultSpecVisitorImpl) VisitHTTPMethodMetaChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, m *pb.HTTPMethodMeta) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, m)
}

func (*DefaultSpecVisitorImpl) LeaveHTTPMethodMeta(self interface{}, c SpecVisitorContext, m *pb.HTTPMethodMeta, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, m, cont)
}

// == Data =====================================================================

func (*DefaultSpecVisitorImpl) EnterData(self interface{}, c SpecVisitorContext, d *pb.Data) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, d)
}

func (*DefaultSpecVisitorImpl) VisitDataChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, d *pb.Data) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, d)
}

func (*DefaultSpecVisitorImpl) LeaveData(self interface{}, c SpecVisitorContext, d *pb.Data, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, d, cont)
}

// == DataMeta ================================================================

func (*DefaultSpecVisitorImpl) EnterDataMeta(self interface{}, c SpecVisitorContext, d *pb.DataMeta) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, d)
}

func (*DefaultSpecVisitorImpl) VisitDataMetaChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, d *pb.DataMeta) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, d)
}

func (*DefaultSpecVisitorImpl) LeaveDataMeta(self interface{}, c SpecVisitorContext, d *pb.DataMeta, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, d, cont)
}

// == HTTPMeta ================================================================

func (*DefaultSpecVisitorImpl) EnterHTTPMeta(self interface{}, c SpecVisitorContext, m *pb.HTTPMeta) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, m)
}

func (*DefaultSpecVisitorImpl) VisitHTTPMetaChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, m *pb.HTTPMeta) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, m)
}

func (*DefaultSpecVisitorImpl) LeaveHTTPMeta(self interface{}, c SpecVisitorContext, m *pb.HTTPMeta, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, m, cont)
}

// == HTTPPath ================================================================

func (*DefaultSpecVisitorImpl) EnterHTTPPath(self interface{}, c SpecVisitorContext, p *pb.HTTPPath) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, p)
}

func (*DefaultSpecVisitorImpl) VisitHTTPPathChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, p *pb.HTTPPath) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, p)
}

func (*DefaultSpecVisitorImpl) LeaveHTTPPath(self interface{}, c SpecVisitorContext, p *pb.HTTPPath, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, p, cont)
}

// == HTTPQuery ===============================================================

func (*DefaultSpecVisitorImpl) EnterHTTPQuery(self interface{}, c SpecVisitorContext, q *pb.HTTPQuery) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, q)
}

func (*DefaultSpecVisitorImpl) VisitHTTPQueryChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, q *pb.HTTPQuery) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, q)
}

func (*DefaultSpecVisitorImpl) LeaveHTTPQuery(self interface{}, c SpecVisitorContext, q *pb.HTTPQuery, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, q, cont)
}

// == HTTPHeader ==============================================================

func (*DefaultSpecVisitorImpl) EnterHTTPHeader(self interface{}, c SpecVisitorContext, b *pb.HTTPHeader) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, b)
}

func (*DefaultSpecVisitorImpl) VisitHTTPHeaderChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, b *pb.HTTPHeader) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, b)
}

func (*DefaultSpecVisitorImpl) LeaveHTTPHeader(self interface{}, c SpecVisitorContext, b *pb.HTTPHeader, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, b, cont)
}

// == HTTPCookie ==============================================================

func (*DefaultSpecVisitorImpl) EnterHTTPCookie(self interface{}, c SpecVisitorContext, ck *pb.HTTPCookie) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, ck)
}

func (*DefaultSpecVisitorImpl) VisitHTTPCookieChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, ck *pb.HTTPCookie) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, ck)
}

func (*DefaultSpecVisitorImpl) LeaveHTTPCookie(self interface{}, c SpecVisitorContext, ck *pb.HTTPCookie, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, ck, cont)
}

// == HTTPBody ================================================================

func (*DefaultSpecVisitorImpl) EnterHTTPBody(self interface{}, c SpecVisitorContext, b *pb.HTTPBody) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, b)
}

func (*DefaultSpecVisitorImpl) VisitHTTPBodyChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, b *pb.HTTPBody) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, b)
}

func (*DefaultSpecVisitorImpl) LeaveHTTPBody(self interface{}, c SpecVisitorContext, b *pb.HTTPBody, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, b, cont)
}

// == HTTPEmpty ===============================================================

func (*DefaultSpecVisitorImpl) EnterHTTPEmpty(self interface{}, c SpecVisitorContext, e *pb.HTTPEmpty) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, e)
}

func (*DefaultSpecVisitorImpl) VisitHTTPEmptyChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, e *pb.HTTPEmpty) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, e)
}

func (*DefaultSpecVisitorImpl) LeaveHTTPEmpty(self interface{}, c SpecVisitorContext, e *pb.HTTPEmpty, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, e, cont)
}

// == HTTPAuth ================================================================

func (*DefaultSpecVisitorImpl) EnterHTTPAuth(self interface{}, c SpecVisitorContext, a *pb.HTTPAuth) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, a)
}

func (*DefaultSpecVisitorImpl) VisitHTTPAuthChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, a *pb.HTTPAuth) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, a)
}

func (*DefaultSpecVisitorImpl) LeaveHTTPAuth(self interface{}, c SpecVisitorContext, a *pb.HTTPAuth, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, a, cont)
}

// == HTTPMultipart ===========================================================

func (*DefaultSpecVisitorImpl) EnterHTTPMultipart(self interface{}, c SpecVisitorContext, m *pb.HTTPMultipart) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, m)
}

func (*DefaultSpecVisitorImpl) VisitHTTPMultipartChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, m *pb.HTTPMultipart) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, m)
}

func (*DefaultSpecVisitorImpl) LeaveHTTPMultipart(self interface{}, c SpecVisitorContext, m *pb.HTTPMultipart, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, m, cont)
}

// == Primitive ===============================================================

func (*DefaultSpecVisitorImpl) EnterPrimitive(self interface{}, c SpecVisitorContext, d *pb.Primitive) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, d)
}

func (*DefaultSpecVisitorImpl) VisitPrimitiveChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, d *pb.Primitive) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, d)
}

func (*DefaultSpecVisitorImpl) LeavePrimitive(self interface{}, c SpecVisitorContext, d *pb.Primitive, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, d, cont)
}

// == Struct ==================================================================

func (*DefaultSpecVisitorImpl) EnterStruct(self interface{}, c SpecVisitorContext, d *pb.Struct) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, d)
}

func (*DefaultSpecVisitorImpl) VisitStructChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, d *pb.Struct) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, d)
}

func (*DefaultSpecVisitorImpl) LeaveStruct(self interface{}, c SpecVisitorContext, d *pb.Struct, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, d, cont)
}

// == List =====================================================================

func (*DefaultSpecVisitorImpl) EnterList(self interface{}, c SpecVisitorContext, d *pb.List) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, d)
}

func (*DefaultSpecVisitorImpl) VisitListChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, d *pb.List) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, d)
}

func (*DefaultSpecVisitorImpl) LeaveList(self interface{}, c SpecVisitorContext, d *pb.List, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, d, cont)
}

// == Optional ================================================================

func (*DefaultSpecVisitorImpl) EnterOptional(self interface{}, c SpecVisitorContext, d *pb.Optional) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, d)
}

func (*DefaultSpecVisitorImpl) VisitOptionalChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, d *pb.Optional) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, d)
}

func (*DefaultSpecVisitorImpl) LeaveOptional(self interface{}, c SpecVisitorContext, d *pb.Optional, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, d, cont)
}

// == OneOf ===================================================================

func (*DefaultSpecVisitorImpl) EnterOneOf(self interface{}, c SpecVisitorContext, d *pb.OneOf) Cont {
	return self.(DefaultSpecVisitor).EnterNode(self, c, d)
}

func (*DefaultSpecVisitorImpl) VisitOneOfChildren(self interface{}, c SpecVisitorContext, vm VisitorManager, d *pb.OneOf) Cont {
	return self.(DefaultSpecVisitor).VisitNodeChildren(self, c, vm, d)
}

func (*DefaultSpecVisitorImpl) LeaveOneOf(self interface{}, c SpecVisitorContext, d *pb.OneOf, cont Cont) Cont {
	return self.(DefaultSpecVisitor).LeaveNode(self, c, d, cont)
}

// extendContext implementation for SpecVisitor.
func extendContext(cin Context, node interface{}) Context {
	ctx, ok := cin.(SpecVisitorContext)
	result := cin
	if !ok {
		panic(fmt.Sprintf("http_rest.extendContext expected SpecVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}

	// Dispatch on type and path.
	switch node := node.(type) {
	case pb.APISpec, pb.Method, pb.Data, pb.Primitive, pb.Struct, pb.List, pb.Optional, pb.OneOf:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		result = extendContext(ctx, &node)
	case *pb.Method:
		// Update the RestPath in the context
		meta := node.GetMeta().GetHttp()
		if meta != nil {
			ctx.setRestOperation(meta.GetMethod())
			ctx = ctx.AppendRestPath(meta.GetHost()).AppendRestPath(meta.GetMethod()).AppendRestPath(meta.GetPathTemplate())
			result = ctx
		}
	case *pb.Data:
		// Update the RestPath in the context
		// HTTPMeta is only valid for the top-level Data object.
		if node.GetMeta() != nil && node.GetMeta().GetHttp() != nil {
			ctx.setTopLevelDataIndex(len(ctx.GetRestPath()) - 1)

			meta := node.GetMeta().GetHttp()
			switch rc := meta.GetResponseCode(); rc {
			case 0: // arg
				ctx.setIsArg(true)
				ctx = ctx.AppendRestPath("Arg")
			default:
				ctx.setIsArg(false)
				ctx = ctx.AppendRestPath("Response")

				responseCode := "default"
				if rc != -1 {
					responseCode = strconv.Itoa(int(rc))
				}
				ctx = ctx.AppendRestPath(responseCode)
				ctx.setResponseCode(responseCode)
			}

			var valueKey string
			if x := meta.GetPath(); x != nil {
				ctx.setValueType(PATH)
				valueKey = x.GetKey()
			} else if x := meta.GetQuery(); x != nil {
				ctx.setValueType(QUERY)
				valueKey = x.GetKey()
			} else if x := meta.GetHeader(); x != nil {
				ctx.setValueType(HEADER)
				valueKey = x.GetKey()
			} else if x := meta.GetCookie(); x != nil {
				ctx.setValueType(COOKIE)
				valueKey = x.GetKey()
			} else if x := meta.GetBody(); x != nil {
				ctx.setValueType(BODY)
				valueKey = x.GetContentType().String()
			} else if x := meta.GetAuth(); x != nil {
				ctx.setValueType(AUTH)
				ctx.setHttpAuthType(x.GetType())
				valueKey = "Authorization"
			}

			ctx = ctx.AppendRestPath(ctx.GetValueType().String())
			ctx = ctx.AppendRestPath(valueKey)

			if node.GetOptional() != nil {
				ctx.setIsOptional()
			}

			// Do nothing for HTTPEmpty
		} else {
			astPath := ctx.GetPath()
			ctx = ctx.AppendRestPath(astPath[len(astPath)-1])
		}
		result = ctx
	}

	return result
}

// enter implementation for SpecVisitor.
func enter(cin Context, visitor interface{}, node interface{}) Cont {
	v, _ := visitor.(SpecVisitor)
	ctx, ok := extendContext(cin, node).(SpecVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.enter expected SpecVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}
	keepGoing := Continue

	// Dispatch on type and path.
	switch node := node.(type) {
	case pb.APISpec, pb.Method, pb.MethodMeta, pb.HTTPMethodMeta, pb.Data, pb.DataMeta, pb.HTTPMeta, pb.HTTPPath, pb.HTTPQuery, pb.HTTPHeader, pb.HTTPCookie, pb.HTTPBody, pb.HTTPAuth, pb.HTTPMultipart, pb.Primitive, pb.Struct, pb.List, pb.Optional, pb.OneOf:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		keepGoing = enter(ctx, visitor, &node)
	case *pb.APISpec:
		keepGoing = v.EnterAPISpec(visitor, ctx, node)
	case *pb.Method:
		keepGoing = v.EnterMethod(visitor, ctx, node)
	case *pb.MethodMeta:
		keepGoing = v.EnterMethodMeta(visitor, ctx, node)
	case *pb.HTTPMethodMeta:
		keepGoing = v.EnterHTTPMethodMeta(visitor, ctx, node)
	case *pb.Data:
		keepGoing = v.EnterData(visitor, ctx, node)
	case *pb.DataMeta:
		keepGoing = v.EnterDataMeta(visitor, ctx, node)
	case *pb.HTTPPath:
		keepGoing = v.EnterHTTPPath(visitor, ctx, node)
	case *pb.HTTPQuery:
		keepGoing = v.EnterHTTPQuery(visitor, ctx, node)
	case *pb.HTTPHeader:
		keepGoing = v.EnterHTTPHeader(visitor, ctx, node)
	case *pb.HTTPCookie:
		keepGoing = v.EnterHTTPCookie(visitor, ctx, node)
	case *pb.HTTPBody:
		keepGoing = v.EnterHTTPBody(visitor, ctx, node)
	case *pb.HTTPEmpty:
		keepGoing = v.EnterHTTPEmpty(visitor, ctx, node)
	case *pb.HTTPAuth:
		keepGoing = v.EnterHTTPAuth(visitor, ctx, node)
	case *pb.HTTPMultipart:
		keepGoing = v.EnterHTTPMultipart(visitor, ctx, node)
	case *pb.Primitive:
		keepGoing = v.EnterPrimitive(visitor, ctx, node)
	case *pb.Struct:
		keepGoing = v.EnterStruct(visitor, ctx, node)
	case *pb.List:
		keepGoing = v.EnterList(visitor, ctx, node)
	case *pb.Optional:
		keepGoing = v.EnterOptional(visitor, ctx, node)
	case *pb.OneOf:
		keepGoing = v.EnterOneOf(visitor, ctx, node)
	default:
		// Just keep going if we don't understand the type.
	}

	return keepGoing
}

// visitChildren implementation for SpecVisitor.
func visitChildren(cin Context, vm VisitorManager, node interface{}) Cont {
	visitor := vm.Visitor()
	v, _ := visitor.(SpecVisitor)
	ctx, ok := cin.(SpecVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.visitChildren expected SpecVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}

	// Dispatch on type and path.
	switch node := node.(type) {
	case pb.APISpec, pb.Method, pb.MethodMeta, pb.HTTPMethodMeta, pb.Data, pb.DataMeta, pb.HTTPMeta, pb.HTTPPath, pb.HTTPQuery, pb.HTTPHeader, pb.HTTPCookie, pb.HTTPBody, pb.HTTPAuth, pb.HTTPMultipart, pb.Primitive, pb.Struct, pb.List, pb.Optional, pb.OneOf:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		return visitChildren(ctx, vm, &node)
	case *pb.APISpec:
		return v.VisitAPISpecChildren(visitor, ctx, vm, node)
	case *pb.Method:
		return v.VisitMethodChildren(visitor, ctx, vm, node)
	case *pb.MethodMeta:
		return v.VisitMethodMetaChildren(visitor, ctx, vm, node)
	case *pb.HTTPMethodMeta:
		return v.VisitHTTPMethodMetaChildren(visitor, ctx, vm, node)
	case *pb.Data:
		return v.VisitDataChildren(visitor, ctx, vm, node)
	case *pb.DataMeta:
		return v.VisitDataMetaChildren(visitor, ctx, vm, node)
	case *pb.HTTPPath:
		return v.VisitHTTPPathChildren(visitor, ctx, vm, node)
	case *pb.HTTPQuery:
		return v.VisitHTTPQueryChildren(visitor, ctx, vm, node)
	case *pb.HTTPHeader:
		return v.VisitHTTPHeaderChildren(visitor, ctx, vm, node)
	case *pb.HTTPCookie:
		return v.VisitHTTPCookieChildren(visitor, ctx, vm, node)
	case *pb.HTTPBody:
		return v.VisitHTTPBodyChildren(visitor, ctx, vm, node)
	case *pb.HTTPEmpty:
		return v.VisitHTTPEmptyChildren(visitor, ctx, vm, node)
	case *pb.HTTPAuth:
		return v.VisitHTTPAuthChildren(visitor, ctx, vm, node)
	case *pb.HTTPMultipart:
		return v.VisitHTTPMultipartChildren(visitor, ctx, vm, node)
	case *pb.Primitive:
		return v.VisitPrimitiveChildren(visitor, ctx, vm, node)
	case *pb.Struct:
		return v.VisitStructChildren(visitor, ctx, vm, node)
	case *pb.List:
		return v.VisitListChildren(visitor, ctx, vm, node)
	case *pb.Optional:
		return v.VisitOptionalChildren(visitor, ctx, vm, node)
	case *pb.OneOf:
		return v.VisitOneOfChildren(visitor, ctx, vm, node)
	default:
		return v.DefaultVisitChildren(visitor, ctx, vm, node)
	}
}

// leave implementation for SpecVisitor.
func leave(cin Context, visitor interface{}, node interface{}, cont Cont) Cont {
	v, _ := visitor.(SpecVisitor)
	ctx, ok := extendContext(cin, node).(SpecVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.leave expected SpecVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}
	keepGoing := cont

	// Dispatch on type and path.
	switch node := node.(type) {
	case pb.APISpec, pb.Method, pb.MethodMeta, pb.HTTPMethodMeta, pb.Data, pb.DataMeta, pb.HTTPMeta, pb.HTTPPath, pb.HTTPQuery, pb.HTTPHeader, pb.HTTPCookie, pb.HTTPBody, pb.HTTPAuth, pb.HTTPMultipart, pb.Primitive, pb.Struct, pb.List, pb.Optional, pb.OneOf:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		keepGoing = leave(ctx, visitor, &node, cont)
	case *pb.APISpec:
		keepGoing = v.LeaveAPISpec(visitor, ctx, node, cont)
	case *pb.Method:
		keepGoing = v.LeaveMethod(visitor, ctx, node, cont)
	case *pb.MethodMeta:
		keepGoing = v.LeaveMethodMeta(visitor, ctx, node, cont)
	case *pb.HTTPMethodMeta:
		keepGoing = v.LeaveHTTPMethodMeta(visitor, ctx, node, cont)
	case *pb.Data:
		keepGoing = v.LeaveData(visitor, ctx, node, cont)
	case *pb.DataMeta:
		keepGoing = v.LeaveDataMeta(visitor, ctx, node, cont)
	case *pb.HTTPPath:
		keepGoing = v.LeaveHTTPPath(visitor, ctx, node, cont)
	case *pb.HTTPQuery:
		keepGoing = v.LeaveHTTPQuery(visitor, ctx, node, cont)
	case *pb.HTTPHeader:
		keepGoing = v.LeaveHTTPHeader(visitor, ctx, node, cont)
	case *pb.HTTPCookie:
		keepGoing = v.LeaveHTTPCookie(visitor, ctx, node, cont)
	case *pb.HTTPBody:
		keepGoing = v.LeaveHTTPBody(visitor, ctx, node, cont)
	case *pb.HTTPEmpty:
		keepGoing = v.LeaveHTTPEmpty(visitor, ctx, node, cont)
	case *pb.HTTPAuth:
		keepGoing = v.LeaveHTTPAuth(visitor, ctx, node, cont)
	case *pb.HTTPMultipart:
		keepGoing = v.LeaveHTTPMultipart(visitor, ctx, node, cont)
	case *pb.Primitive:
		keepGoing = v.LeavePrimitive(visitor, ctx, node, cont)
	case *pb.Struct:
		keepGoing = v.LeaveStruct(visitor, ctx, node, cont)
	case *pb.List:
		keepGoing = v.LeaveList(visitor, ctx, node, cont)
	case *pb.Optional:
		keepGoing = v.LeaveOptional(visitor, ctx, node, cont)
	case *pb.OneOf:
		keepGoing = v.LeaveOneOf(visitor, ctx, node, cont)
	default:
		// Just keep going if we don't understand the type.
	}

	return keepGoing
}

// Visits m with v.
func Apply(v SpecVisitor, m interface{}) Cont {
	c := new(httpRestSpecVisitorContext)
	vis := NewVisitorManager(c, v, enter, visitChildren, leave, extendContext)
	return go_ast.Apply(vis, m)
}

func GetPrimitiveType(p *pb.Primitive) reflect.Type {
	if t := p.GetBoolValue(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetBytesValue(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetStringValue(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetInt32Value(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetInt64Value(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetUint32Value(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetUint64Value(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetDoubleValue(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetFloatValue(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else {
		panic("Unknown primitive type")
	}
}

func GetPrimitiveValue(p *pb.Primitive) string {
	if t := p.GetBoolValue(); t != nil {
		return strconv.FormatBool(t.Value)
	} else if t := p.GetBytesValue(); t != nil {
		return string(t.Value)
	} else if t := p.GetStringValue(); t != nil {
		return t.Value
	} else if t := p.GetInt32Value(); t != nil {
		return strconv.Itoa(int(t.Value))
	} else if t := p.GetInt64Value(); t != nil {
		return strconv.Itoa(int(t.Value))
	} else if t := p.GetUint32Value(); t != nil {
		return strconv.FormatUint(uint64(t.Value), 10)
	} else if t := p.GetUint64Value(); t != nil {
		return strconv.FormatUint(t.Value, 10)
	} else if t := p.GetDoubleValue(); t != nil {
		return fmt.Sprintf("%f", t.Value)
	} else if t := p.GetFloatValue(); t != nil {
		return fmt.Sprintf("%f", t.Value)
	} else {
		panic("Unknown primitive type")
	}
}

type PrintVisitor struct {
	DefaultSpecVisitorImpl
}

func (*PrintVisitor) EnterData(ctx SpecVisitorContext, d *pb.Data) Cont {
	fmt.Printf("%s %s\n", strings.Join(ctx.GetRestPath(), "."), d)
	return Continue
}

func (*PrintVisitor) EnterPrimitive(ctx SpecVisitorContext, p *pb.Primitive) Cont {
	fmt.Printf("%s %s (%s)\n", strings.Join(ctx.GetRestPath(), "."), GetPrimitiveValue(p), GetPrimitiveType(p))
	return Continue
}
