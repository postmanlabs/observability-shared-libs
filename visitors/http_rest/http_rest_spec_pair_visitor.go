package http_rest

import (
	"fmt"
	"reflect"
	"runtime"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	. "github.com/akitasoftware/akita-libs/visitors"
	"github.com/akitasoftware/akita-libs/visitors/go_ast_pair"
)

// A PairVisitorManager that lets you read each message in a pair of APISpecs,
// starting with the APISpec messages themselves. When the visitor encounters
// a type difference between the two halves of the pair, EnterDifferentTypes
// and LeaveDifferentTypes is used to enter and leave the nodes, but the nodes'
// children are not visited; EnterDifferentTypes must never return Continue.
type HttpRestSpecPairVisitor interface {
	EnterAPISpecs(HttpRestSpecPairVisitorContext, *pb.APISpec, *pb.APISpec) Cont
	VisitAPISpecChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.APISpec, *pb.APISpec) Cont
	LeaveAPISpecs(HttpRestSpecPairVisitorContext, *pb.APISpec, *pb.APISpec, Cont) Cont

	EnterMethods(HttpRestSpecPairVisitorContext, *pb.Method, *pb.Method) Cont
	VisitMethodChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.Method, *pb.Method) Cont
	LeaveMethods(HttpRestSpecPairVisitorContext, *pb.Method, *pb.Method, Cont) Cont

	EnterMethodMetas(HttpRestSpecPairVisitorContext, *pb.MethodMeta, *pb.MethodMeta) Cont
	VisitMethodMetaChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.MethodMeta, *pb.MethodMeta) Cont
	LeaveMethodMetas(HttpRestSpecPairVisitorContext, *pb.MethodMeta, *pb.MethodMeta, Cont) Cont

	EnterHTTPMethodMetas(HttpRestSpecPairVisitorContext, *pb.HTTPMethodMeta, *pb.HTTPMethodMeta) Cont
	VisitHTTPMethodMetaChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.HTTPMethodMeta, *pb.HTTPMethodMeta) Cont
	LeaveHTTPMethodMetas(HttpRestSpecPairVisitorContext, *pb.HTTPMethodMeta, *pb.HTTPMethodMeta, Cont) Cont

	EnterData(HttpRestSpecPairVisitorContext, *pb.Data, *pb.Data) Cont
	VisitDataChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.Data, *pb.Data) Cont
	LeaveData(HttpRestSpecPairVisitorContext, *pb.Data, *pb.Data, Cont) Cont

	EnterDataMetas(HttpRestSpecPairVisitorContext, *pb.DataMeta, *pb.DataMeta) Cont
	VisitDataMetaChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.DataMeta, *pb.DataMeta) Cont
	LeaveDataMetas(HttpRestSpecPairVisitorContext, *pb.DataMeta, *pb.DataMeta, Cont) Cont

	EnterHTTPMetas(HttpRestSpecPairVisitorContext, *pb.HTTPMeta, *pb.HTTPMeta) Cont
	VisitHTTPMetaChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.HTTPMeta, *pb.HTTPMeta) Cont
	LeaveHTTPMetas(HttpRestSpecPairVisitorContext, *pb.HTTPMeta, *pb.HTTPMeta, Cont) Cont

	EnterHTTPPaths(HttpRestSpecPairVisitorContext, *pb.HTTPPath, *pb.HTTPPath) Cont
	VisitHTTPPathChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.HTTPPath, *pb.HTTPPath) Cont
	LeaveHTTPPaths(HttpRestSpecPairVisitorContext, *pb.HTTPPath, *pb.HTTPPath, Cont) Cont

	EnterHTTPQueries(HttpRestSpecPairVisitorContext, *pb.HTTPQuery, *pb.HTTPQuery) Cont
	VisitHTTPQueryChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.HTTPQuery, *pb.HTTPQuery) Cont
	LeaveHTTPQueries(HttpRestSpecPairVisitorContext, *pb.HTTPQuery, *pb.HTTPQuery, Cont) Cont

	EnterHTTPHeaders(HttpRestSpecPairVisitorContext, *pb.HTTPHeader, *pb.HTTPHeader) Cont
	VisitHTTPHeaderChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.HTTPHeader, *pb.HTTPHeader) Cont
	LeaveHTTPHeaders(HttpRestSpecPairVisitorContext, *pb.HTTPHeader, *pb.HTTPHeader, Cont) Cont

	EnterHTTPCookies(HttpRestSpecPairVisitorContext, *pb.HTTPCookie, *pb.HTTPCookie) Cont
	VisitHTTPCookieChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.HTTPCookie, *pb.HTTPCookie) Cont
	LeaveHTTPCookies(HttpRestSpecPairVisitorContext, *pb.HTTPCookie, *pb.HTTPCookie, Cont) Cont

	EnterHTTPBodies(HttpRestSpecPairVisitorContext, *pb.HTTPBody, *pb.HTTPBody) Cont
	VisitHTTPBodyChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.HTTPBody, *pb.HTTPBody) Cont
	LeaveHTTPBodies(HttpRestSpecPairVisitorContext, *pb.HTTPBody, *pb.HTTPBody, Cont) Cont

	EnterHTTPEmpties(HttpRestSpecPairVisitorContext, *pb.HTTPEmpty, *pb.HTTPEmpty) Cont
	VisitHTTPEmptyChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.HTTPEmpty, *pb.HTTPEmpty) Cont
	LeaveHTTPEmpties(HttpRestSpecPairVisitorContext, *pb.HTTPEmpty, *pb.HTTPEmpty, Cont) Cont

	EnterHTTPAuths(HttpRestSpecPairVisitorContext, *pb.HTTPAuth, *pb.HTTPAuth) Cont
	VisitHTTPAuthChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.HTTPAuth, *pb.HTTPAuth) Cont
	LeaveHTTPAuths(HttpRestSpecPairVisitorContext, *pb.HTTPAuth, *pb.HTTPAuth, Cont) Cont

	EnterHTTPMultiparts(HttpRestSpecPairVisitorContext, *pb.HTTPMultipart, *pb.HTTPMultipart) Cont
	VisitHTTPMultipartChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.HTTPMultipart, *pb.HTTPMultipart) Cont
	LeaveHTTPMultiparts(HttpRestSpecPairVisitorContext, *pb.HTTPMultipart, *pb.HTTPMultipart, Cont) Cont

	EnterPrimitives(HttpRestSpecPairVisitorContext, *pb.Primitive, *pb.Primitive) Cont
	VisitPrimitiveChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.Primitive, *pb.Primitive) Cont
	LeavePrimitives(HttpRestSpecPairVisitorContext, *pb.Primitive, *pb.Primitive, Cont) Cont

	EnterStructs(HttpRestSpecPairVisitorContext, *pb.Struct, *pb.Struct) Cont
	VisitStructChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.Struct, *pb.Struct) Cont
	LeaveStructs(HttpRestSpecPairVisitorContext, *pb.Struct, *pb.Struct, Cont) Cont

	EnterLists(HttpRestSpecPairVisitorContext, *pb.List, *pb.List) Cont
	VisitListChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.List, *pb.List) Cont
	LeaveLists(HttpRestSpecPairVisitorContext, *pb.List, *pb.List, Cont) Cont

	EnterOptionals(HttpRestSpecPairVisitorContext, *pb.Optional, *pb.Optional) Cont
	VisitOptionalChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.Optional, *pb.Optional) Cont
	LeaveOptionals(HttpRestSpecPairVisitorContext, *pb.Optional, *pb.Optional, Cont) Cont

	EnterOneOfs(HttpRestSpecPairVisitorContext, *pb.OneOf, *pb.OneOf) Cont
	VisitOneOfChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, *pb.OneOf, *pb.OneOf) Cont
	LeaveOneOfs(HttpRestSpecPairVisitorContext, *pb.OneOf, *pb.OneOf, Cont) Cont

	DefaultVisitChildren(HttpRestSpecPairVisitorContext, PairVisitorManager, interface{}, interface{}) Cont

	// Used when the visitor tries to enter two nodes with different types. This
	// cannot return Continue; otherwise, visitChildren will panic.
	EnterDifferentTypes(HttpRestSpecPairVisitorContext, interface{}, interface{}) Cont

	// Used when the visitor tries to leave two nodes with different types.
	LeaveDifferentTypes(HttpRestSpecPairVisitorContext, interface{}, interface{}, Cont) Cont
}

