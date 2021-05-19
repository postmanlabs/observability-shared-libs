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

	// A utility function for visiting a set of arguments in a method request or
	// response. This is needed since the arguments are store in maps whose keys
	// do not reliably identify the arguments.
	VisitMethodArgs(self interface{}, ctxt PairContext, vm PairVisitorManager, left, right map[string]*pb.Data) Cont

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

	// Visits the children of an unknown node type.
	DefaultVisitChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right interface{}) Cont

	// Used when the visitor tries to enter two nodes with different types. This
	// cannot return Continue; otherwise, visitChildren will panic.
	EnterDifferentTypes(self interface{}, ctxt SpecPairVisitorContext, left, right interface{}) Cont

	// Used when the visitor tries to leave two nodes with different types.
	LeaveDifferentTypes(self interface{}, ctxt SpecPairVisitorContext, left, right interface{}, cont Cont) Cont
}

// A SpecPairVisitor with methods for providing default visiting behaviour.
type DefaultSpecPairVisitor interface {
	SpecPairVisitor

	EnterNodes(self interface{}, ctxt SpecPairVisitorContext, left, right interface{}) Cont

	VisitNodeChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right interface{}) Cont

	LeaveNodes(self interface{}, ctxt SpecPairVisitorContext, left, right interface{}, cont Cont) Cont
}

type DefaultSpecPairVisitorImpl struct{}

var _ DefaultSpecPairVisitor = (*DefaultSpecPairVisitorImpl)(nil)

func (*DefaultSpecPairVisitorImpl) EnterNodes(self interface{}, ctxt SpecPairVisitorContext, left, right interface{}) Cont {
	return Continue
}

func (*DefaultSpecPairVisitorImpl) VisitNodeChildren(self interface{}, ctxt SpecPairVisitorContext, vm PairVisitorManager, left, right interface{}) Cont {
	return go_ast_pair.DefaultVisitChildren(ctxt, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveNodes(self interface{}, ctxt SpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	return cont
}

func (*DefaultSpecPairVisitorImpl) DefaultVisitChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right interface{}) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

// == APISpec =================================================================

func (*DefaultSpecPairVisitorImpl) EnterAPISpecs(self interface{}, c SpecPairVisitorContext, left, right *pb.APISpec) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitAPISpecChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.APISpec) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveAPISpecs(self interface{}, c SpecPairVisitorContext, left, right *pb.APISpec, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == Method ==================================================================

func (*DefaultSpecPairVisitorImpl) EnterMethods(self interface{}, c SpecPairVisitorContext, left, right *pb.Method) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

// Need to do special casing here, since the keys used in the method's Args
// and Responses do not reliably identify an arg or response.
func (*DefaultSpecPairVisitorImpl) VisitMethodChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Method) Cont {
	v := self.(DefaultSpecPairVisitor)

	if left == nil || right == nil {
		return Continue
	}

	keepGoing := go_ast_pair.ApplyWithContext(vm, c.EnterStructs(left, "Id", right, "Id"), left.Id, right.Id)
	if result := handleKeepGoing(keepGoing); result != nil {
		return *result
	}

	keepGoing = v.VisitMethodArgs(self, c.EnterStructs(left, "Args", right, "Args"), vm, left.Args, right.Args)
	if result := handleKeepGoing(keepGoing); result != nil {
		return *result
	}

	keepGoing = v.VisitMethodArgs(self, c.EnterStructs(left, "Responses", right, "Responses"), vm, left.Responses, right.Responses)
	if result := handleKeepGoing(keepGoing); result != nil {
		return *result
	}

	keepGoing = go_ast_pair.ApplyWithContext(vm, c.EnterStructs(left, "Meta", right, "Meta"), left.Meta, right.Meta)
	if result := handleKeepGoing(keepGoing); result != nil {
		return *result
	}

	return keepGoing
}

func handleKeepGoing(keepGoing Cont) *Cont {
	switch keepGoing {
	case Abort, Stop:
		return &keepGoing
	case Continue:
		return nil
	case SkipChildren:
		panic("Unexpected SkipChildren")
	default:
		panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
	}
}

