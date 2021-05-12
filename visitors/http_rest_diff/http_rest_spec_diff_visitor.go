package http_rest_diff

import (
	"fmt"
	"reflect"
	"sort"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/spec_util"
	. "github.com/akitasoftware/akita-libs/visitors"
	"github.com/akitasoftware/akita-libs/visitors/go_ast_pair"
	"github.com/akitasoftware/akita-libs/visitors/http_rest"
)

// An HttpRestSpecPairVisitor with hooks for processing each difference found
// between two IR trees. A node is considered changed if a difference can be
// observed at that level of the IR. For example, HTTPAuth nodes with different
// Types are considered changed, but their parents might not necessarily be
// considered changed.
//
// Go lacks virtual functions, so all functions here take the visitor itself as
// an argument, and call functions on that instance.
type HttpRestSpecDiffVisitor interface {
	http_rest.HttpRestSpecPairVisitor

	EnterAddedOrRemovedHTTPAuth(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont
	EnterChangedHTTPAuth(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont
	LeaveAddedOrRemovedHTTPAuth(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth, cont Cont) Cont
	LeaveChangedHTTPAuth(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth, cont Cont) Cont

	EnterAddedOrRemovedData(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Data) Cont
	LeaveAddedOrRemovedData(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont

	EnterAddedOrRemovedList(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.List) Cont
	LeaveAddedOrRemovedList(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.List, cont Cont) Cont

	EnterAddedOrRemovedOneOf(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.OneOf) Cont
	LeaveAddedOrRemovedOneOf(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont

	EnterAddedOrRemovedOptional(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Optional) Cont
	LeaveAddedOrRemovedOptional(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont

	EnterAddedOrRemovedPrimitive(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive) Cont
	EnterChangedPrimitive(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive) Cont
	LeaveAddedOrRemovedPrimitive(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont
	LeaveChangedPrimitive(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont

	EnterAddedOrRemovedStruct(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Struct) Cont
	LeaveAddedOrRemovedStruct(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont
}

// An HttpRestSpecDiffVisitor with convenience functions for entering and
// leaving nodes with diffs.
type DefaultHttpRestSpecDiffVisitor interface {
	HttpRestSpecDiffVisitor

	EnterDiff(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}) Cont
	LeaveDiff(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}, cont Cont) Cont

	// Delegates to EnterDiff by default.
	EnterAddedOrRemovedNode(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}) Cont

	// Delegates to EnterDiff by default.
	EnterChangedNode(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}) Cont

	// Delegates to LeaveDiff by default.
	LeaveAddedOrRemovedNode(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}, cont Cont) Cont

	// Delegates to LeaveDiff by default.
	LeaveChangedNode(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}, cont Cont) Cont
}

// An HttpRestSpecDiffVisitor implementation. This does not traverse into the
// children of nodes that were added, removed, or changed.
type DefaultHttpRestSpecDiffVisitorImpl struct {
	http_rest.DefaultHttpRestSpecPairVisitor
}

var _ DefaultHttpRestSpecDiffVisitor = &DefaultHttpRestSpecDiffVisitorImpl{}

// == Default implementations =================================================

func (*DefaultHttpRestSpecDiffVisitorImpl) EnterDiff(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}) Cont {
	return SkipChildren
}

func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveDiff(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	return cont
}

// Delegates to EnterDiff.
func (*DefaultHttpRestSpecDiffVisitorImpl) EnterAddedOrRemovedNode(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.EnterDiff(self, ctx, left, right)
}

// Delegates to EnterDiff.
func (*DefaultHttpRestSpecDiffVisitorImpl) EnterChangedNode(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.EnterDiff(self, ctx, left, right)
}

// Delegates to LeaveDiff.
func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveAddedOrRemovedNode(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.LeaveDiff(self, ctx, left, right, cont)
}

// Delegates to LeaveDiff.
func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveChangedNode(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.LeaveDiff(self, ctx, left, right, cont)
}

// == HTTPAuth ================================================================

func (*DefaultHttpRestSpecDiffVisitorImpl) EnterHTTPAuths(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont {
	v := self.(HttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedHTTPAuth(self, ctx, left, right)
	}

	if left.Type != right.Type {
		return v.EnterChangedHTTPAuth(self, ctx, left, right)
	}

	return Continue
}

func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveHTTPAuths(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth, cont Cont) Cont {
	v := self.(HttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedHTTPAuth(self, ctx, left, right, cont)
	}

	if left.Type != right.Type {
		return v.LeaveChangedHTTPAuth(self, ctx, left, right, cont)
	}

	return cont
}

// Delegates to EnterAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) EnterAddedOrRemovedHTTPAuth(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to EnterChangedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) EnterChangedHTTPAuth(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.EnterChangedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveAddedOrRemovedHTTPAuth(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// Delegates to LeaveChangedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveChangedHTTPAuth(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.LeaveChangedNode(self, ctx, left, right, cont)
}

// == Data ====================================================================

func (*DefaultHttpRestSpecDiffVisitorImpl) EnterData(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Data) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedData(self, ctx, left, right)
	}

	return Continue
}

func (*DefaultHttpRestSpecDiffVisitorImpl) VisitDataChildren(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Data) Cont {
	// Only visit the value.
	childCtx := ctx.AppendPaths("Value", "Value")
	return go_ast_pair.ApplyWithContext(vm, childCtx, left.Value, right.Value)
}

func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveData(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedData(self, ctx, left, right, cont)
	}

	return cont
}

// Delegates to EnterAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) EnterAddedOrRemovedData(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Data) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveAddedOrRemovedData(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// == List ====================================================================