type DefaultHttpRestSpecPairVisitor struct{}

func (*DefaultHttpRestSpecPairVisitor) DefaultVisitChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right interface{}) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

// == APISpec =================================================================

func (*DefaultHttpRestSpecPairVisitor) EnterAPISpecs(c HttpRestSpecPairVisitorContext, left, right *pb.APISpec) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitAPISpecChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.APISpec) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveAPISpecs(c HttpRestSpecPairVisitorContext, left, right *pb.APISpec, cont Cont) Cont {
	return cont
}

// == Method ==================================================================

func (*DefaultHttpRestSpecPairVisitor) EnterMethods(c HttpRestSpecPairVisitorContext, left, right *pb.Method) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitMethodChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Method) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveMethods(c HttpRestSpecPairVisitorContext, left, right *pb.Method, cont Cont) Cont {
	return cont
}

// == MethodMeta ==============================================================

func (*DefaultHttpRestSpecPairVisitor) EnterMethodMetas(c HttpRestSpecPairVisitorContext, left, right *pb.MethodMeta) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitMethodMetaChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.MethodMeta) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveMethodMetas(c HttpRestSpecPairVisitorContext, left, right *pb.MethodMeta, cont Cont) Cont {
	return cont
}

// == HTTPMethodMeta ==========================================================

