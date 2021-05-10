package go_ast_pair

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	. "github.com/akitasoftware/akita-libs/visitors"
)

// Structurally recurses through `leftNode` and `rightNode` in tandem.
func Apply(v PairVisitorManager, leftNode, rightNode interface{}) Cont {
	return ApplyWithContext(v, v.Context(), leftNode, rightNode)
}

func ApplyWithContext(v PairVisitorManager, ctx PairContext, leftNode, rightNode interface{}) Cont {
	astv := astPairVisitor{vm: v}
	return astv.visit(ctx, leftNode, rightNode)
}

type astPairVisitor struct {
	vm PairVisitorManager
}

// Visits the nodes left and right in tandem, whose context is c. At least one
// of left or right must be non-nil. When the visitor finds a structural
// difference between the two sides (e.g., one side is nil or the two sides
// have different kinds), the nodes are entered, but their children are not.
//
// This should never return SkipChildren.
func (t *astPairVisitor) visit(c PairContext, left, right interface{}) Cont {
	if left == nil && right == nil {
		return Continue
	}

	keepGoing := t.vm.EnterNodes(c, t.vm.Visitor(), left, right)
	switch keepGoing {
	case Abort:
		return Abort
	case Continue:
	case SkipChildren:
	case Stop:
	default:
		panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
	}

	// Don't visit children if we are stopping or skipping children.
	if keepGoing == Continue {
		newContext := t.vm.ExtendContext(c, left, right)
		keepGoing = t.vm.VisitChildren(newContext, t.vm, left, right)
	}

	keepGoing = t.vm.LeaveNodes(c, t.vm.Visitor(), left, right, keepGoing)

	// For convenience, convert SkipChildren into Continue, so that LeaveNodes
	// implementations can just return keepGoing unchanged.
	if keepGoing == SkipChildren {
		keepGoing = Continue
	}
	return keepGoing
}

func DefaultVisitChildren(newContext PairContext, vm PairVisitorManager, left, right interface{}) Cont {
	if left == nil {
		return Continue
	}
	leftT := reflect.TypeOf(left)
	leftV := reflect.ValueOf(left)

	if right == nil {
		return Continue
	}
	rightT := reflect.TypeOf(right)
	rightV := reflect.ValueOf(right)

	// If we visited a pointer, don't also visit the object; just descend into
	// it. Prune the traversal if left or right is nil.
	for leftT.Kind() == reflect.Ptr {
		if leftV.IsNil() {
			return Continue
		}
		leftT = leftT.Elem()
		leftV = leftV.Elem()
	}
	for rightT.Kind() == reflect.Ptr {
		if rightV.IsNil() {
			return Continue
		}
		rightT = rightT.Elem()
		rightV = rightV.Elem()
	}

	// Don't visit children if left and right have different kinds.
	if leftT.Kind() != rightT.Kind() {
		return Continue
	}

	// Recurse into data structures. Extend the context when visiting children,
	// but not between siblings.
	astv := astPairVisitor{vm: vm}
	switch leftT.Kind() {
	case reflect.Struct:
		return astv.visitStructChildren(newContext, leftT, leftV, rightT, rightV)

	case reflect.Array, reflect.Slice:
		return astv.visitArrayChildren(newContext, leftV, rightV)

	case reflect.Map:
		// Don't visit children if left and right have different key types.
		if leftT.Key() != rightT.Key() {
			return Continue
		}
		return astv.visitMapChildren(newContext, leftV, rightV)

	default:
		return Continue
	}
}