func (*DefaultHttpRestSpecDiffVisitorImpl) EnterLists(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.List) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedList(self, ctx, left, right)
	}

	return Continue
}

func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveLists(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.List, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedList(self, ctx, left, right, cont)
	}

	return cont
}

// Delegates to EnterAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) EnterAddedOrRemovedList(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.List) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveAddedOrRemovedList(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.List, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// == OneOf ===================================================================

func (*DefaultHttpRestSpecDiffVisitorImpl) EnterOneOfs(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.OneOf) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedOneOf(self, ctx, left, right)
	}

	return Continue
}

func (*DefaultHttpRestSpecDiffVisitorImpl) VisitOneOfChildren(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.OneOf) Cont {
	// Override visitor behaviour for OneOf nodes by manually pairing up the
	// options to see if we can get things to match.
	childCtx := ctx.AppendPaths("Options", "Options")
	rightOptions := make(map[string]*pb.Data, len(right.Options))
	for k, v := range right.Options {
		rightOptions[k] = v
	}
OUTER:
	for _, leftOption := range left.Options {
		for rightKey, rightOption := range rightOptions {
			if IsSameData(leftOption, rightOption) {
				// Found a match.
				delete(rightOptions, rightKey)
				continue OUTER
			}
		}
		// No match found for leftOption.
		switch keepGoing := go_ast_pair.ApplyWithContext(vm, childCtx.AppendPaths(fmt.Sprint(leftOption), ""), leftOption, nil); keepGoing {
		case Abort:
			return Abort
		case Continue:
		case SkipChildren:
			panic("go_ast_pair.ApplyWithContext returned SkipChildren")
		case Stop:
			return Stop
		default:
			panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
		}
	}

	// Anything remaining in `rightOptions` has no match in `left`.
	for _, rightOption := range rightOptions {
		switch keepGoing := go_ast_pair.ApplyWithContext(vm, childCtx.AppendPaths("", fmt.Sprint(rightOption)), nil, rightOption); keepGoing {
		case Abort:
			return Abort
		case Continue:
		case SkipChildren:
			panic("go_ast_pair.ApplyWithContext returned SkipChildren")
		case Stop:
			return Stop
		default:
			panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
		}
	}

	return SkipChildren
}

func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveOneOfs(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedOneOf(self, ctx, left, right, cont)
	}

	return cont
}

// Delegates to EnterAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) EnterAddedOrRemovedOneOf(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.OneOf) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveAddedOrRemovedOneOf(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// == Optional ================================================================

func (*DefaultHttpRestSpecDiffVisitorImpl) EnterOptionals(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Optional) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedOptional(self, ctx, left, right)
	}

	return Continue
}

func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveOptionals(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedOptional(self, ctx, left, right, cont)
	}

	return cont
}

// Delegates to EnterAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) EnterAddedOrRemovedOptional(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Optional) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveAddedOrRemovedOptional(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// == Primitive ===============================================================

func (*DefaultHttpRestSpecDiffVisitorImpl) EnterPrimitives(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedPrimitive(self, ctx, left, right)
	}

	if primitivesDiffer(left, right) {
		return v.EnterChangedPrimitive(self, ctx, left, right)
	}

	return SkipChildren
}

// Determines whether two primitives are different.
func primitivesDiffer(p1, p2 *pb.Primitive) bool {
	// Compare types.
	type1 := spec_util.TypeOfPrimitive(p1)
	type2 := spec_util.TypeOfPrimitive(p2)
	if type1 != type2 {
		return true
	}

	// Compare format kinds.
	if p1.FormatKind != p2.FormatKind {
		return true
	}

	// Compare formats.
	formats1 := formatsOfPrimitive(p1)
	formats2 := formatsOfPrimitive(p2)
	if !reflect.DeepEqual(formats1, formats2) {
		return true
	}

	return false
}

// Extracts a list of formats from a primitive.
func formatsOfPrimitive(p *pb.Primitive) []string {
	result := make([]string, 0, len(p.Formats))
	for format, present := range p.Formats {
		if present {
			result = append(result, format)
		}
	}
	sort.Strings(result)
	return result
}

func (*DefaultHttpRestSpecDiffVisitorImpl) LeavePrimitives(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedPrimitive(self, ctx, left, right, cont)
	}

	if primitivesDiffer(left, right) {
		return v.LeaveChangedPrimitive(self, ctx, left, right, cont)
	}

	return cont
}

// Delegates to EnterAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) EnterAddedOrRemovedPrimitive(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to EnterChangedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) EnterChangedPrimitive(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.EnterChangedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveAddedOrRemovedPrimitive(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// Delegates to LeaveChangedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveChangedPrimitive(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.LeaveChangedNode(self, ctx, left, right, cont)
}

// == Struct ==================================================================

func (*DefaultHttpRestSpecDiffVisitorImpl) EnterStructs(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Struct) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedStruct(self, ctx, left, right)
	}

	return Continue
}

func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveStructs(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)

	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedStruct(self, ctx, left, right, cont)
	}

	return cont
}

// Delegates to EnterAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) EnterAddedOrRemovedStruct(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Struct) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveAddedOrRemovedStruct(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// Delegates to EnterDiff.
func (*DefaultHttpRestSpecDiffVisitorImpl) EnterDifferentTypes(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.EnterDiff(self, ctx, left, right)
}

// Delegates to LeaveDiff.
func (*DefaultHttpRestSpecDiffVisitorImpl) LeaveDifferentTypes(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	v := self.(DefaultHttpRestSpecDiffVisitor)
	return v.LeaveDiff(self, ctx, left, right, cont)
}
