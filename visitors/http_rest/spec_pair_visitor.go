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
//
// Go lacks virtual functions, so all functions here take the visitor itself as
// an argument, and call functions on that instance.
type SpecPairVisitor interface {
	EnterAPISpecs(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.APISpec) Cont
	VisitAPISpecChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.APISpec) Cont
	LeaveAPISpecs(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.APISpec, cont Cont) Cont

	EnterMethods(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.Method) Cont
	VisitMethodChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Method) Cont
	LeaveMethods(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.Method, cont Cont) Cont

	EnterMethodMetas(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.MethodMeta) Cont
	VisitMethodMetaChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.MethodMeta) Cont
	LeaveMethodMetas(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.MethodMeta, cont Cont) Cont

	EnterHTTPMethodMetas(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPMethodMeta) Cont
	VisitHTTPMethodMetaChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPMethodMeta) Cont
	LeaveHTTPMethodMetas(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPMethodMeta, cont Cont) Cont

	EnterData(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.Data) Cont
	VisitDataChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Data) Cont
	LeaveData(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont

	EnterDataMetas(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.DataMeta) Cont
	VisitDataMetaChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.DataMeta) Cont
	LeaveDataMetas(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.DataMeta, cont Cont) Cont

	EnterHTTPMetas(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPMeta) Cont
	VisitHTTPMetaChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPMeta) Cont
	LeaveHTTPMetas(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPMeta, cont Cont) Cont

	EnterHTTPPaths(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPPath) Cont
	VisitHTTPPathChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPPath) Cont
	LeaveHTTPPaths(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPPath, cont Cont) Cont

	EnterHTTPQueries(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPQuery) Cont
	VisitHTTPQueryChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPQuery) Cont
	LeaveHTTPQueries(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPQuery, cont Cont) Cont

	EnterHTTPHeaders(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPHeader) Cont
	VisitHTTPHeaderChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPHeader) Cont
	LeaveHTTPHeaders(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPHeader, cont Cont) Cont

	EnterHTTPCookies(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPCookie) Cont
	VisitHTTPCookieChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPCookie) Cont
	LeaveHTTPCookies(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPCookie, cont Cont) Cont

	EnterHTTPBodies(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPBody) Cont
	VisitHTTPBodyChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPBody) Cont
	LeaveHTTPBodies(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPBody, cont Cont) Cont

	EnterHTTPEmpties(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPEmpty) Cont
	VisitHTTPEmptyChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPEmpty) Cont
	LeaveHTTPEmpties(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPEmpty, cont Cont) Cont

	EnterHTTPAuths(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPAuth) Cont
	VisitHTTPAuthChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPAuth) Cont
	LeaveHTTPAuths(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPAuth, cont Cont) Cont

	EnterHTTPMultiparts(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPMultipart) Cont
	VisitHTTPMultipartChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPMultipart) Cont
	LeaveHTTPMultiparts(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.HTTPMultipart, cont Cont) Cont

	EnterPrimitives(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.Primitive) Cont
	VisitPrimitiveChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Primitive) Cont
	LeavePrimitives(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont

	EnterStructs(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.Struct) Cont
	VisitStructChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Struct) Cont
	LeaveStructs(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont

	EnterLists(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.List) Cont
	VisitListChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.List) Cont
	LeaveLists(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.List, cont Cont) Cont

	EnterOptionals(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.Optional) Cont
	VisitOptionalChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Optional) Cont
	LeaveOptionals(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont

	EnterOneOfs(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.OneOf) Cont
	VisitOneOfChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.OneOf) Cont
	LeaveOneOfs(self interface{}, ctxt SpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont

	DefaultVisitChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right interface{}) Cont

	// Used when the visitor tries to enter two nodes with different types. This
	// cannot return Continue; otherwise, visitChildren will panic.
	EnterDifferentTypes(self interface{}, ctxt SpecPairVisitorContext, left, right interface{}) Cont

	// Used when the visitor tries to leave two nodes with different types.
	LeaveDifferentTypes(self interface{}, ctxt SpecPairVisitorContext, left, right interface{}, cont Cont) Cont
}

type DefaultSpecPairVisitor struct{}

func (*DefaultSpecPairVisitor) DefaultVisitChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right interface{}) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

// == APISpec =================================================================

