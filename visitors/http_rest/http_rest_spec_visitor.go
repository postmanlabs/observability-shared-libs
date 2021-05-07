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
type HttpRestSpecVisitor interface {
	EnterAPISpec(HttpRestSpecVisitorContext, *pb.APISpec) Cont
	VisitAPISpecChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.APISpec) Cont
	LeaveAPISpec(HttpRestSpecVisitorContext, *pb.APISpec, Cont) Cont

	EnterMethod(HttpRestSpecVisitorContext, *pb.Method) Cont
	VisitMethodChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.Method) Cont
	LeaveMethod(HttpRestSpecVisitorContext, *pb.Method, Cont) Cont

	EnterMethodMeta(HttpRestSpecVisitorContext, *pb.MethodMeta) Cont
	VisitMethodMetaChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.MethodMeta) Cont
	LeaveMethodMeta(HttpRestSpecVisitorContext, *pb.MethodMeta, Cont) Cont

	EnterHTTPMethodMeta(HttpRestSpecVisitorContext, *pb.HTTPMethodMeta) Cont
	VisitHTTPMethodMetaChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.HTTPMethodMeta) Cont
	LeaveHTTPMethodMeta(HttpRestSpecVisitorContext, *pb.HTTPMethodMeta, Cont) Cont

	EnterData(HttpRestSpecVisitorContext, *pb.Data) Cont
	VisitDataChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.Data) Cont
	LeaveData(HttpRestSpecVisitorContext, *pb.Data, Cont) Cont

	EnterDataMeta(HttpRestSpecVisitorContext, *pb.DataMeta) Cont
	VisitDataMetaChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.DataMeta) Cont
	LeaveDataMeta(HttpRestSpecVisitorContext, *pb.DataMeta, Cont) Cont

	EnterHTTPMeta(HttpRestSpecVisitorContext, *pb.HTTPMeta) Cont
	VisitHTTPMetaChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.HTTPMeta) Cont
	LeaveHTTPMeta(HttpRestSpecVisitorContext, *pb.HTTPMeta, Cont) Cont

	EnterHTTPPath(HttpRestSpecVisitorContext, *pb.HTTPPath) Cont
	VisitHTTPPathChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.HTTPPath) Cont
	LeaveHTTPPath(HttpRestSpecVisitorContext, *pb.HTTPPath, Cont) Cont

	EnterHTTPQuery(HttpRestSpecVisitorContext, *pb.HTTPQuery) Cont
	VisitHTTPQueryChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.HTTPQuery) Cont
	LeaveHTTPQuery(HttpRestSpecVisitorContext, *pb.HTTPQuery, Cont) Cont

	EnterHTTPHeader(HttpRestSpecVisitorContext, *pb.HTTPHeader) Cont
	VisitHTTPHeaderChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.HTTPHeader) Cont
	LeaveHTTPHeader(HttpRestSpecVisitorContext, *pb.HTTPHeader, Cont) Cont

	EnterHTTPCookie(HttpRestSpecVisitorContext, *pb.HTTPCookie) Cont
	VisitHTTPCookieChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.HTTPCookie) Cont
	LeaveHTTPCookie(HttpRestSpecVisitorContext, *pb.HTTPCookie, Cont) Cont

	EnterHTTPBody(HttpRestSpecVisitorContext, *pb.HTTPBody) Cont
	VisitHTTPBodyChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.HTTPBody) Cont
	LeaveHTTPBody(HttpRestSpecVisitorContext, *pb.HTTPBody, Cont) Cont

	EnterHTTPEmpty(HttpRestSpecVisitorContext, *pb.HTTPEmpty) Cont
	VisitHTTPEmptyChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.HTTPEmpty) Cont
	LeaveHTTPEmpty(HttpRestSpecVisitorContext, *pb.HTTPEmpty, Cont) Cont

	EnterHTTPAuth(HttpRestSpecVisitorContext, *pb.HTTPAuth) Cont
	VisitHTTPAuthChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.HTTPAuth) Cont
	LeaveHTTPAuth(HttpRestSpecVisitorContext, *pb.HTTPAuth, Cont) Cont

	EnterHTTPMultipart(HttpRestSpecVisitorContext, *pb.HTTPMultipart) Cont
	VisitHTTPMultipartChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.HTTPMultipart) Cont
	LeaveHTTPMultipart(HttpRestSpecVisitorContext, *pb.HTTPMultipart, Cont) Cont

	EnterPrimitive(HttpRestSpecVisitorContext, *pb.Primitive) Cont
	VisitPrimitiveChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.Primitive) Cont
	LeavePrimitive(HttpRestSpecVisitorContext, *pb.Primitive, Cont) Cont

	EnterStruct(HttpRestSpecVisitorContext, *pb.Struct) Cont
	VisitStructChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.Struct) Cont
	LeaveStruct(HttpRestSpecVisitorContext, *pb.Struct, Cont) Cont

	EnterList(HttpRestSpecVisitorContext, *pb.List) Cont
	VisitListChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.List) Cont
	LeaveList(HttpRestSpecVisitorContext, *pb.List, Cont) Cont

	EnterOptional(HttpRestSpecVisitorContext, *pb.Optional) Cont
	VisitOptionalChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.Optional) Cont
	LeaveOptional(HttpRestSpecVisitorContext, *pb.Optional, Cont) Cont

	EnterOneOf(HttpRestSpecVisitorContext, *pb.OneOf) Cont
	VisitOneOfChildren(HttpRestSpecVisitorContext, VisitorManager, *pb.OneOf) Cont
	LeaveOneOf(HttpRestSpecVisitorContext, *pb.OneOf, Cont) Cont

	DefaultVisitChildren(HttpRestSpecVisitorContext, VisitorManager, interface{}) Cont
}

