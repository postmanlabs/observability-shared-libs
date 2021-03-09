// Package go_ast implements a depth-first traversal of a go data structure,
// applying the vm to each node.
package go_ast

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode"

	"github.com/akitasoftware/akita-libs/visitors"
)

type TraversalOrder int

const (
	PREORDER = iota
	POSTORDER
)

func (t TraversalOrder) String() string {
	return []string{"PREORDER", "POSTORDER"}[t]
}

// Structurally recurses through `a`.  At each term, invokes
//
//   c, keepGoing := v.Apply(c, v.Visitor(), t)
//
// Aborts the traversal if v.Apply returns false.
//
func Apply(order TraversalOrder, v visitors.VisitorManager, a interface{}) bool {
	astv := astVisitor{order: order, vm: v}
	return astv.visit(v.Context(), a)
}

type astVisitor struct {
	order TraversalOrder
	vm    visitors.VisitorManager
}

func (t *astVisitor) visit(c visitors.Context, m interface{}) bool {
	if m == nil {
		return true
	}
	keepGoing := true

	if t.order == PREORDER {
		keepGoing = t.vm.Apply(c, t.vm.Visitor(), m)
		if !keepGoing {
			return false
		}
	}

	newContext := t.vm.ExtendContext(c, t.vm.Visitor(), m)

	// Traverse m's children
	mt := reflect.TypeOf(m)
	mv := reflect.ValueOf(m)

	// If we visited a pointer, don't also visit the object; just descend into it
	for mt.Kind() == reflect.Ptr {
		if mv.IsNil() {
			return true
		}
		mt = mt.Elem()
		mv = mv.Elem()
	}

	// Recurse into data structures.  Extend the context when visiting
	// children, but not between siblings.
	if mt.Kind() == reflect.Struct {
		for i := 0; i < mt.NumField(); i++ {
			ft := mt.Field(i)
			fv := mv.Field(i)
			// Skip private fields and invalid values.
			if !fv.IsValid() || unicode.IsLower([]rune(ft.Name)[0]) {
				continue
			}
			keepGoing = t.visit(newContext.AppendPath(ft.Name), mv.Field(i).Interface())
		}
	} else if mt.Kind() == reflect.Array || mt.Kind() == reflect.Slice {
		for i := 0; i < mv.Len(); i++ {
			keepGoing = t.visit(newContext.AppendPath(strconv.Itoa(i)), mv.Index(i).Interface())
		}
	} else if mt.Kind() == reflect.Map {
		// TODO(cs): Need to visit (k,v), then k, then v for each k, v.
		for _, k := range mv.MapKeys() {
			keepGoing = t.visit(newContext.AppendPath(fmt.Sprint(k.Interface())), mv.MapIndex(k).Interface())
		}
	}

	if t.order == POSTORDER && keepGoing {
		keepGoing = t.vm.Apply(c, t.vm.Visitor(), m)
	}

	return keepGoing
}
