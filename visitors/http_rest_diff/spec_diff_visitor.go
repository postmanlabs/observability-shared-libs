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

// A SpecPairVisitor with hooks for processing each difference found between
// two IR trees. A node is considered changed if a difference can be observed
// at that level of the IR. For example, HTTPAuth nodes with different Types
// are considered changed, but their parents might not necessarily be
// considered changed.
//
// Go lacks virtual functions, so all functions here take the visitor itself as
// an argument, and call functions on that instance.
type SpecDiffVisitor interface {
	http_rest.SpecPairVisitor

	EnterAddedOrRemovedData(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Data) Cont
	LeaveAddedOrRemovedData(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont

	EnterAddedOrRemovedList(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.List) Cont
	LeaveAddedOrRemovedList(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.List, cont Cont) Cont

	EnterAddedOrRemovedOneOf(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.OneOf) Cont
	LeaveAddedOrRemovedOneOf(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont

	EnterAddedOrRemovedOptional(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Optional) Cont
	LeaveAddedOrRemovedOptional(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont

	EnterAddedOrRemovedPrimitive(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Primitive) Cont
	EnterChangedPrimitive(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Primitive) Cont
	LeaveAddedOrRemovedPrimitive(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont
	LeaveChangedPrimitive(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont

	EnterAddedOrRemovedStruct(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Struct) Cont
	LeaveAddedOrRemovedStruct(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont
}

// A SpecDiffVisitor with convenience functions for entering and leaving nodes
// with diffs.
type DefaultSpecDiffVisitor interface {
	SpecDiffVisitor

	EnterDiff(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}) Cont
	LeaveDiff(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}, cont Cont) Cont

	// Delegates to EnterDiff by default.
	EnterAddedOrRemovedNode(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}) Cont

	// Delegates to EnterDiff by default.
	EnterChangedNode(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}) Cont

	// Delegates to LeaveDiff by default.
	LeaveAddedOrRemovedNode(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}, cont Cont) Cont

	// Delegates to LeaveDiff by default.
	LeaveChangedNode(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}, cont Cont) Cont
}

// A SpecDiffVisitor implementation. This does not traverse into the children
// of nodes that were added, removed, or changed.
type DefaultSpecDiffVisitorImpl struct {
	http_rest.DefaultSpecPairVisitorImpl
}

var _ DefaultSpecDiffVisitor = (*DefaultSpecDiffVisitorImpl)(nil)

// == Default implementations =================================================

func (*DefaultSpecDiffVisitorImpl) EnterDiff(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}) Cont {
	return SkipChildren
}

func (*DefaultSpecDiffVisitorImpl) LeaveDiff(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	return cont
}

// Delegates to EnterDiff.
func (*DefaultSpecDiffVisitorImpl) EnterAddedOrRemovedNode(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.EnterDiff(self, ctx, left, right)
}

// Delegates to EnterDiff.
func (*DefaultSpecDiffVisitorImpl) EnterChangedNode(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.EnterDiff(self, ctx, left, right)
}

// Delegates to LeaveDiff.
func (*DefaultSpecDiffVisitorImpl) LeaveAddedOrRemovedNode(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.LeaveDiff(self, ctx, left, right, cont)
}

// Delegates to LeaveDiff.
func (*DefaultSpecDiffVisitorImpl) LeaveChangedNode(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.LeaveDiff(self, ctx, left, right, cont)
}

// == Data ====================================================================

func (*DefaultSpecDiffVisitorImpl) EnterData(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Data) Cont {
	v := self.(DefaultSpecDiffVisitor)

	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedData(self, ctx, left, right)
	}

	return Continue
}

func (*DefaultSpecDiffVisitorImpl) VisitDataChildren(self interface{}, ctx http_rest.SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Data) Cont {
	// Only visit the value.
	childCtx := ctx.EnterStructs(left, "Value", right, "Value")
	return go_ast_pair.ApplyWithContext(vm, childCtx, left.Value, right.Value)
}

func (*DefaultSpecDiffVisitorImpl) LeaveData(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)

	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedData(self, ctx, left, right, cont)
	}

	return cont
}

// Delegates to EnterAddedOrRemovedNode.
func (*DefaultSpecDiffVisitorImpl) EnterAddedOrRemovedData(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Data) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultSpecDiffVisitorImpl) LeaveAddedOrRemovedData(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// == List ====================================================================

func (*DefaultSpecDiffVisitorImpl) EnterLists(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.List) Cont {
	v := self.(DefaultSpecDiffVisitor)

	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedList(self, ctx, left, right)
	}

	return Continue
}

func (*DefaultSpecDiffVisitorImpl) LeaveLists(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.List, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)

	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedList(self, ctx, left, right, cont)
	}

	return cont
}