// Defines nops for all visitor methods in HttpRestSpecVisitor.
type DefaultHttpRestSpecVisitor struct{}

func (*DefaultHttpRestSpecVisitor) DefaultVisitChildren(c HttpRestSpecVisitorContext, vm VisitorManager, node interface{}) Cont {
	return go_ast.DefaultVisitChildren(c, vm, node)
}

// == APISpec =================================================================

func (*DefaultHttpRestSpecVisitor) EnterAPISpec(c HttpRestSpecVisitorContext, spec *pb.APISpec) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitAPISpecChildren(c HttpRestSpecVisitorContext, vm VisitorManager, spec *pb.APISpec) Cont {
	return go_ast.DefaultVisitChildren(c, vm, spec)
}

func (*DefaultHttpRestSpecVisitor) LeaveAPISpec(c HttpRestSpecVisitorContext, spec *pb.APISpec, cont Cont) Cont {
	return cont
}

// == Method ==================================================================

func (*DefaultHttpRestSpecVisitor) EnterMethod(c HttpRestSpecVisitorContext, m *pb.Method) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitMethodChildren(c HttpRestSpecVisitorContext, vm VisitorManager, m *pb.Method) Cont {
	return go_ast.DefaultVisitChildren(c, vm, m)
}

func (*DefaultHttpRestSpecVisitor) LeaveMethod(c HttpRestSpecVisitorContext, m *pb.Method, cont Cont) Cont {
	return cont
}

// == MethodMeta ==============================================================

func (*DefaultHttpRestSpecVisitor) EnterMethodMeta(c HttpRestSpecVisitorContext, m *pb.MethodMeta) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitMethodMetaChildren(c HttpRestSpecVisitorContext, vm VisitorManager, m *pb.MethodMeta) Cont {
	return go_ast.DefaultVisitChildren(c, vm, m)
}

func (*DefaultHttpRestSpecVisitor) LeaveMethodMeta(c HttpRestSpecVisitorContext, m *pb.MethodMeta, cont Cont) Cont {
	return cont
}

// == HTTPMethodMeta ==========================================================