func (*DefaultSpecPairVisitor) EnterAPISpecs(self interface{}, c SpecPairVisitorContext, left, right *pb.APISpec) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitAPISpecChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.APISpec) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveAPISpecs(self interface{}, c SpecPairVisitorContext, left, right *pb.APISpec, cont Cont) Cont {
	return cont
}

// == Method ==================================================================

func (*DefaultSpecPairVisitor) EnterMethods(self interface{}, c SpecPairVisitorContext, left, right *pb.Method) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitMethodChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Method) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveMethods(self interface{}, c SpecPairVisitorContext, left, right *pb.Method, cont Cont) Cont {
	return cont
}

// == MethodMeta ==============================================================

func (*DefaultSpecPairVisitor) EnterMethodMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.MethodMeta) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitMethodMetaChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.MethodMeta) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveMethodMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.MethodMeta, cont Cont) Cont {
	return cont
}

// == HTTPMethodMeta ==========================================================

func (*DefaultSpecPairVisitor) EnterHTTPMethodMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPMethodMeta) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitHTTPMethodMetaChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPMethodMeta) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveHTTPMethodMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPMethodMeta, cont Cont) Cont {
	return cont
}

// == Data =====================================================================

func (*DefaultSpecPairVisitor) EnterData(self interface{}, c SpecPairVisitorContext, left, right *pb.Data) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitDataChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Data) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveData(self interface{}, c SpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont {
	return cont
}

// == DataMeta ================================================================

func (*DefaultSpecPairVisitor) EnterDataMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.DataMeta) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitDataMetaChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.DataMeta) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveDataMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.DataMeta, cont Cont) Cont {
	return cont
}

// == HTTPMeta ================================================================

func (*DefaultSpecPairVisitor) EnterHTTPMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPMeta) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitHTTPMetaChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPMeta) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveHTTPMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPMeta, cont Cont) Cont {
	return cont
}

// == HTTPPath ================================================================

func (*DefaultSpecPairVisitor) EnterHTTPPaths(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPPath) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitHTTPPathChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPPath) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveHTTPPaths(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPPath, cont Cont) Cont {
	return cont
}

// == HTTPQuery ===============================================================

func (*DefaultSpecPairVisitor) EnterHTTPQueries(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPQuery) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitHTTPQueryChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPQuery) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveHTTPQueries(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPQuery, cont Cont) Cont {
	return cont
}

// == HTTPHeader ==============================================================

func (*DefaultSpecPairVisitor) EnterHTTPHeaders(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPHeader) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitHTTPHeaderChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPHeader) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveHTTPHeaders(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPHeader, cont Cont) Cont {
	return cont
}

// == HTTPCookie ==============================================================

func (*DefaultSpecPairVisitor) EnterHTTPCookies(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPCookie) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitHTTPCookieChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPCookie) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveHTTPCookies(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPCookie, cont Cont) Cont {
	return cont
}

// == HTTPBody ================================================================

func (*DefaultSpecPairVisitor) EnterHTTPBodies(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPBody) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitHTTPBodyChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPBody) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveHTTPBodies(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPBody, cont Cont) Cont {
	return cont
}

// == HTTPEmpty ===============================================================

func (*DefaultSpecPairVisitor) EnterHTTPEmpties(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPEmpty) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitHTTPEmptyChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPEmpty) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveHTTPEmpties(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPEmpty, cont Cont) Cont {
	return cont
}

// == HTTPAuth ================================================================

func (*DefaultSpecPairVisitor) EnterHTTPAuths(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPAuth) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitHTTPAuthChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPAuth) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveHTTPAuths(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPAuth, cont Cont) Cont {
	return cont
}

// == HTTPMultipart ===========================================================

func (*DefaultSpecPairVisitor) EnterHTTPMultiparts(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPMultipart) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitHTTPMultipartChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPMultipart) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveHTTPMultiparts(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPMultipart, cont Cont) Cont {
	return cont
}

// == Primitive ===============================================================

func (*DefaultSpecPairVisitor) EnterPrimitives(self interface{}, c SpecPairVisitorContext, left, right *pb.Primitive) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitPrimitiveChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Primitive) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeavePrimitives(self interface{}, c SpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont {
	return cont
}

// == Struct ==================================================================

func (*DefaultSpecPairVisitor) EnterStructs(self interface{}, c SpecPairVisitorContext, left, right *pb.Struct) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitStructChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Struct) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveStructs(self interface{}, c SpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont {
	return cont
}

// == List ====================================================================

func (*DefaultSpecPairVisitor) EnterLists(self interface{}, c SpecPairVisitorContext, left, right *pb.List) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitListChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.List) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveLists(self interface{}, c SpecPairVisitorContext, left, right *pb.List, cont Cont) Cont {
	return cont
}