func (*DefaultHttpRestSpecPairVisitor) EnterHTTPMethodMetas(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPMethodMeta) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitHTTPMethodMetaChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPMethodMeta) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveHTTPMethodMetas(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPMethodMeta, cont Cont) Cont {
	return cont
}

// == Data =====================================================================

func (*DefaultHttpRestSpecPairVisitor) EnterData(c HttpRestSpecPairVisitorContext, left, right *pb.Data) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitDataChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Data) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveData(c HttpRestSpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont {
	return cont
}

// == DataMeta ================================================================

func (*DefaultHttpRestSpecPairVisitor) EnterDataMetas(c HttpRestSpecPairVisitorContext, left, right *pb.DataMeta) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitDataMetaChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.DataMeta) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveDataMetas(c HttpRestSpecPairVisitorContext, left, right *pb.DataMeta, cont Cont) Cont {
	return cont
}

// == HTTPMeta ================================================================

func (*DefaultHttpRestSpecPairVisitor) EnterHTTPMetas(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPMeta) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitHTTPMetaChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPMeta) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveHTTPMetas(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPMeta, cont Cont) Cont {
	return cont
}

// == HTTPPath ================================================================

func (*DefaultHttpRestSpecPairVisitor) EnterHTTPPaths(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPPath) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitHTTPPathChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPPath) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveHTTPPaths(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPPath, cont Cont) Cont {
	return cont
}

// == HTTPQuery ===============================================================

func (*DefaultHttpRestSpecPairVisitor) EnterHTTPQueries(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPQuery) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitHTTPQueryChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPQuery) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveHTTPQueries(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPQuery, cont Cont) Cont {
	return cont
}

// == HTTPHeader ==============================================================

func (*DefaultHttpRestSpecPairVisitor) EnterHTTPHeaders(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPHeader) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitHTTPHeaderChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPHeader) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveHTTPHeaders(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPHeader, cont Cont) Cont {
	return cont
}

// == HTTPCookie ==============================================================

func (*DefaultHttpRestSpecPairVisitor) EnterHTTPCookies(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPCookie) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitHTTPCookieChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPCookie) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveHTTPCookies(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPCookie, cont Cont) Cont {
	return cont
}

// == HTTPBody ================================================================

func (*DefaultHttpRestSpecPairVisitor) EnterHTTPBodies(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPBody) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitHTTPBodyChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPBody) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveHTTPBodies(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPBody, cont Cont) Cont {
	return cont
}

// == HTTPEmpty ===============================================================