func (*DefaultSpecPairVisitorImpl) LeaveMethods(self interface{}, c SpecPairVisitorContext, left, right *pb.Method, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

func (*DefaultSpecPairVisitorImpl) VisitMethodArgs(self interface{}, ctxt PairContext, vm PairVisitorManager, leftArgs, rightArgs map[string]*pb.Data) Cont {
	keepGoing := Continue
	specPairContext := ctxt.(SpecPairVisitorContext)

	// Obtain the methods' HTTP metadata from the context.
	leftMethodMeta := (*pb.HTTPMethodMeta)(nil)
	rightMethodMeta := (*pb.HTTPMethodMeta)(nil)
	{
		leftNode, rightNode, _ := specPairContext.GetInnermostNode(reflect.TypeOf((*pb.Method)(nil)))
		leftMethodMeta = leftNode.(*pb.Method).GetMeta().GetHttp()
		rightMethodMeta = rightNode.(*pb.Method).GetMeta().GetHttp()
	}

	// Normalize arguments on both sides.
	// XXX Ignoring errors.
	normalizedLeft, _ := GetNormalizedArgNames(leftArgs, leftMethodMeta)
	normalizedRight, _ := GetNormalizedArgNames(rightArgs, rightMethodMeta)

	// Line up left arguments with the right and visit in pairs. Remove any
	// matching arguments on the right.
	for name, leftName := range normalizedLeft {
		leftArg := leftArgs[leftName]
		if rightName, ok := normalizedRight[name]; ok {
			rightArg := rightArgs[rightName]
			ctxt := ctxt.EnterMapValues(leftArgs, leftName, rightArgs, rightName)
			keepGoing = go_ast_pair.ApplyWithContext(vm, ctxt, leftArg, rightArg)
			delete(normalizedRight, name)
		} else {
			ctxt := ctxt.EnterMapValues(leftArgs, leftName, rightArgs, nil)
			keepGoing = go_ast_pair.ApplyWithContext(vm, ctxt, leftArg, go_ast_pair.ZeroOf(leftArg))
		}

		if result := handleKeepGoing(keepGoing); result != nil {
			return *result
		}
	}

	// Any remaining arguments on the right don't have a match on the left.
	for _, rightName := range normalizedRight {
		rightArg := rightArgs[rightName]
		ctxt := ctxt.EnterMapValues(leftArgs, nil, rightArgs, rightName)
		keepGoing = go_ast_pair.ApplyWithContext(vm, ctxt, (go_ast_pair.ZeroOf(rightArg)), rightArg)

		if result := handleKeepGoing(keepGoing); result != nil {
			return *result
		}
	}

	return Continue
}

// == MethodMeta ==============================================================

func (*DefaultSpecPairVisitorImpl) EnterMethodMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.MethodMeta) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitMethodMetaChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.MethodMeta) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveMethodMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.MethodMeta, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == HTTPMethodMeta ==========================================================

func (*DefaultSpecPairVisitorImpl) EnterHTTPMethodMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPMethodMeta) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitHTTPMethodMetaChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPMethodMeta) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveHTTPMethodMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPMethodMeta, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == Data ====================================================================

func (*DefaultSpecPairVisitorImpl) EnterData(self interface{}, c SpecPairVisitorContext, left, right *pb.Data) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitDataChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Data) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveData(self interface{}, c SpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == DataMeta ================================================================

func (*DefaultSpecPairVisitorImpl) EnterDataMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.DataMeta) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitDataMetaChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.DataMeta) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveDataMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.DataMeta, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == HTTPMeta ================================================================