func (*DefaultHttpRestSpecVisitor) EnterHTTPMethodMeta(c HttpRestSpecVisitorContext, m *pb.HTTPMethodMeta) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitHTTPMethodMetaChildren(c HttpRestSpecVisitorContext, vm VisitorManager, m *pb.HTTPMethodMeta) Cont {
	return go_ast.DefaultVisitChildren(c, vm, m)
}

func (*DefaultHttpRestSpecVisitor) LeaveHTTPMethodMeta(c HttpRestSpecVisitorContext, m *pb.HTTPMethodMeta, cont Cont) Cont {
	return cont
}

// == Data =====================================================================

func (*DefaultHttpRestSpecVisitor) EnterData(c HttpRestSpecVisitorContext, d *pb.Data) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitDataChildren(c HttpRestSpecVisitorContext, vm VisitorManager, d *pb.Data) Cont {
	return go_ast.DefaultVisitChildren(c, vm, d)
}

func (*DefaultHttpRestSpecVisitor) LeaveData(c HttpRestSpecVisitorContext, d *pb.Data, cont Cont) Cont {
	return cont
}

// == DataMeta ================================================================

func (*DefaultHttpRestSpecVisitor) EnterDataMeta(c HttpRestSpecVisitorContext, d *pb.DataMeta) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitDataMetaChildren(c HttpRestSpecVisitorContext, vm VisitorManager, d *pb.DataMeta) Cont {
	return go_ast.DefaultVisitChildren(c, vm, d)
}

func (*DefaultHttpRestSpecVisitor) LeaveDataMeta(c HttpRestSpecVisitorContext, d *pb.DataMeta, cont Cont) Cont {
	return cont
}

// == HTTPMeta ================================================================

func (*DefaultHttpRestSpecVisitor) EnterHTTPMeta(c HttpRestSpecVisitorContext, m *pb.HTTPMeta) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitHTTPMetaChildren(c HttpRestSpecVisitorContext, vm VisitorManager, m *pb.HTTPMeta) Cont {
	return go_ast.DefaultVisitChildren(c, vm, m)
}

func (*DefaultHttpRestSpecVisitor) LeaveHTTPMeta(c HttpRestSpecVisitorContext, m *pb.HTTPMeta, cont Cont) Cont {
	return cont
}

// == HTTPPath ================================================================

func (*DefaultHttpRestSpecVisitor) EnterHTTPPath(c HttpRestSpecVisitorContext, p *pb.HTTPPath) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitHTTPPathChildren(c HttpRestSpecVisitorContext, vm VisitorManager, p *pb.HTTPPath) Cont {
	return go_ast.DefaultVisitChildren(c, vm, p)
}

func (*DefaultHttpRestSpecVisitor) LeaveHTTPPath(c HttpRestSpecVisitorContext, p *pb.HTTPPath, cont Cont) Cont {
	return cont
}

// == HTTPQuery ===============================================================

func (*DefaultHttpRestSpecVisitor) EnterHTTPQuery(c HttpRestSpecVisitorContext, q *pb.HTTPQuery) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitHTTPQueryChildren(c HttpRestSpecVisitorContext, vm VisitorManager, q *pb.HTTPQuery) Cont {
	return go_ast.DefaultVisitChildren(c, vm, q)
}

func (*DefaultHttpRestSpecVisitor) LeaveHTTPQuery(c HttpRestSpecVisitorContext, q *pb.HTTPQuery, cont Cont) Cont {
	return cont
}

// == HTTPHeader ==============================================================

func (*DefaultHttpRestSpecVisitor) EnterHTTPHeader(c HttpRestSpecVisitorContext, b *pb.HTTPHeader) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitHTTPHeaderChildren(c HttpRestSpecVisitorContext, vm VisitorManager, b *pb.HTTPHeader) Cont {
	return go_ast.DefaultVisitChildren(c, vm, b)
}

func (*DefaultHttpRestSpecVisitor) LeaveHTTPHeader(c HttpRestSpecVisitorContext, b *pb.HTTPHeader, cont Cont) Cont {
	return cont
}

// == HTTPCookie ==============================================================