func (*DefaultHttpRestSpecPairVisitor) EnterHTTPEmpties(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPEmpty) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitHTTPEmptyChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPEmpty) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveHTTPEmpties(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPEmpty, cont Cont) Cont {
	return cont
}

// == HTTPAuth ================================================================

func (*DefaultHttpRestSpecPairVisitor) EnterHTTPAuths(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitHTTPAuthChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPAuth) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveHTTPAuths(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth, cont Cont) Cont {
	return cont
}

// == HTTPMultipart ===========================================================

func (*DefaultHttpRestSpecPairVisitor) EnterHTTPMultiparts(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPMultipart) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitHTTPMultipartChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPMultipart) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveHTTPMultiparts(c HttpRestSpecPairVisitorContext, left, right *pb.HTTPMultipart, cont Cont) Cont {
	return cont
}

// == Primitive ===============================================================

func (*DefaultHttpRestSpecPairVisitor) EnterPrimitives(c HttpRestSpecPairVisitorContext, left, right *pb.Primitive) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitPrimitiveChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Primitive) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeavePrimitives(c HttpRestSpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont {
	return cont
}

// == Struct ===================================================================

func (*DefaultHttpRestSpecPairVisitor) EnterStructs(c HttpRestSpecPairVisitorContext, left, right *pb.Struct) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitStructChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Struct) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveStructs(c HttpRestSpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont {
	return cont
}

// == List =====================================================================

func (*DefaultHttpRestSpecPairVisitor) EnterLists(c HttpRestSpecPairVisitorContext, left, right *pb.List) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitListChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.List) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveLists(c HttpRestSpecPairVisitorContext, left, right *pb.List, cont Cont) Cont {
	return cont
}

// == Optional =================================================================