func (*DefaultSpecPairVisitorImpl) EnterHTTPMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPMeta) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitHTTPMetaChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPMeta) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveHTTPMetas(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPMeta, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == HTTPPath ================================================================

func (*DefaultSpecPairVisitorImpl) EnterHTTPPaths(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPPath) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitHTTPPathChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPPath) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveHTTPPaths(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPPath, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == HTTPQuery ===============================================================

func (*DefaultSpecPairVisitorImpl) EnterHTTPQueries(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPQuery) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitHTTPQueryChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPQuery) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveHTTPQueries(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPQuery, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == HTTPHeader ==============================================================

func (*DefaultSpecPairVisitorImpl) EnterHTTPHeaders(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPHeader) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitHTTPHeaderChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPHeader) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveHTTPHeaders(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPHeader, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == HTTPCookie ==============================================================

func (*DefaultSpecPairVisitorImpl) EnterHTTPCookies(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPCookie) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitHTTPCookieChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPCookie) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveHTTPCookies(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPCookie, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == HTTPBody ================================================================

func (*DefaultSpecPairVisitorImpl) EnterHTTPBodies(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPBody) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitHTTPBodyChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPBody) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveHTTPBodies(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPBody, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == HTTPEmpty ===============================================================

func (*DefaultSpecPairVisitorImpl) EnterHTTPEmpties(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPEmpty) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitHTTPEmptyChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPEmpty) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveHTTPEmpties(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPEmpty, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == HTTPAuth ================================================================

func (*DefaultSpecPairVisitorImpl) EnterHTTPAuths(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPAuth) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitHTTPAuthChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPAuth) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveHTTPAuths(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPAuth, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == HTTPMultipart ===========================================================

func (*DefaultSpecPairVisitorImpl) EnterHTTPMultiparts(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPMultipart) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitHTTPMultipartChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.HTTPMultipart) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveHTTPMultiparts(self interface{}, c SpecPairVisitorContext, left, right *pb.HTTPMultipart, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == Primitive ===============================================================

func (*DefaultSpecPairVisitorImpl) EnterPrimitives(self interface{}, c SpecPairVisitorContext, left, right *pb.Primitive) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitPrimitiveChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Primitive) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeavePrimitives(self interface{}, c SpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == Struct ==================================================================

func (*DefaultSpecPairVisitorImpl) EnterStructs(self interface{}, c SpecPairVisitorContext, left, right *pb.Struct) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitStructChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Struct) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveStructs(self interface{}, c SpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == List ====================================================================

func (*DefaultSpecPairVisitorImpl) EnterLists(self interface{}, c SpecPairVisitorContext, left, right *pb.List) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitListChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.List) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveLists(self interface{}, c SpecPairVisitorContext, left, right *pb.List, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == Optional ================================================================

func (*DefaultSpecPairVisitorImpl) EnterOptionals(self interface{}, c SpecPairVisitorContext, left, right *pb.Optional) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitOptionalChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Optional) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveOptionals(self interface{}, c SpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == OneOf ===================================================================

func (*DefaultSpecPairVisitorImpl) EnterOneOfs(self interface{}, c SpecPairVisitorContext, left, right *pb.OneOf) Cont {
	return self.(DefaultSpecPairVisitor).EnterNodes(self, c, left, right)
}

func (*DefaultSpecPairVisitorImpl) VisitOneOfChildren(self interface{}, c SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.OneOf) Cont {
	return self.(DefaultSpecPairVisitor).VisitNodeChildren(self, c, vm, left, right)
}

func (*DefaultSpecPairVisitorImpl) LeaveOneOfs(self interface{}, c SpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont {
	return self.(DefaultSpecPairVisitor).LeaveNodes(self, c, left, right, cont)
}

// == Different types =========================================================

func (*DefaultSpecPairVisitorImpl) EnterDifferentTypes(self interface{}, c SpecPairVisitorContext, left, right interface{}) Cont {
	return SkipChildren
}

func (*DefaultSpecPairVisitorImpl) LeaveDifferentTypes(self interface{}, c SpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	return cont
}

// extendContext implementation for SpecPairVisitor.
func extendPairContext(cin PairContext, left, right interface{}) {
	ctx, ok := cin.(SpecPairVisitorContext)
	if !ok {
		panic(fmt.Sprintf("http_rest.extendPairContext expected SpecPairVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}

	ctx.ExtendLeftContext(left)
	ctx.ExtendRightContext(right)
}

// enter implementation for SpecVisitor.
func enterPair(cin PairContext, visitor interface{}, left, right interface{}) Cont {
	v, _ := visitor.(SpecPairVisitor)
	ctx, ok := cin.(SpecPairVisitorContext)
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
	ctx, ok := cin.(SpecPairVisitorContext)
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