func (*DefaultHttpRestSpecVisitor) EnterHTTPCookie(c HttpRestSpecVisitorContext, ck *pb.HTTPCookie) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitHTTPCookieChildren(c HttpRestSpecVisitorContext, vm VisitorManager, ck *pb.HTTPCookie) Cont {
	return go_ast.DefaultVisitChildren(c, vm, ck)
}

func (*DefaultHttpRestSpecVisitor) LeaveHTTPCookie(c HttpRestSpecVisitorContext, ck *pb.HTTPCookie, cont Cont) Cont {
	return cont
}

// == HTTPBody ================================================================

func (*DefaultHttpRestSpecVisitor) EnterHTTPBody(c HttpRestSpecVisitorContext, b *pb.HTTPBody) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitHTTPBodyChildren(c HttpRestSpecVisitorContext, vm VisitorManager, b *pb.HTTPBody) Cont {
	return go_ast.DefaultVisitChildren(c, vm, b)
}

func (*DefaultHttpRestSpecVisitor) LeaveHTTPBody(c HttpRestSpecVisitorContext, b *pb.HTTPBody, cont Cont) Cont {
	return cont
}

// == HTTPEmpty ===============================================================

func (*DefaultHttpRestSpecVisitor) EnterHTTPEmpty(c HttpRestSpecVisitorContext, e *pb.HTTPEmpty) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitHTTPEmptyChildren(c HttpRestSpecVisitorContext, vm VisitorManager, e *pb.HTTPEmpty) Cont {
	return go_ast.DefaultVisitChildren(c, vm, e)
}

func (*DefaultHttpRestSpecVisitor) LeaveHTTPEmpty(c HttpRestSpecVisitorContext, e *pb.HTTPEmpty, cont Cont) Cont {
	return cont
}

// == HTTPAuth ================================================================

func (*DefaultHttpRestSpecVisitor) EnterHTTPAuth(c HttpRestSpecVisitorContext, a *pb.HTTPAuth) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitHTTPAuthChildren(c HttpRestSpecVisitorContext, vm VisitorManager, a *pb.HTTPAuth) Cont {
	return go_ast.DefaultVisitChildren(c, vm, a)
}

func (*DefaultHttpRestSpecVisitor) LeaveHTTPAuth(c HttpRestSpecVisitorContext, a *pb.HTTPAuth, cont Cont) Cont {
	return cont
}

// == HTTPMultipart ===========================================================

func (*DefaultHttpRestSpecVisitor) EnterHTTPMultipart(c HttpRestSpecVisitorContext, m *pb.HTTPMultipart) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitHTTPMultipartChildren(c HttpRestSpecVisitorContext, vm VisitorManager, m *pb.HTTPMultipart) Cont {
	return go_ast.DefaultVisitChildren(c, vm, m)
}

func (*DefaultHttpRestSpecVisitor) LeaveHTTPMultipart(c HttpRestSpecVisitorContext, m *pb.HTTPMultipart, cont Cont) Cont {
	return cont
}

// == Primitive ===============================================================

func (*DefaultHttpRestSpecVisitor) EnterPrimitive(c HttpRestSpecVisitorContext, d *pb.Primitive) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitPrimitiveChildren(c HttpRestSpecVisitorContext, vm VisitorManager, d *pb.Primitive) Cont {
	return go_ast.DefaultVisitChildren(c, vm, d)
}

func (*DefaultHttpRestSpecVisitor) LeavePrimitive(c HttpRestSpecVisitorContext, d *pb.Primitive, cont Cont) Cont {
	return cont
}

// == Struct ==================================================================

func (*DefaultHttpRestSpecVisitor) EnterStruct(c HttpRestSpecVisitorContext, d *pb.Struct) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitStructChildren(c HttpRestSpecVisitorContext, vm VisitorManager, d *pb.Struct) Cont {
	return go_ast.DefaultVisitChildren(c, vm, d)
}

func (*DefaultHttpRestSpecVisitor) LeaveStruct(c HttpRestSpecVisitorContext, d *pb.Struct, cont Cont) Cont {
	return cont
}