func (*DefaultHttpRestSpecPairVisitor) EnterOptionals(c HttpRestSpecPairVisitorContext, left, right *pb.Optional) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitOptionalChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Optional) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveOptionals(c HttpRestSpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont {
	return cont
}

// == OneOf ====================================================================

func (*DefaultHttpRestSpecPairVisitor) EnterOneOfs(c HttpRestSpecPairVisitorContext, left, right *pb.OneOf) Cont {
	return Continue
}

func (*DefaultHttpRestSpecPairVisitor) VisitOneOfChildren(c HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.OneOf) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultHttpRestSpecPairVisitor) LeaveOneOfs(c HttpRestSpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont {
	return cont
}

// == Different types =========================================================

func (*DefaultHttpRestSpecPairVisitor) EnterDifferentTypes(c HttpRestSpecPairVisitorContext, left, right interface{}) Cont {
	return SkipChildren
}

func (*DefaultHttpRestSpecPairVisitor) LeaveDifferentTypes(c HttpRestSpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	return cont
}

// extendContext implementation for HttpRestSpecPairVisitor.
func extendPairContext(cin PairContext, left, right interface{}) PairContext {
	ctx, ok := cin.(HttpRestSpecPairVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.extendPairContext expected HttpRestSpecPairVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}

	ctx.ExtendLeftContext(left)
	ctx.ExtendRightContext(right)
	return ctx
}

// enter implementation for HttpRestSpecVisitor.
func enterPair(cin PairContext, visitor interface{}, left, right interface{}) Cont {
	v, _ := visitor.(HttpRestSpecPairVisitor)
	ctx, ok := extendPairContext(cin, left, right).(HttpRestSpecPairVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.enterPair expected HttpRestSpecPairVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}
	keepGoing := Continue

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return v.EnterDifferentTypes(ctx, left, right)
	}

	// Dispatch on type and path.
	switch leftNode := left.(type) {
	case pb.APISpec, pb.Method, pb.MethodMeta, pb.HTTPMethodMeta, pb.Data, pb.DataMeta, pb.HTTPMeta, pb.HTTPPath, pb.HTTPQuery, pb.HTTPHeader, pb.HTTPCookie, pb.HTTPBody, pb.HTTPAuth, pb.HTTPMultipart, pb.Primitive, pb.Struct, pb.List, pb.Optional, pb.OneOf:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		return enterPair(ctx, v, &leftNode, &right)

	case *pb.APISpec:
		rightNode := right.(*pb.APISpec)
		return v.EnterAPISpecs(ctx, leftNode, rightNode)

	case *pb.Method:
		rightNode := right.(*pb.Method)
		return v.EnterMethods(ctx, leftNode, rightNode)

	case *pb.MethodMeta:
		rightNode := right.(*pb.MethodMeta)
		return v.EnterMethodMetas(ctx, leftNode, rightNode)

	case *pb.HTTPMethodMeta:
		rightNode := right.(*pb.HTTPMethodMeta)
		return v.EnterHTTPMethodMetas(ctx, leftNode, rightNode)

	case *pb.Data:
		rightNode := right.(*pb.Data)
		return v.EnterData(ctx, leftNode, rightNode)

	case *pb.DataMeta:
		rightNode := right.(*pb.DataMeta)
		return v.EnterDataMetas(ctx, leftNode, rightNode)

	case *pb.HTTPPath:
		rightNode := right.(*pb.HTTPPath)
		return v.EnterHTTPPaths(ctx, leftNode, rightNode)

	case *pb.HTTPQuery:
		rightNode := right.(*pb.HTTPQuery)
		return v.EnterHTTPQueries(ctx, leftNode, rightNode)

	case *pb.HTTPHeader:
		rightNode := right.(*pb.HTTPHeader)
		return v.EnterHTTPHeaders(ctx, leftNode, rightNode)

	case *pb.HTTPCookie:
		rightNode := right.(*pb.HTTPCookie)
		return v.EnterHTTPCookies(ctx, leftNode, rightNode)

	case *pb.HTTPBody:
		rightNode := right.(*pb.HTTPBody)
		return v.EnterHTTPBodies(ctx, leftNode, rightNode)

	case *pb.HTTPEmpty:
		rightNode := right.(*pb.HTTPEmpty)
		return v.EnterHTTPEmpties(ctx, leftNode, rightNode)

	case *pb.HTTPAuth:
		rightNode := right.(*pb.HTTPAuth)
		return v.EnterHTTPAuths(ctx, leftNode, rightNode)

	case *pb.HTTPMultipart:
		rightNode := right.(*pb.HTTPMultipart)
		return v.EnterHTTPMultiparts(ctx, leftNode, rightNode)

	case *pb.Primitive:
		rightNode := right.(*pb.Primitive)
		return v.EnterPrimitives(ctx, leftNode, rightNode)

	case *pb.Struct:
		rightNode := right.(*pb.Struct)
		return v.EnterStructs(ctx, leftNode, rightNode)

	case *pb.List:
		rightNode := right.(*pb.List)
		return v.EnterLists(ctx, leftNode, rightNode)

	case *pb.Optional:
		rightNode := right.(*pb.Optional)
		return v.EnterOptionals(ctx, leftNode, rightNode)

	case *pb.OneOf:
		rightNode := right.(*pb.OneOf)
		return v.EnterOneOfs(ctx, leftNode, rightNode)
	}

	// Didn't understand the type. Just keep going.
	return keepGoing
}

