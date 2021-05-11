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
type HttpRestSpecDiffVisitor interface {
	http_rest.HttpRestSpecPairVisitor

	EnterAddedOrRemovedHTTPAuth(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont
	EnterChangedHTTPAuth(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont
	LeaveAddedOrRemovedHTTPAuth(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont
	LeaveChangedHTTPAuth(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont

	EnterAddedOrRemovedData(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Data) Cont
	LeaveAddedOrRemovedData(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont

	EnterAddedOrRemovedList(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.List) Cont
	LeaveAddedOrRemovedList(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.List, cont Cont) Cont

	EnterAddedOrRemovedOneOf(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.OneOf) Cont
	LeaveAddedOrRemovedOneOf(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont

	EnterAddedOrRemovedOptional(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Optional) Cont
	LeaveAddedOrRemovedOptional(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont

	EnterAddedOrRemovedPrimitive(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive) Cont
	EnterChangedPrimitive(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive) Cont
	LeaveAddedOrRemovedPrimitive(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont
	LeaveChangedPrimitive(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive) Cont

	EnterAddedOrRemovedStruct(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Struct) Cont
	LeaveAddedOrRemovedStruct(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont
}

// An HttpRestSpecDiffVisitor implementation. This does not traverse into the
// children of nodes that were added, removed, or changed.
type DefaultHttpRestSpecDiffVisitor struct {
	http_rest.DefaultHttpRestSpecPairVisitor
}

// == Default implementations =================================================

func (v *DefaultHttpRestSpecDiffVisitor) DefaultEnterDiff(ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}) Cont {
	return SkipChildren
}

func (v *DefaultHttpRestSpecDiffVisitor) DefaultLeaveDiff(ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	return cont
}

// Delegates to DefaultEnterDiff.
func (v *DefaultHttpRestSpecDiffVisitor) DefaultEnterAddedOrRemovedNode(ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}) Cont {
	return v.DefaultEnterDiff(ctx, left, right)
}

// Delegates to DefaultEnterDiff.
func (v *DefaultHttpRestSpecDiffVisitor) DefaultEnterChangedNode(ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}) Cont {
	return v.DefaultEnterDiff(ctx, left, right)
}

// Delegates to DefaultLeaveDiff.
func (v *DefaultHttpRestSpecDiffVisitor) DefaultLeaveAddedOrRemovedNode(ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	return v.DefaultLeaveDiff(ctx, left, right, cont)
}

// Delegates to DefaultLeaveDiff.
func (v *DefaultHttpRestSpecDiffVisitor) DefaultLeaveChangedNode(ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	return v.DefaultLeaveDiff(ctx, left, right, cont)
}

// == HTTPAuth ================================================================

func (v *DefaultHttpRestSpecDiffVisitor) EnterHTTPAuth(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont {
	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedHTTPAuth(ctx, left, right)
	}

	if left.Type != right.Type {
		return v.EnterChangedHTTPAuth(ctx, left, right)
	}

	return Continue
}

func (v *DefaultHttpRestSpecDiffVisitor) LeaveHTTPAuth(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth, cont Cont) Cont {
	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedHTTPAuth(ctx, left, right, cont)
	}

	if left.Type != right.Type {
		return v.LeaveChangedHTTPAuth(ctx, left, right, cont)
	}

	return cont
}

// Delegates to DefaultEnterAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) EnterAddedOrRemovedHTTPAuth(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont {
	return v.DefaultEnterAddedOrRemovedNode(ctx, left, right)
}

// Delegates to DefaultEnterChangedNode.
func (v *DefaultHttpRestSpecDiffVisitor) EnterChangedHTTPAuth(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth) Cont {
	return v.DefaultEnterChangedNode(ctx, left, right)
}

// Delegates to DefaultLeaveAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) LeaveAddedOrRemovedHTTPAuth(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth, cont Cont) Cont {
	return v.DefaultLeaveAddedOrRemovedNode(ctx, left, right, cont)
}

// Delegates to DefaultLeaveChangedNode.
func (v *DefaultHttpRestSpecDiffVisitor) LeaveChangedHTTPAuth(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.HTTPAuth, cont Cont) Cont {
	return v.DefaultLeaveChangedNode(ctx, left, right, cont)
}

// == Data ====================================================================

func (v *DefaultHttpRestSpecDiffVisitor) EnterData(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Data) Cont {
	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedData(ctx, left, right)
	}

	return Continue
}

func (*DefaultHttpRestSpecDiffVisitor) VisitDataChildren(ctx http_rest.HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.Data) Cont {
	// Only visit the value.
	childCtx := ctx.AppendPaths("Value", "Value")
	return go_ast_pair.ApplyWithContext(vm, childCtx, left.Value, right.Value)
}

func (v *DefaultHttpRestSpecDiffVisitor) LeaveData(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont {
	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedData(ctx, left, right, cont)
	}

	return cont
}

// Delegates to DefaultEnterAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) EnterAddedOrRemovedData(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Data) Cont {
	return v.DefaultEnterAddedOrRemovedNode(ctx, left, right)
}

// Delegates to DefaultLeaveAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) LeaveAddedOrRemovedData(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Data, cont Cont) Cont {
	return v.DefaultLeaveAddedOrRemovedNode(ctx, left, right, cont)
}

// == List ====================================================================

func (v *DefaultHttpRestSpecDiffVisitor) EnterList(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.List) Cont {
	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedList(ctx, left, right)
	}

	return Continue
}

