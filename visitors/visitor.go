// This package defines a generic VisitorManager interface, which is implemented
// in various sub-packages, like http_rest_spec_visitor for traversing
// REST specs.
//
// The go_ast subpackage is the infrastructure that implements the
// visitor traversal.  It takes a VisitorManager and applies it to a Go data
// structure, calling VisitorManager.Visit on each subterm in the data structure.
//
// See http_rest_spec_visitor/http_rest_spec_visitor_test.go for an example
// of using a REST spec visitor.
package visitors

import (
	"fmt"
	"strconv"
)

type Context interface {
	// Returns a new Context in which the path is appended to indicate a
	// traversal into the given field of the given struct.
	EnterStruct(structNode interface{}, fieldName string) Context

	// Returns a new Context in which the path is appended to indicate a
	// traversal into the given element of the given array.
	EnterArray(arrayNode interface{}, elementIndex int) Context

	// Returns a new Context in which the path is appended to indicate a
	// traversal into the value at the given key of the given map.
	EnterMapValue(mapNode, mapKey interface{}) Context

	// Returns the path through the structure being traversed.
	GetPath() ContextPath

	// Returns the parent context from which this context was derived. Returns
	// nil if this is the root context.
	GetOuter() Context
}

// Represents a path through a data structure.
type ContextPath []ContextPathElement

func (c ContextPath) IsEmpty() bool {
	return len(c) == 0
}

func (c ContextPath) GetLast() ContextPathElement {
	return c[len(c)-1]
}

// Represents an ancestor node and an outgoing edge from that node.
type ContextPathElement struct {
	AncestorNode interface{}
	OutEdge      ContextPathEdge
}

// Represents an edge in the path from the root node to the node being visited.
type ContextPathEdge interface {
	isContextPathEdge()
	String() string
}

// A context-path edge indicating that a field of a struct is being visited.
type StructFieldEdge struct {
	FieldName string
}

var _ ContextPathEdge = (*StructFieldEdge)(nil)

func NewStructFieldEdge(fieldName string) *StructFieldEdge {
	return &StructFieldEdge{
		FieldName: fieldName,
	}
}

func (*StructFieldEdge) isContextPathEdge() {}
func (e *StructFieldEdge) String() string {
	return e.FieldName
}

// A context-path edge indicating that an element of an array is being visited.
type ArrayElementEdge struct {
	ElementIndex int
}

var _ ContextPathEdge = (*StructFieldEdge)(nil)

func NewArrayElementEdge(elementIndex int) *ArrayElementEdge {
	return &ArrayElementEdge{
		ElementIndex: elementIndex,
	}
}

func (*ArrayElementEdge) isContextPathEdge() {}
func (e *ArrayElementEdge) String() string {
	return strconv.Itoa(e.ElementIndex)
}

// A context-path edge indicating that a value of a map is being visited.
type MapValueEdge struct {
	MapKey interface{}
}

var _ ContextPathEdge = (*MapValueEdge)(nil)

func NewMapValueEdge(mapKey interface{}) *MapValueEdge {
	return &MapValueEdge{
		MapKey: mapKey,
	}
}

func (*MapValueEdge) isContextPathEdge() {}
func (e *MapValueEdge) String() string {
	return fmt.Sprint(e.MapKey)
}

func NewContext() Context {
	return new(context)
}

// An enum to indicate how a visitor's traversal should continue.
type Cont int

const (
	// Indicates that the visitor should continue with normal traversal.
	Continue Cont = iota

	// Indicates that the visitor should end its traversal, but perform 'leave'
	// operations as the traversal stack is unwound back to the root node.
	Stop

	// Indicates that the visitor should stop its traversal immediately. No
	// 'leave' operations will be performed as the traversal stack is unwound
	// back to the root node.
	Abort

	// Indicates that the visitor should not visit the children of the current
	// node.
	SkipChildren
)