// == List =====================================================================

func (*DefaultHttpRestSpecVisitor) EnterList(c HttpRestSpecVisitorContext, d *pb.List) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitListChildren(c HttpRestSpecVisitorContext, vm VisitorManager, d *pb.List) Cont {
	return go_ast.DefaultVisitChildren(c, vm, d)
}

func (*DefaultHttpRestSpecVisitor) LeaveList(c HttpRestSpecVisitorContext, d *pb.List, cont Cont) Cont {
	return cont
}

// == Optional ================================================================

func (*DefaultHttpRestSpecVisitor) EnterOptional(c HttpRestSpecVisitorContext, d *pb.Optional) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitOptionalChildren(c HttpRestSpecVisitorContext, vm VisitorManager, d *pb.Optional) Cont {
	return go_ast.DefaultVisitChildren(c, vm, d)
}

func (*DefaultHttpRestSpecVisitor) LeaveOptional(c HttpRestSpecVisitorContext, d *pb.Optional, cont Cont) Cont {
	return cont
}

// == OneOf ===================================================================

func (*DefaultHttpRestSpecVisitor) EnterOneOf(c HttpRestSpecVisitorContext, d *pb.OneOf) Cont {
	return Continue
}

func (*DefaultHttpRestSpecVisitor) VisitOneOfChildren(c HttpRestSpecVisitorContext, vm VisitorManager, d *pb.OneOf) Cont {
	return go_ast.DefaultVisitChildren(c, vm, d)
}

func (*DefaultHttpRestSpecVisitor) LeaveOneOf(c HttpRestSpecVisitorContext, d *pb.OneOf, cont Cont) Cont {
	return cont
}

// extendContext implementation for HttpRestSpecVisitor.
func extendContext(cin Context, node interface{}) Context {
	ctx, ok := cin.(HttpRestSpecVisitorContext)
	result := cin
	if !ok {
		panic(fmt.Sprintf("http_rest.extendContext expected HttpRestSpecVisitorContext, got %s",
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
				if rc == -1 {
					ctx = ctx.AppendRestPath("default")
				} else {
					ctx = ctx.AppendRestPath(strconv.Itoa(int(rc)))
				}
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
			}

			ctx = ctx.AppendRestPath(ctx.GetValueType().String())
			ctx = ctx.AppendRestPath(valueKey)

			// Do nothing for HTTPEmpty
		} else {
			astPath := ctx.GetPath()
			ctx = ctx.AppendRestPath(astPath[len(astPath)-1])
		}
		result = ctx
	}

	return result
}

// enter implementation for HttpRestSpecVisitor.
func enter(cin Context, visitor interface{}, node interface{}) Cont {
	v, _ := visitor.(HttpRestSpecVisitor)
	ctx, ok := extendContext(cin, node).(HttpRestSpecVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.enter expected HttpRestSpecVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}
	keepGoing := Continue

	// Dispatch on type and path.
	switch node := node.(type) {
	case pb.APISpec, pb.Method, pb.MethodMeta, pb.HTTPMethodMeta, pb.Data, pb.DataMeta, pb.HTTPMeta, pb.HTTPPath, pb.HTTPQuery, pb.HTTPHeader, pb.HTTPCookie, pb.HTTPBody, pb.HTTPAuth, pb.HTTPMultipart, pb.Primitive, pb.Struct, pb.List, pb.Optional, pb.OneOf:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		keepGoing = enter(ctx, v, &node)
	case *pb.APISpec:
		keepGoing = v.EnterAPISpec(ctx, node)
	case *pb.Method:
		keepGoing = v.EnterMethod(ctx, node)
	case *pb.MethodMeta:
		keepGoing = v.EnterMethodMeta(ctx, node)
	case *pb.HTTPMethodMeta:
		keepGoing = v.EnterHTTPMethodMeta(ctx, node)
	case *pb.Data:
		keepGoing = v.EnterData(ctx, node)
	case *pb.DataMeta:
		keepGoing = v.EnterDataMeta(ctx, node)
	case *pb.HTTPPath:
		keepGoing = v.EnterHTTPPath(ctx, node)
	case *pb.HTTPQuery:
		keepGoing = v.EnterHTTPQuery(ctx, node)
	case *pb.HTTPHeader:
		keepGoing = v.EnterHTTPHeader(ctx, node)
	case *pb.HTTPCookie:
		keepGoing = v.EnterHTTPCookie(ctx, node)
	case *pb.HTTPBody:
		keepGoing = v.EnterHTTPBody(ctx, node)
	case *pb.HTTPEmpty:
		keepGoing = v.EnterHTTPEmpty(ctx, node)
	case *pb.HTTPAuth:
		keepGoing = v.EnterHTTPAuth(ctx, node)
	case *pb.HTTPMultipart:
		keepGoing = v.EnterHTTPMultipart(ctx, node)
	case *pb.Primitive:
		keepGoing = v.EnterPrimitive(ctx, node)
	case *pb.Struct:
		keepGoing = v.EnterStruct(ctx, node)
	case *pb.List:
		keepGoing = v.EnterList(ctx, node)
	case *pb.Optional:
		keepGoing = v.EnterOptional(ctx, node)
	case *pb.OneOf:
		keepGoing = v.EnterOneOf(ctx, node)
	default:
		// Just keep going if we don't understand the type.
	}

	return keepGoing
}