// == Optional ================================================================

func (*DefaultSpecPairVisitor) EnterOptionals(self interface{}, c SpecPairVisitorContext, left, right *pb.Optional) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitOptionalChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Optional) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveOptionals(self interface{}, c SpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont {
	return cont
}

// == OneOf ===================================================================

func (*DefaultSpecPairVisitor) EnterOneOfs(self interface{}, c SpecPairVisitorContext, left, right *pb.OneOf) Cont {
	return Continue
}

func (*DefaultSpecPairVisitor) VisitOneOfChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.OneOf) Cont {
	return go_ast_pair.DefaultVisitChildren(c, vm, left, right)
}

func (*DefaultSpecPairVisitor) LeaveOneOfs(self interface{}, c SpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont {
	return cont
}

// == Different types =========================================================

func (*DefaultSpecPairVisitor) EnterDifferentTypes(self interface{}, c SpecPairVisitorContext, left, right interface{}) Cont {
	return SkipChildren
}

func (*DefaultSpecPairVisitor) LeaveDifferentTypes(self interface{}, c SpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	return cont
}

// extendContext implementation for SpecPairVisitor.
func extendPairContext(cin PairContext, left, right interface{}) PairContext {
	ctx, ok := cin.(SpecPairVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.extendPairContext expected SpecPairVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}

	ctx.ExtendLeftContext(left)
	ctx.ExtendRightContext(right)
	return ctx
}

// enter implementation for SpecVisitor.
func enterPair(cin PairContext, visitor interface{}, left, right interface{}) Cont {
	v, _ := visitor.(SpecPairVisitor)
	ctx, ok := extendPairContext(cin, left, right).(SpecPairVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.enterPair expected SpecPairVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}
	keepGoing := Continue

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return v.EnterDifferentTypes(v, ctx, left, right)
	}

	// Dispatch on type and path.
	switch leftNode := left.(type) {
	case pb.APISpec, pb.Method, pb.MethodMeta, pb.HTTPMethodMeta, pb.Data, pb.DataMeta, pb.HTTPMeta, pb.HTTPPath, pb.HTTPQuery, pb.HTTPHeader, pb.HTTPCookie, pb.HTTPBody, pb.HTTPAuth, pb.HTTPMultipart, pb.Primitive, pb.Struct, pb.List, pb.Optional, pb.OneOf:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		return enterPair(ctx, visitor, &leftNode, &right)

	case *pb.APISpec:
		rightNode := right.(*pb.APISpec)
		return v.EnterAPISpecs(visitor, ctx, leftNode, rightNode)

	case *pb.Method:
		rightNode := right.(*pb.Method)
		return v.EnterMethods(visitor, ctx, leftNode, rightNode)

	case *pb.MethodMeta:
		rightNode := right.(*pb.MethodMeta)
		return v.EnterMethodMetas(visitor, ctx, leftNode, rightNode)

	case *pb.HTTPMethodMeta:
		rightNode := right.(*pb.HTTPMethodMeta)
		return v.EnterHTTPMethodMetas(visitor, ctx, leftNode, rightNode)

	case *pb.Data:
		rightNode := right.(*pb.Data)
		return v.EnterData(visitor, ctx, leftNode, rightNode)

	case *pb.DataMeta:
		rightNode := right.(*pb.DataMeta)
		return v.EnterDataMetas(visitor, ctx, leftNode, rightNode)

	case *pb.HTTPPath:
		rightNode := right.(*pb.HTTPPath)
		return v.EnterHTTPPaths(visitor, ctx, leftNode, rightNode)

	case *pb.HTTPQuery:
		rightNode := right.(*pb.HTTPQuery)
		return v.EnterHTTPQueries(visitor, ctx, leftNode, rightNode)

	case *pb.HTTPHeader:
		rightNode := right.(*pb.HTTPHeader)
		return v.EnterHTTPHeaders(visitor, ctx, leftNode, rightNode)

	case *pb.HTTPCookie:
		rightNode := right.(*pb.HTTPCookie)
		return v.EnterHTTPCookies(visitor, ctx, leftNode, rightNode)

	case *pb.HTTPBody:
		rightNode := right.(*pb.HTTPBody)
		return v.EnterHTTPBodies(visitor, ctx, leftNode, rightNode)

	case *pb.HTTPEmpty:
		rightNode := right.(*pb.HTTPEmpty)
		return v.EnterHTTPEmpties(visitor, ctx, leftNode, rightNode)

	case *pb.HTTPAuth:
		rightNode := right.(*pb.HTTPAuth)
		return v.EnterHTTPAuths(visitor, ctx, leftNode, rightNode)

	case *pb.HTTPMultipart:
		rightNode := right.(*pb.HTTPMultipart)
		return v.EnterHTTPMultiparts(visitor, ctx, leftNode, rightNode)

	case *pb.Primitive:
		rightNode := right.(*pb.Primitive)
		return v.EnterPrimitives(visitor, ctx, leftNode, rightNode)

	case *pb.Struct:
		rightNode := right.(*pb.Struct)
		return v.EnterStructs(visitor, ctx, leftNode, rightNode)

	case *pb.List:
		rightNode := right.(*pb.List)
		return v.EnterLists(visitor, ctx, leftNode, rightNode)

	case *pb.Optional:
		rightNode := right.(*pb.Optional)
		return v.EnterOptionals(visitor, ctx, leftNode, rightNode)

	case *pb.OneOf:
		rightNode := right.(*pb.OneOf)
		return v.EnterOneOfs(visitor, ctx, leftNode, rightNode)
	}

	// Didn't understand the type. Just keep going.
	return keepGoing
}

