// Package go_ast implements a depth-first traversal of a go data structure,
// applying the vm to each node.
package go_ast

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode"

	. "github.com/akitasoftware/akita-libs/visitors"
)

// Structurally recurses through `a`.  At each term, invokes
//
//   c, keepGoing := v.Apply(c, v.Visitor(), t)
//
// Aborts the traversal if v.Apply returns false.
//
func Apply(v VisitorManager, a interface{}) Cont {
	astv := astVisitor{vm: v}
	return astv.visit(v.Context(), a)
}

type astVisitor struct {
	vm VisitorManager
}

// Visits the node m, whose context is c. This should never return SkipChildren.
func (t *astVisitor) visit(c Context, m interface{}) Cont {
	if m == nil {
		return Continue
	}

	keepGoing := t.vm.EnterNode(c, t.vm.Visitor(), m)
	switch keepGoing {
	case Abort:
		return Abort
	case Continue:
	case SkipChildren:
	case Stop:
	default:
		panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
	}

	newContext := t.vm.ExtendContext(c, t.vm.Visitor(), m)

	// Don't visit children if we are stopping or skipping children.
	if keepGoing == Continue {
		// Traverse m's children.
		mt := reflect.TypeOf(m)
		mv := reflect.ValueOf(m)

		// If we visited a pointer, don't also visit the object; just descend into
		// it.
		for mt.Kind() == reflect.Ptr {
			if mv.IsNil() {
				return Continue
			}
			mt = mt.Elem()
			mv = mv.Elem()
		}

		// Recurse into data structures.  Extend the context when visiting
		// children, but not between siblings.
		if mt.Kind() == reflect.Struct {
			keepGoing = t.visitStructChildren(newContext, mt, mv, keepGoing)
		} else if mt.Kind() == reflect.Array || mt.Kind() == reflect.Slice {
			keepGoing = t.visitArrayChildren(newContext, mv, keepGoing)
		} else if mt.Kind() == reflect.Map {
			keepGoing = t.visitMapChildren(newContext, mv, keepGoing)
		}
	}

	keepGoing = t.vm.LeaveNode(c, t.vm.Visitor(), m, keepGoing)

	// For convenience, convert SkipChildren into Continue, so that LeaveNode
	// implementations can just return keepGoing unchanged.
	if keepGoing == SkipChildren {
		keepGoing = Continue
	}
	return keepGoing
}

// Helper for visiting the children of a struct mv having type mt in context
// ctx.
func (t *astVisitor) visitStructChildren(ctx Context, mt reflect.Type, mv reflect.Value, keepGoing Cont) Cont {
	for i := 0; i < mt.NumField(); i++ {
		ft := mt.Field(i)
		fv := mv.Field(i)

		// Skip private fields and invalid values.
		if !fv.IsValid() || unicode.IsLower([]rune(ft.Name)[0]) {
			continue
		}

		keepGoing = t.visit(ctx.AppendPath(ft.Name), mv.Field(i).Interface())

		switch keepGoing {
		case Abort, Stop:
			return keepGoing
		case Continue:
		case SkipChildren:
			panic(fmt.Sprintf("astVisitor.visit returned SkipChildren"))
		default:
			panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
		}
	}

	return keepGoing
}

// Helper for visiting the children of an array mv in context ctx.
func (t *astVisitor) visitArrayChildren(ctx Context, mv reflect.Value, keepGoing Cont) Cont {
	for i := 0; i < mv.Len(); i++ {
		keepGoing = t.visit(ctx.AppendPath(strconv.Itoa(i)), mv.Index(i).Interface())
		switch keepGoing {
		case Abort:
			return Abort
		case Continue:
		case SkipChildren:
			panic(fmt.Sprintf("astVisitor.visit returned SkipChildren"))
		case Stop:
			break
		default:
			panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
		}
	}

	return keepGoing
}

// Helper for visiting the children of a map mv in context ctx.
func (t *astVisitor) visitMapChildren(ctx Context, mv reflect.Value, keepGoing Cont) Cont {
	// TODO(cs): Need to visit (k,v), then k, then v for each k, v.
	for _, k := range mv.MapKeys() {
		keepGoing = t.visit(ctx.AppendPath(fmt.Sprint(k.Interface())), mv.MapIndex(k).Interface())
		switch keepGoing {
		case Abort:
			return Abort
		case Continue:
		case SkipChildren:
			panic(fmt.Sprintf("astVisitor.visit returned SkipChildren"))
		case Stop:
			break
		default:
			panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
		}
	}

	return keepGoing
}