// A visitor is made up of a context (which may extend the Context defined
// above), an arbitrary visitor object, and an Apply function that takes the
// context, the visitor object, and a term to visit.  It returns the context
// passed in (possibly with modifications) as well as a value indicating how to
// continue the traversal.
//
// Typically, the Apply method will use the context and the term to figure
// out which method of the visitor object to call, and then call it with
// the context and the term.
//
// Factoring out the dispatch logic into Apply means the logic for figuring
// out which visitor method to call is implemented once for a given data
// structure, and custom visitors for that data structure can simply override
// the methods on the vistor object they care about.  See
// http_rest_spec_visitor for an example.
//
// ExtendContext creates a new context for a given term without visiting the
// term.  This makes it possible to create the correct context for children
// when applying a postorder traversal.
type VisitorManager interface {
	Context() Context
	Visitor() interface{}

	// Provides functionality for entering a node, before visiting the node's
	// children.
	EnterNode(c Context, visitor interface{}, node interface{}) Cont

	// Visits a node's children with the given context.
	VisitChildren(c Context, vm VisitorManager, node interface{}) Cont

	// Provides functionality for leaving a node, after visiting the node's
	// children.
	//
	// The given cont can be Continue, Stop, or SkipChildren, indicating the
	// state of the traversal before leaving the node. Most implementations will
	// want to return this value unchanged: for convenience, if SkipChildren is
	// returned, the visitor framework will interpret this as Continue.
	LeaveNode(c Context, visitor interface{}, node interface{}, cont Cont) Cont

	// Augments the given context with information from the given node.
	ExtendContext(c Context, visitor interface{}, term interface{})
}

func NewVisitorManager(
	c Context,
	v interface{},
	enter func(c Context, visitor interface{}, term interface{}) Cont,
	visitChildren func(c Context, vm VisitorManager, term interface{}) Cont,
	leave func(c Context, visitor interface{}, term interface{}, cont Cont) Cont,
	extendContext func(c Context, term interface{}),
) VisitorManager {
	rv := visitor{
		context:       c,
		visitor:       v,
		enter:         enter,
		visitChildren: visitChildren,
		leave:         leave,
		extendContext: extendContext,
	}
	return &rv
}

type context struct {
	path  ContextPath
	outer Context
}

var _ Context = (*context)(nil)

func (c *context) EnterStruct(structNode interface{}, fieldName string) Context {
	return c.appendPath(ContextPathElement{
		AncestorNode: structNode,
		OutEdge:      &StructFieldEdge{FieldName: fieldName},
	})
}

func (c *context) EnterArray(arrayNode interface{}, elementIndex int) Context {
	return c.appendPath(ContextPathElement{
		AncestorNode: arrayNode,
		OutEdge:      &ArrayElementEdge{ElementIndex: elementIndex},
	})
}

func (c *context) EnterMapValue(mapNode, mapKey interface{}) Context {
	return c.appendPath(ContextPathElement{
		AncestorNode: mapNode,
		OutEdge:      &MapValueEdge{MapKey: mapKey},
	})
}

func (c *context) appendPath(e ContextPathElement) Context {
	return &context{
		path:  append(c.path, e),
		outer: c,
	}
}

func (c *context) GetPath() ContextPath {
	return c.path
}

func (c *context) GetOuter() Context {
	return c.outer
}

type visitor struct {
	context       Context
	visitor       interface{}
	enter         func(c Context, visitor interface{}, term interface{}) Cont
	visitChildren func(c Context, vm VisitorManager, term interface{}) Cont
	leave         func(c Context, visitor interface{}, term interface{}, cont Cont) Cont
	extendContext func(c Context, term interface{})
}

var _ VisitorManager = (*visitor)(nil)

func (v *visitor) Context() Context {
	return v.context
}

func (v *visitor) Visitor() interface{} {
	return v.visitor
}

func (v *visitor) EnterNode(c Context, visitor interface{}, term interface{}) Cont {
	return v.enter(c, visitor, term)
}

func (v *visitor) VisitChildren(c Context, vm VisitorManager, term interface{}) Cont {
	return v.visitChildren(c, vm, term)
}

func (v *visitor) LeaveNode(c Context, visitor interface{}, term interface{}, cont Cont) Cont {
	return v.leave(c, visitor, term, cont)
}

func (v *visitor) ExtendContext(c Context, visitor interface{}, term interface{}) {
	v.extendContext(c, term)
}