// Delegates to EnterAddedOrRemovedNode.
func (*DefaultSpecDiffVisitorImpl) EnterAddedOrRemovedList(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.List) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultSpecDiffVisitorImpl) LeaveAddedOrRemovedList(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.List, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// == OneOf ===================================================================

func (*DefaultSpecDiffVisitorImpl) EnterOneOfs(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.OneOf) Cont {
	v := self.(DefaultSpecDiffVisitor)

	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedOneOf(self, ctx, left, right)
	}

	return Continue
}

func (*DefaultSpecDiffVisitorImpl) VisitOneOfChildren(self interface{}, ctx http_rest.SpecPairVisitorContext, vm PairVisitorManager, left, right *pb.OneOf) Cont {
	// Override visitor behaviour for OneOf nodes by manually pairing up the
	// options to see if we can get things to match.
	childCtx := ctx.EnterStructs(left, "Options", right, "Options")
	rightOptions := make(map[string]*pb.Data, len(right.Options))
	for k, v := range right.Options {
		rightOptions[k] = v
	}
OUTER:
	for leftKey, leftOption := range left.Options {
		for rightKey, rightOption := range rightOptions {
			if IsSameData(leftOption, rightOption) {
				// Found a match.
				delete(rightOptions, rightKey)
				continue OUTER
			}
		}
		// No match found for leftOption.
		childCtx := childCtx.EnterMapValues(left.Options, leftKey, right.Options, nil)
		switch keepGoing := go_ast_pair.ApplyWithContext(vm, childCtx, leftOption, go_ast_pair.ZeroOf(leftOption)); keepGoing {
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
	for rightKey, rightOption := range rightOptions {
		childCtx := childCtx.EnterMapValues(left.Options, nil, right.Options, rightKey)
		switch keepGoing := go_ast_pair.ApplyWithContext(vm, childCtx, go_ast_pair.ZeroOf(rightOption), rightOption); keepGoing {
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

func (*DefaultSpecDiffVisitorImpl) LeaveOneOfs(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)

	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedOneOf(self, ctx, left, right, cont)
	}

	return cont
}

// Delegates to EnterAddedOrRemovedNode.
func (*DefaultSpecDiffVisitorImpl) EnterAddedOrRemovedOneOf(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.OneOf) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultSpecDiffVisitorImpl) LeaveAddedOrRemovedOneOf(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// == Optional ================================================================

func (*DefaultSpecDiffVisitorImpl) EnterOptionals(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Optional) Cont {
	v := self.(DefaultSpecDiffVisitor)

	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedOptional(self, ctx, left, right)
	}

	return Continue
}

func (*DefaultSpecDiffVisitorImpl) LeaveOptionals(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)

	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedOptional(self, ctx, left, right, cont)
	}

	return cont
}

// Delegates to EnterAddedOrRemovedNode.
func (*DefaultSpecDiffVisitorImpl) EnterAddedOrRemovedOptional(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Optional) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultSpecDiffVisitorImpl) LeaveAddedOrRemovedOptional(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// == Primitive ===============================================================

func (*DefaultSpecDiffVisitorImpl) EnterPrimitives(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Primitive) Cont {
	v := self.(DefaultSpecDiffVisitor)

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

func (*DefaultSpecDiffVisitorImpl) LeavePrimitives(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)

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
func (*DefaultSpecDiffVisitorImpl) EnterAddedOrRemovedPrimitive(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Primitive) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to EnterChangedNode.
func (*DefaultSpecDiffVisitorImpl) EnterChangedPrimitive(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Primitive) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.EnterChangedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultSpecDiffVisitorImpl) LeaveAddedOrRemovedPrimitive(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// Delegates to LeaveChangedNode.
func (*DefaultSpecDiffVisitorImpl) LeaveChangedPrimitive(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.LeaveChangedNode(self, ctx, left, right, cont)
}

// == Struct ==================================================================

func (*DefaultSpecDiffVisitorImpl) EnterStructs(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Struct) Cont {
	v := self.(DefaultSpecDiffVisitor)

	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedStruct(self, ctx, left, right)
	}

	return Continue
}

func (*DefaultSpecDiffVisitorImpl) LeaveStructs(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)

	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedStruct(self, ctx, left, right, cont)
	}

	return cont
}

// Delegates to EnterAddedOrRemovedNode.
func (*DefaultSpecDiffVisitorImpl) EnterAddedOrRemovedStruct(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Struct) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.EnterAddedOrRemovedNode(self, ctx, left, right)
}

// Delegates to LeaveAddedOrRemovedNode.
func (*DefaultSpecDiffVisitorImpl) LeaveAddedOrRemovedStruct(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.LeaveAddedOrRemovedNode(self, ctx, left, right, cont)
}

// Delegates to EnterDiff.
func (*DefaultSpecDiffVisitorImpl) EnterDifferentTypes(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.EnterChangedNode(self, ctx, left, right)
}

// Delegates to LeaveDiff.
func (*DefaultSpecDiffVisitorImpl) LeaveDifferentTypes(self interface{}, ctx http_rest.SpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	v := self.(DefaultSpecDiffVisitor)
	return v.LeaveChangedNode(self, ctx, left, right, cont)
}