func (v *DefaultHttpRestSpecDiffVisitor) LeaveList(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.List, cont Cont) Cont {
	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedList(ctx, left, right, cont)
	}

	return cont
}

// Delegates to DefaultEnterAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) EnterAddedOrRemovedList(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.List) Cont {
	return v.DefaultEnterAddedOrRemovedNode(ctx, left, right)
}

// Delegates to DefaultLeaveAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) LeaveAddedOrRemovedList(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.List, cont Cont) Cont {
	return v.DefaultLeaveAddedOrRemovedNode(ctx, left, right, cont)
}

// == OneOf ===================================================================

func (v *DefaultHttpRestSpecDiffVisitor) EnterOneOf(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.OneOf) Cont {
	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedOneOf(ctx, left, right)
	}

	return Continue
}

func (v *DefaultHttpRestSpecDiffVisitor) VisitOneOfChildren(ctx http_rest.HttpRestSpecPairVisitorContext, vm PairVisitorManager, left, right *pb.OneOf) Cont {
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

func (v *DefaultHttpRestSpecDiffVisitor) LeaveOneOf(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont {
	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedOneOf(ctx, left, right, cont)
	}

	return cont
}

// Delegates to DefaultEnterAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) EnterAddedOrRemovedOneOf(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.OneOf) Cont {
	return v.DefaultEnterAddedOrRemovedNode(ctx, left, right)
}

// Delegates to DefaultLeaveAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) LeaveAddedOrRemovedOneOf(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.OneOf, cont Cont) Cont {
	return v.DefaultLeaveAddedOrRemovedNode(ctx, left, right, cont)
}

// == Optional ================================================================

func (v *DefaultHttpRestSpecDiffVisitor) EnterOptional(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Optional) Cont {
	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedOptional(ctx, left, right)
	}

	return Continue
}

func (v *DefaultHttpRestSpecDiffVisitor) LeaveOptional(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont {
	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedOptional(ctx, left, right, cont)
	}

	return cont
}

// Delegates to DefaultEnterAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) EnterAddedOrRemovedOptional(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Optional) Cont {
	return v.DefaultEnterAddedOrRemovedNode(ctx, left, right)
}

// Delegates to DefaultLeaveAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) LeaveAddedOrRemovedOptional(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Optional, cont Cont) Cont {
	return v.DefaultLeaveAddedOrRemovedNode(ctx, left, right, cont)
}

// == Primitive ===============================================================

func (v *DefaultHttpRestSpecDiffVisitor) EnterPrimitive(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive) Cont {
	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedPrimitive(ctx, left, right)
	}

	if primitivesDiffer(left, right) {
		return v.EnterChangedPrimitive(ctx, left, right)
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

func (v *DefaultHttpRestSpecDiffVisitor) LeavePrimitive(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont {
	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedPrimitive(ctx, left, right, cont)
	}

	if primitivesDiffer(left, right) {
		return v.LeaveChangedPrimitive(ctx, left, right, cont)
	}

	return cont
}

// Delegates to DefaultEnterAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) EnterAddedOrRemovedPrimitive(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive) Cont {
	return v.DefaultEnterAddedOrRemovedNode(ctx, left, right)
}

// Delegates to DefaultEnterChangedNode.
func (v *DefaultHttpRestSpecDiffVisitor) EnterChangedPrimitive(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive) Cont {
	return v.DefaultEnterChangedNode(ctx, left, right)
}

// Delegates to DefaultLeaveAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) LeaveAddedOrRemovedPrimitive(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont {
	return v.DefaultLeaveAddedOrRemovedNode(ctx, left, right, cont)
}

// Delegates to DefaultLeaveChangedNode.
func (v *DefaultHttpRestSpecDiffVisitor) LeaveChangedPrimitive(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Primitive, cont Cont) Cont {
	return v.DefaultLeaveChangedNode(ctx, left, right, cont)
}

// == Struct ==================================================================

func (v *DefaultHttpRestSpecDiffVisitor) EnterStruct(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Struct) Cont {
	if left == nil && right == nil {
		return SkipChildren
	}

	if left == nil || right == nil {
		return v.EnterAddedOrRemovedStruct(ctx, left, right)
	}

	return Continue
}

func (v *DefaultHttpRestSpecDiffVisitor) LeaveStruct(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont {
	if left == nil && right == nil {
		return cont
	}

	if left == nil || right == nil {
		return v.LeaveAddedOrRemovedStruct(ctx, left, right, cont)
	}

	return cont
}

// Delegates to DefaultEnterAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) EnterAddedOrRemovedStruct(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Struct) Cont {
	return v.DefaultEnterAddedOrRemovedNode(ctx, left, right)
}

// Delegates to DefaultLeaveAddedOrRemovedNode.
func (v *DefaultHttpRestSpecDiffVisitor) LeaveAddedOrRemovedStruct(ctx http_rest.HttpRestSpecPairVisitorContext, left, right *pb.Struct, cont Cont) Cont {
	return v.DefaultLeaveAddedOrRemovedNode(ctx, left, right, cont)
}

// Delegates to DefaultEnterDiff.
func (v *DefaultHttpRestSpecDiffVisitor) EnterDifferentTypes(ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}) Cont {
	return v.DefaultEnterDiff(ctx, left, right)
}

// Delegates to DefaultLeaveDiff.
func (v *DefaultHttpRestSpecDiffVisitor) LeaveDifferentTypes(ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}, cont Cont) Cont {
	return v.DefaultLeaveDiff(ctx, left, right, cont)
}