// visitChildren implementation for SpecPairVisitor.
func visitPairChildren(cin PairContext, vm PairVisitorManager, left, right interface{}) Cont {
	visitor := vm.Visitor()
	v, _ := visitor.(SpecPairVisitor)
	ctx, ok := cin.(SpecPairVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.visitPairChildren expected SpecPairVisitorContext, got %s",
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
		return v.VisitAPISpecChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.Method:
		rightNode := right.(*pb.Method)
		return v.VisitMethodChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.MethodMeta:
		rightNode := right.(*pb.MethodMeta)
		return v.VisitMethodMetaChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.HTTPMethodMeta:
		rightNode := right.(*pb.HTTPMethodMeta)
		return v.VisitHTTPMethodMetaChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.Data:
		rightNode := right.(*pb.Data)
		return v.VisitDataChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.DataMeta:
		rightNode := right.(*pb.DataMeta)
		return v.VisitDataMetaChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.HTTPPath:
		rightNode := right.(*pb.HTTPPath)
		return v.VisitHTTPPathChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.HTTPQuery:
		rightNode := right.(*pb.HTTPQuery)
		return v.VisitHTTPQueryChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.HTTPHeader:
		rightNode := right.(*pb.HTTPHeader)
		return v.VisitHTTPHeaderChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.HTTPCookie:
		rightNode := right.(*pb.HTTPCookie)
		return v.VisitHTTPCookieChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.HTTPBody:
		rightNode := right.(*pb.HTTPBody)
		return v.VisitHTTPBodyChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.HTTPEmpty:
		rightNode := right.(*pb.HTTPEmpty)
		return v.VisitHTTPEmptyChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.HTTPAuth:
		rightNode := right.(*pb.HTTPAuth)
		return v.VisitHTTPAuthChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.HTTPMultipart:
		rightNode := right.(*pb.HTTPMultipart)
		return v.VisitHTTPMultipartChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.Primitive:
		rightNode := right.(*pb.Primitive)
		return v.VisitPrimitiveChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.Struct:
		rightNode := right.(*pb.Struct)
		return v.VisitStructChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.List:
		rightNode := right.(*pb.List)
		return v.VisitListChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.Optional:
		rightNode := right.(*pb.Optional)
		return v.VisitOptionalChildren(visitor, ctx, vm, leftNode, rightNode)

	case *pb.OneOf:
		rightNode := right.(*pb.OneOf)
		return v.VisitOneOfChildren(visitor, ctx, vm, leftNode, rightNode)

	default:
		return v.DefaultVisitChildren(visitor, ctx, vm, left, right)
	}
}