// visitChildren implementation for HttpRestSpecVisitor.
func visitChildren(cin Context, vm VisitorManager, node interface{}) Cont {
	visitor := vm.Visitor()
	v, _ := visitor.(HttpRestSpecVisitor)
	ctx, ok := cin.(HttpRestSpecVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.visitChildren expected HttpRestSpecVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}

	// Dispatch on type and path.
	switch node := node.(type) {
	case pb.APISpec, pb.Method, pb.MethodMeta, pb.HTTPMethodMeta, pb.Data, pb.DataMeta, pb.HTTPMeta, pb.HTTPPath, pb.HTTPQuery, pb.HTTPHeader, pb.HTTPCookie, pb.HTTPBody, pb.HTTPAuth, pb.HTTPMultipart, pb.Primitive, pb.Struct, pb.List, pb.Optional, pb.OneOf:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		return visitChildren(ctx, vm, &node)
	case *pb.APISpec:
		return v.VisitAPISpecChildren(ctx, vm, node)
	case *pb.Method:
		return v.VisitMethodChildren(ctx, vm, node)
	case *pb.MethodMeta:
		return v.VisitMethodMetaChildren(ctx, vm, node)
	case *pb.HTTPMethodMeta:
		return v.VisitHTTPMethodMetaChildren(ctx, vm, node)
	case *pb.Data:
		return v.VisitDataChildren(ctx, vm, node)
	case *pb.DataMeta:
		return v.VisitDataMetaChildren(ctx, vm, node)
	case *pb.HTTPPath:
		return v.VisitHTTPPathChildren(ctx, vm, node)
	case *pb.HTTPQuery:
		return v.VisitHTTPQueryChildren(ctx, vm, node)
	case *pb.HTTPHeader:
		return v.VisitHTTPHeaderChildren(ctx, vm, node)
	case *pb.HTTPCookie:
		return v.VisitHTTPCookieChildren(ctx, vm, node)
	case *pb.HTTPBody:
		return v.VisitHTTPBodyChildren(ctx, vm, node)
	case *pb.HTTPEmpty:
		return v.VisitHTTPEmptyChildren(ctx, vm, node)
	case *pb.HTTPAuth:
		return v.VisitHTTPAuthChildren(ctx, vm, node)
	case *pb.HTTPMultipart:
		return v.VisitHTTPMultipartChildren(ctx, vm, node)
	case *pb.Primitive:
		return v.VisitPrimitiveChildren(ctx, vm, node)
	case *pb.Struct:
		return v.VisitStructChildren(ctx, vm, node)
	case *pb.List:
		return v.VisitListChildren(ctx, vm, node)
	case *pb.Optional:
		return v.VisitOptionalChildren(ctx, vm, node)
	case *pb.OneOf:
		return v.VisitOneOfChildren(ctx, vm, node)
	default:
		return v.DefaultVisitChildren(ctx, vm, node)
	}
}