// visitChildren implementation for HttpRestSpecPairVisitor.
func visitPairChildren(cin PairContext, vm PairVisitorManager, left, right interface{}) Cont {
	visitor := vm.Visitor()
	v, _ := visitor.(HttpRestSpecPairVisitor)
	ctx, ok := cin.(HttpRestSpecPairVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.visitPairChildren expected HttpRestSpecPairVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}

	// Expect left and right to be the same type.
	assertSameType(left, right)

	// Dispatch on type and path.
	switch leftNode := left.(type) {
	case pb.APISpec, pb.Method, pb.MethodMeta, pb.HTTPMethodMeta, pb.Data, pb.DataMeta, pb.HTTPMeta, pb.HTTPPath, pb.HTTPQuery, pb.HTTPHeader, pb.HTTPCookie, pb.HTTPBody, pb.HTTPAuth, pb.HTTPMultipart, pb.Primitive, pb.Struct, pb.List, pb.Optional, pb.OneOf:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		return visitPairChildren(ctx, vm, &left, &right)

	case *pb.APISpec:
		rightNode := right.(*pb.APISpec)
		return v.VisitAPISpecChildren(ctx, vm, leftNode, rightNode)

	case *pb.Method:
		rightNode := right.(*pb.Method)
		return v.VisitMethodChildren(ctx, vm, leftNode, rightNode)

	case *pb.MethodMeta:
		rightNode := right.(*pb.MethodMeta)
		return v.VisitMethodMetaChildren(ctx, vm, leftNode, rightNode)

	case *pb.HTTPMethodMeta:
		rightNode := right.(*pb.HTTPMethodMeta)
		return v.VisitHTTPMethodMetaChildren(ctx, vm, leftNode, rightNode)

	case *pb.Data:
		rightNode := right.(*pb.Data)
		return v.VisitDataChildren(ctx, vm, leftNode, rightNode)

	case *pb.DataMeta:
		rightNode := right.(*pb.DataMeta)
		return v.VisitDataMetaChildren(ctx, vm, leftNode, rightNode)

	case *pb.HTTPPath:
		rightNode := right.(*pb.HTTPPath)
		return v.VisitHTTPPathChildren(ctx, vm, leftNode, rightNode)

	case *pb.HTTPQuery:
		rightNode := right.(*pb.HTTPQuery)
		return v.VisitHTTPQueryChildren(ctx, vm, leftNode, rightNode)

	case *pb.HTTPHeader:
		rightNode := right.(*pb.HTTPHeader)
		return v.VisitHTTPHeaderChildren(ctx, vm, leftNode, rightNode)

	case *pb.HTTPCookie:
		rightNode := right.(*pb.HTTPCookie)
		return v.VisitHTTPCookieChildren(ctx, vm, leftNode, rightNode)

	case *pb.HTTPBody:
		rightNode := right.(*pb.HTTPBody)
		return v.VisitHTTPBodyChildren(ctx, vm, leftNode, rightNode)

	case *pb.HTTPEmpty:
		rightNode := right.(*pb.HTTPEmpty)
		return v.VisitHTTPEmptyChildren(ctx, vm, leftNode, rightNode)

	case *pb.HTTPAuth:
		rightNode := right.(*pb.HTTPAuth)
		return v.VisitHTTPAuthChildren(ctx, vm, leftNode, rightNode)

	case *pb.HTTPMultipart:
		rightNode := right.(*pb.HTTPMultipart)
		return v.VisitHTTPMultipartChildren(ctx, vm, leftNode, rightNode)

	case *pb.Primitive:
		rightNode := right.(*pb.Primitive)
		return v.VisitPrimitiveChildren(ctx, vm, leftNode, rightNode)

	case *pb.Struct:
		rightNode := right.(*pb.Struct)
		return v.VisitStructChildren(ctx, vm, leftNode, rightNode)

	case *pb.List:
		rightNode := right.(*pb.List)
		return v.VisitListChildren(ctx, vm, leftNode, rightNode)

	case *pb.Optional:
		rightNode := right.(*pb.Optional)
		return v.VisitOptionalChildren(ctx, vm, leftNode, rightNode)

	case *pb.OneOf:
		rightNode := right.(*pb.OneOf)
		return v.VisitOneOfChildren(ctx, vm, leftNode, rightNode)

	default:
		return v.DefaultVisitChildren(ctx, vm, left, right)
	}
}