// Helper for visiting the children of structs leftV and rightV having
// respective types leftT and rightT in context ctx.
func (t *astPairVisitor) visitStructChildren(ctx PairContext, leftT reflect.Type, leftV reflect.Value, rightT reflect.Type, rightV reflect.Value) Cont {
	keepGoing := Continue
	namesVisited := make(map[string]struct{})

	// Visit fields on the left.
	for i := 0; i < leftT.NumField(); i++ {
		fieldName := leftT.Field(i).Name
		leftFieldV := leftV.Field(i)

		// Skip private fields and invalid values.
		if !leftFieldV.IsValid() || unicode.IsLower([]rune(fieldName)[0]) {
			continue
		}

		// XXX Skip Protobuf-generated fields, identified by names beginning with
		// "XXX_"
		if strings.HasPrefix(fieldName, "XXX_") {
			continue
		}

		namesVisited[fieldName] = struct{}{}

		rightFieldV := rightV.FieldByName(fieldName)
		if rightFieldV.IsValid() {
			keepGoing = t.visit(ctx.AppendPaths(fieldName, fieldName), leftFieldV.Interface(), rightFieldV.Interface())
		} else {
			keepGoing = t.visit(ctx.AppendPaths(fieldName, fieldName), leftFieldV.Interface(), nil)
		}

		switch keepGoing {
		case Abort, Stop:
			return keepGoing
		case Continue:
		case SkipChildren:
			panic("astPairVisitor.visit returned SkipChildren")
		default:
			panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
		}
	}

	// Visit any unvisited fields on the right.
	for i := 0; i < rightT.NumField(); i++ {
		fieldName := rightT.Field(i).Name
		if _, ok := namesVisited[fieldName]; ok {
			// Field already visited.
			continue
		}

		// Skip private fields and invalid values.
		rightFieldV := rightV.Field(i)
		if !rightFieldV.IsValid() || unicode.IsLower([]rune(fieldName)[0]) {
			continue
		}

		// XXX Skip Protobuf-generated fields, identified by names beginning with
		// "XXX_"
		if strings.HasPrefix(fieldName, "XXX_") {
			continue
		}

		keepGoing = t.visit(ctx.AppendPaths(fieldName, fieldName), nil, rightFieldV.Interface())

		switch keepGoing {
		case Abort, Stop:
			return keepGoing
		case Continue:
		case SkipChildren:
			panic("astPairVisitor.visit returned SkipChildren")
		default:
			panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
		}
	}

	return keepGoing
}

// Helper for visiting the children of arrays leftV and rightV having
// respective types leftT and rightT in context ctx.
func (t *astPairVisitor) visitArrayChildren(ctx PairContext, leftV, rightV reflect.Value) Cont {
	keepGoing := Continue
	for i := 0; i < leftV.Len() || i < rightV.Len(); i++ {
		var leftElt interface{} = nil
		var rightElt interface{} = nil

		if i < leftV.Len() {
			leftElt = leftV.Index(i).Interface()
		}

		if i < rightV.Len() {
			rightElt = rightV.Index(i).Interface()
		}

		idxStr := strconv.Itoa(i)
		keepGoing = t.visit(ctx.AppendPaths(idxStr, idxStr), leftElt, rightElt)
		switch keepGoing {
		case Abort:
			return Abort
		case Continue:
		case SkipChildren:
			panic("astPairVisitor.visit returned SkipChildren")
		case Stop:
			return Stop
		default:
			panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
		}
	}

	return keepGoing
}

// Helper for visiting the children of maps leftV and rightV in context ctx.
// Assumes that the two maps have the same key type.
func (t *astPairVisitor) visitMapChildren(ctx PairContext, leftV, rightV reflect.Value) Cont {
	// TODO(cs): Need to visit (k,v), then k, then v for each k, v.
	keepGoing := Continue

	// Build a set of keys for each map.
	leftKeys := make(map[interface{}]struct{})
	for _, k := range leftV.MapKeys() {
		leftKeys[k.Interface()] = struct{}{}
	}
	rightKeys := make(map[interface{}]struct{})
	for _, k := range rightV.MapKeys() {
		rightKeys[k.Interface()] = struct{}{}
	}

	// Visit values on left.
	for _, k := range leftV.MapKeys() {
		leftElt := leftV.MapIndex(k).Interface()

		var rightElt interface{} = nil
		if _, ok := rightKeys[k.Interface()]; ok {
			rightElt = rightV.MapIndex(k).Interface()
		}

		pathStr := fmt.Sprint(k.Interface())
		keepGoing = t.visit(ctx.AppendPaths(pathStr, pathStr), leftElt, rightElt)
		switch keepGoing {
		case Abort:
			return Abort
		case Continue:
		case SkipChildren:
			panic("astPairVisitor.visit returned SkipChildren")
		case Stop:
			return Stop
		default:
			panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
		}
	}

	// Visit any unvisited values on right.
	for _, k := range rightV.MapKeys() {
		if _, ok := leftKeys[k.Interface()]; ok {
			// Already visited.
			continue
		}

		rightElt := rightV.MapIndex(k).Interface()
		pathStr := fmt.Sprint(k.Interface())
		keepGoing = t.visit(ctx.AppendPaths(pathStr, pathStr), nil, rightElt)
		switch keepGoing {
		case Abort:
			return Abort
		case Continue:
		case SkipChildren:
			panic("astPairVisitor.visit returned SkipChildren")
		case Stop:
			return Stop
		default:
			panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
		}
	}

	return keepGoing
}