// leave implementation for HttpRestSpecVisitor.
func leave(cin Context, visitor interface{}, node interface{}, cont Cont) Cont {
	v, _ := visitor.(HttpRestSpecVisitor)
	ctx, ok := extendContext(cin, node).(HttpRestSpecVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.leave expected HttpRestSpecVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}
	keepGoing := cont

	// Dispatch on type and path.
	switch node := node.(type) {
	case pb.APISpec, pb.Method, pb.MethodMeta, pb.HTTPMethodMeta, pb.Data, pb.DataMeta, pb.HTTPMeta, pb.HTTPPath, pb.HTTPQuery, pb.HTTPHeader, pb.HTTPCookie, pb.HTTPBody, pb.HTTPAuth, pb.HTTPMultipart, pb.Primitive, pb.Struct, pb.List, pb.Optional, pb.OneOf:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		keepGoing = leave(ctx, v, &node, cont)
	case *pb.APISpec:
		keepGoing = v.LeaveAPISpec(ctx, node, cont)
	case *pb.Method:
		keepGoing = v.LeaveMethod(ctx, node, cont)
	case *pb.MethodMeta:
		keepGoing = v.LeaveMethodMeta(ctx, node, cont)
	case *pb.HTTPMethodMeta:
		keepGoing = v.LeaveHTTPMethodMeta(ctx, node, cont)
	case *pb.Data:
		keepGoing = v.LeaveData(ctx, node, cont)
	case *pb.DataMeta:
		keepGoing = v.LeaveDataMeta(ctx, node, cont)
	case *pb.HTTPPath:
		keepGoing = v.LeaveHTTPPath(ctx, node, cont)
	case *pb.HTTPQuery:
		keepGoing = v.LeaveHTTPQuery(ctx, node, cont)
	case *pb.HTTPHeader:
		keepGoing = v.LeaveHTTPHeader(ctx, node, cont)
	case *pb.HTTPCookie:
		keepGoing = v.LeaveHTTPCookie(ctx, node, cont)
	case *pb.HTTPBody:
		keepGoing = v.LeaveHTTPBody(ctx, node, cont)
	case *pb.HTTPEmpty:
		keepGoing = v.LeaveHTTPEmpty(ctx, node, cont)
	case *pb.HTTPAuth:
		keepGoing = v.LeaveHTTPAuth(ctx, node, cont)
	case *pb.HTTPMultipart:
		keepGoing = v.LeaveHTTPMultipart(ctx, node, cont)
	case *pb.Primitive:
		keepGoing = v.LeavePrimitive(ctx, node, cont)
	case *pb.Struct:
		keepGoing = v.LeaveStruct(ctx, node, cont)
	case *pb.List:
		keepGoing = v.LeaveList(ctx, node, cont)
	case *pb.Optional:
		keepGoing = v.LeaveOptional(ctx, node, cont)
	case *pb.OneOf:
		keepGoing = v.LeaveOneOf(ctx, node, cont)
	default:
		// Just keep going if we don't understand the type.
	}

	return keepGoing
}

// Visits m with v.
func Apply(v HttpRestSpecVisitor, m interface{}) Cont {
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
	DefaultHttpRestSpecVisitor
}

func (*PrintVisitor) EnterData(ctx HttpRestSpecVisitorContext, d *pb.Data) Cont {
	fmt.Printf("%s %s\n", strings.Join(ctx.GetRestPath(), "."), d)
	return Continue
}

func (*PrintVisitor) EnterPrimitive(ctx HttpRestSpecVisitorContext, p *pb.Primitive) Cont {
	fmt.Printf("%s %s (%s)\n", strings.Join(ctx.GetRestPath(), "."), GetPrimitiveValue(p), GetPrimitiveType(p))
	return Continue
}