// leave implementation for HttpRestSpecPairVisitor.
func leavePair(cin PairContext, visitor interface{}, left, right interface{}, cont Cont) Cont {
	v, _ := visitor.(HttpRestSpecPairVisitor)
	ctx, ok := extendPairContext(cin, left, right).(HttpRestSpecPairVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.leave expected HttpRestSpecPairVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}
	keepGoing := cont

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return v.LeaveDifferentTypes(ctx, left, right, cont)
	}

	// Dispatch on type and path.
	switch leftNode := left.(type) {
	case pb.APISpec, pb.Method, pb.MethodMeta, pb.HTTPMethodMeta, pb.Data, pb.DataMeta, pb.HTTPMeta, pb.HTTPPath, pb.HTTPQuery, pb.HTTPHeader, pb.HTTPCookie, pb.HTTPBody, pb.HTTPAuth, pb.HTTPMultipart, pb.Primitive, pb.Struct, pb.List, pb.Optional, pb.OneOf:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		return leavePair(ctx, v, &left, &right, cont)

	case *pb.APISpec:
		rightNode := right.(*pb.APISpec)
		return v.LeaveAPISpecs(ctx, leftNode, rightNode, cont)

	case *pb.Method:
		rightNode := right.(*pb.Method)
		return v.LeaveMethods(ctx, leftNode, rightNode, cont)

	case *pb.MethodMeta:
		rightNode := right.(*pb.MethodMeta)
		return v.LeaveMethodMetas(ctx, leftNode, rightNode, cont)

	case *pb.HTTPMethodMeta:
		rightNode := right.(*pb.HTTPMethodMeta)
		return v.LeaveHTTPMethodMetas(ctx, leftNode, rightNode, cont)

	case *pb.Data:
		rightNode := right.(*pb.Data)
		return v.LeaveData(ctx, leftNode, rightNode, cont)

	case *pb.DataMeta:
		rightNode := right.(*pb.DataMeta)
		return v.LeaveDataMetas(ctx, leftNode, rightNode, cont)

	case *pb.HTTPPath:
		rightNode := right.(*pb.HTTPPath)
		return v.LeaveHTTPPaths(ctx, leftNode, rightNode, cont)

	case *pb.HTTPQuery:
		rightNode := right.(*pb.HTTPQuery)
		return v.LeaveHTTPQueries(ctx, leftNode, rightNode, cont)

	case *pb.HTTPHeader:
		rightNode := right.(*pb.HTTPHeader)
		return v.LeaveHTTPHeaders(ctx, leftNode, rightNode, cont)

	case *pb.HTTPCookie:
		rightNode := right.(*pb.HTTPCookie)
		return v.LeaveHTTPCookies(ctx, leftNode, rightNode, cont)

	case *pb.HTTPBody:
		rightNode := right.(*pb.HTTPBody)
		return v.LeaveHTTPBodies(ctx, leftNode, rightNode, cont)

	case *pb.HTTPEmpty:
		rightNode := right.(*pb.HTTPEmpty)
		return v.LeaveHTTPEmpties(ctx, leftNode, rightNode, cont)

	case *pb.HTTPAuth:
		rightNode := right.(*pb.HTTPAuth)
		return v.LeaveHTTPAuths(ctx, leftNode, rightNode, cont)

	case *pb.HTTPMultipart:
		rightNode := right.(*pb.HTTPMultipart)
		return v.LeaveHTTPMultiparts(ctx, leftNode, rightNode, cont)

	case *pb.Primitive:
		rightNode := right.(*pb.Primitive)
		return v.LeavePrimitives(ctx, leftNode, rightNode, cont)

	case *pb.Struct:
		rightNode := right.(*pb.Struct)
		return v.LeaveStructs(ctx, leftNode, rightNode, cont)

	case *pb.List:
		rightNode := right.(*pb.List)
		return v.LeaveLists(ctx, leftNode, rightNode, cont)

	case *pb.Optional:
		rightNode := right.(*pb.Optional)
		return v.LeaveOptionals(ctx, leftNode, rightNode, cont)

	case *pb.OneOf:
		rightNode := right.(*pb.OneOf)
		return v.LeaveOneOfs(ctx, leftNode, rightNode, cont)
	}

	// Didn't understand the type. Just keep going.
	return keepGoing
}

// Visits left and right with v in tandem.
func ApplyPair(v HttpRestSpecPairVisitor, left, right interface{}) Cont {
	c := newHttpRestSpecPairVisitorContext()
	vis := NewPairVisitorManager(c, v, enterPair, visitPairChildren, leavePair, extendPairContext)
	return go_ast_pair.Apply(vis, left, right)
}

// Panics if the two arguments have different types.
func assertSameType(x, y interface{}) {
	xt := reflect.TypeOf(x)
	yt := reflect.TypeOf(y)
	if xt != yt {
		callerName := ""
		pc, _, _, ok := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		if ok && details != nil {
			callerName = fmt.Sprintf("%s ", details.Name())
		}
		panic(fmt.Sprintf("%sexpected nodes of the same type, but got %s and %s", callerName, xt, yt))
	}
}