// leave implementation for SpecPairVisitor.
func leavePair(cin PairContext, visitor interface{}, left, right interface{}, cont Cont) Cont {
	v, _ := visitor.(SpecPairVisitor)
	ctx, ok := extendPairContext(cin, left, right).(SpecPairVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.leave expected SpecPairVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}
	keepGoing := cont

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return v.LeaveDifferentTypes(visitor, ctx, left, right, cont)
	}

	// Dispatch on type and path.
	switch leftNode := left.(type) {
	case pb.APISpec, pb.Method, pb.MethodMeta, pb.HTTPMethodMeta, pb.Data, pb.DataMeta, pb.HTTPMeta, pb.HTTPPath, pb.HTTPQuery, pb.HTTPHeader, pb.HTTPCookie, pb.HTTPBody, pb.HTTPAuth, pb.HTTPMultipart, pb.Primitive, pb.Struct, pb.List, pb.Optional, pb.OneOf:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		return leavePair(ctx, visitor, &left, &right, cont)

	case *pb.APISpec:
		rightNode := right.(*pb.APISpec)
		return v.LeaveAPISpecs(visitor, ctx, leftNode, rightNode, cont)

	case *pb.Method:
		rightNode := right.(*pb.Method)
		return v.LeaveMethods(visitor, ctx, leftNode, rightNode, cont)

	case *pb.MethodMeta:
		rightNode := right.(*pb.MethodMeta)
		return v.LeaveMethodMetas(visitor, ctx, leftNode, rightNode, cont)

	case *pb.HTTPMethodMeta:
		rightNode := right.(*pb.HTTPMethodMeta)
		return v.LeaveHTTPMethodMetas(visitor, ctx, leftNode, rightNode, cont)

	case *pb.Data:
		rightNode := right.(*pb.Data)
		return v.LeaveData(visitor, ctx, leftNode, rightNode, cont)

	case *pb.DataMeta:
		rightNode := right.(*pb.DataMeta)
		return v.LeaveDataMetas(visitor, ctx, leftNode, rightNode, cont)

	case *pb.HTTPPath:
		rightNode := right.(*pb.HTTPPath)
		return v.LeaveHTTPPaths(visitor, ctx, leftNode, rightNode, cont)

	case *pb.HTTPQuery:
		rightNode := right.(*pb.HTTPQuery)
		return v.LeaveHTTPQueries(visitor, ctx, leftNode, rightNode, cont)

	case *pb.HTTPHeader:
		rightNode := right.(*pb.HTTPHeader)
		return v.LeaveHTTPHeaders(visitor, ctx, leftNode, rightNode, cont)

	case *pb.HTTPCookie:
		rightNode := right.(*pb.HTTPCookie)
		return v.LeaveHTTPCookies(visitor, ctx, leftNode, rightNode, cont)

	case *pb.HTTPBody:
		rightNode := right.(*pb.HTTPBody)
		return v.LeaveHTTPBodies(visitor, ctx, leftNode, rightNode, cont)

	case *pb.HTTPEmpty:
		rightNode := right.(*pb.HTTPEmpty)
		return v.LeaveHTTPEmpties(visitor, ctx, leftNode, rightNode, cont)

	case *pb.HTTPAuth:
		rightNode := right.(*pb.HTTPAuth)
		return v.LeaveHTTPAuths(visitor, ctx, leftNode, rightNode, cont)

	case *pb.HTTPMultipart:
		rightNode := right.(*pb.HTTPMultipart)
		return v.LeaveHTTPMultiparts(visitor, ctx, leftNode, rightNode, cont)

	case *pb.Primitive:
		rightNode := right.(*pb.Primitive)
		return v.LeavePrimitives(visitor, ctx, leftNode, rightNode, cont)

	case *pb.Struct:
		rightNode := right.(*pb.Struct)
		return v.LeaveStructs(visitor, ctx, leftNode, rightNode, cont)

	case *pb.List:
		rightNode := right.(*pb.List)
		return v.LeaveLists(visitor, ctx, leftNode, rightNode, cont)

	case *pb.Optional:
		rightNode := right.(*pb.Optional)
		return v.LeaveOptionals(visitor, ctx, leftNode, rightNode, cont)

	case *pb.OneOf:
		rightNode := right.(*pb.OneOf)
		return v.LeaveOneOfs(visitor, ctx, leftNode, rightNode, cont)
	}

	// Didn't understand the type. Just keep going.
	return keepGoing
}

// Visits left and right with v in tandem.
func ApplyPair(v SpecPairVisitor, left, right interface{}) Cont {
	c := newSpecPairVisitorContext()
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
