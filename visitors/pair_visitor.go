package visitors

type PairContext interface {
	// Returns a new PairContext in which the paths are appended to indicate a
	// traversal into the given fields of the given structs.
	EnterStructs(leftStruct interface{}, leftFieldName string, rightStruct interface{}, rightFieldName string) PairContext

	// Returns a new PairContext in which the path on one side is appended to
	// indicate a traversal into the given field of the given struct.
	EnterStruct(leftOrRight LeftOrRight, structNode interface{}, fieldName string) PairContext

	// Returns a new PairContext in which the paths are appended to indicate a
	// traversal into the given elements of the given arrays.
	EnterArrays(leftArray interface{}, leftIndex int, rightArray interface{}, rightIndex int) PairContext

	// Returns a new PairContext in which the path on one side is appended to
	// indicate a traversal into the given element of the given array.
	EnterArray(leftOrRight LeftOrRight, arrayNode interface{}, elementIndex int) PairContext

	// Returns a new PairContext in which the paths are appended to indicate a
	// traversal into the values at the given keys of the given maps.
	EnterMapValues(leftMap, leftKey, rightMap, rightKey interface{}) PairContext

	// Returns a new PairContext in which the path on one side is appended to
	// indicate a traversal into the value at the given key of the given map.
	EnterMapValue(leftOrRight LeftOrRight, mapNode, mapKey interface{}) PairContext

	// Returns the paths through the structures being traversed.
	// See Context.GetPath().
	GetPaths() (ContextPath, ContextPath)
}

type LeftOrRight bool

const (
	LeftSide  LeftOrRight = true
	RightSide LeftOrRight = false
)

func NewPairContext() PairContext {
	return &pairContext{
		left:  NewContext(),
		right: NewContext(),
	}
}

// A version of VisitorManager that traverses a pair of nodes in tandem.
type PairVisitorManager interface {
	// Creates an empty context.
	Context() PairContext

	// Returns the visitor implementation.
	Visitor() interface{}

	// Provides functionality for entering a pair of nodes, before visiting the
	// nodes' children.
	EnterNodes(c PairContext, visitor interface{}, leftNode, rightNode interface{}) Cont

	// Visits the nodes' children with the given context.
	VisitChildren(c PairContext, vm PairVisitorManager, leftNode, rightNode interface{}) Cont

	// Provides functionality for leaving a pair of nodes, after visiting the
	// nodes' children.
	//
	// The given cont can be Continue, Stop, or SkipChildren, indicating the
	// state of the traversal before leaving the node. Most implementations will
	// want to return this value unchanged: for convenience, if SkipChildren is
	// returned, the visitor framework will interpret this as Continue.
	LeaveNodes(c PairContext, visitor interface{}, leftNode, rightNode interface{}, cont Cont) Cont

	ExtendContext(c PairContext, leftNode, rightNode interface{})
}

func NewPairVisitorManager(
	c PairContext,
	v interface{},
	enter func(c PairContext, visitor interface{}, left, right interface{}) Cont,
	visitChildren func(c PairContext, vm PairVisitorManager, left, right interface{}) Cont,
	leave func(c PairContext, visitor interface{}, left, right interface{}, cont Cont) Cont,
	extendContext func(c PairContext, left, right interface{}),
) PairVisitorManager {
	rv := pairVisitor{
		context:       c,
		visitor:       v,
		enter:         enter,
		visitChildren: visitChildren,
		leave:         leave,
		extendContext: extendContext,
	}
	return &rv
}

type pairContext struct {
	left  Context
	right Context
}

var _ PairContext = (*pairContext)(nil)

func (c *pairContext) EnterStructs(leftStruct interface{}, leftFieldName string, rightStruct interface{}, rightFieldName string) PairContext {
	return &pairContext{
		left:  c.left.EnterStruct(leftStruct, leftFieldName),
		right: c.right.EnterStruct(rightStruct, rightFieldName),
	}
}

func (c *pairContext) EnterStruct(leftOrRight LeftOrRight, structNode interface{}, fieldName string) PairContext {
	if leftOrRight == LeftSide {
		return &pairContext{
			left:  c.left.EnterStruct(structNode, fieldName),
			right: c.right,
		}
	}

	return &pairContext{
		left:  c.left,
		right: c.right.EnterStruct(structNode, fieldName),
	}
}

// Returns a new PairContext in which the paths are appended to indicate a
// traversal into the given elements of the given arrays.
func (c *pairContext) EnterArrays(leftArray interface{}, leftIndex int, rightArray interface{}, rightIndex int) PairContext {
	return &pairContext{
		left:  c.left.EnterArray(leftArray, leftIndex),
		right: c.right.EnterArray(rightArray, rightIndex),
	}
}

// Returns a new PairContext in which the path on one side is appended to
// indicate a traversal into the given element of the given array.
func (c *pairContext) EnterArray(leftOrRight LeftOrRight, arrayNode interface{}, elementIndex int) PairContext {
	if leftOrRight == LeftSide {
		return &pairContext{
			left:  c.left.EnterArray(arrayNode, elementIndex),
			right: c.right,
		}
	}

	return &pairContext{
		left:  c.left,
		right: c.right.EnterArray(arrayNode, elementIndex),
	}
}

// Returns a new PairContext in which the paths are appended to indicate a
// traversal into the values at the given keys of the given maps.
func (c *pairContext) EnterMapValues(leftMap, leftKey, rightMap, rightKey interface{}) PairContext {
	return &pairContext{
		left:  c.left.EnterMapValue(leftMap, leftKey),
		right: c.right.EnterMapValue(rightMap, rightKey),
	}
}

// Returns a new PairContext in which the path on one side is appended to
// indicate a traversal into the value at the given key of the given map.
func (c *pairContext) EnterMapValue(leftOrRight LeftOrRight, mapNode, mapKey interface{}) PairContext {
	if leftOrRight == LeftSide {
		return &pairContext{
			left:  c.left.EnterMapValue(mapNode, mapKey),
			right: c.right,
		}
	}
	return &pairContext{
		left:  c.left,
		right: c.right.EnterMapValue(mapNode, mapKey),
	}
}

func (c *pairContext) GetPaths() (ContextPath, ContextPath) {
	return c.left.GetPath(), c.right.GetPath()
}

type pairVisitor struct {
	context       PairContext
	visitor       interface{}
	enter         func(c PairContext, visitor interface{}, left, right interface{}) Cont
	visitChildren func(c PairContext, vm PairVisitorManager, left, right interface{}) Cont
	leave         func(c PairContext, visitor interface{}, left, right interface{}, cont Cont) Cont
	extendContext func(c PairContext, left, right interface{})
}

func (v *pairVisitor) Context() PairContext {
	return v.context
}

func (v *pairVisitor) Visitor() interface{} {
	return v.visitor
}

func (v *pairVisitor) EnterNodes(c PairContext, visitor interface{}, left, right interface{}) Cont {
	return v.enter(c, visitor, left, right)
}

func (v *pairVisitor) VisitChildren(c PairContext, vm PairVisitorManager, left, right interface{}) Cont {
	return v.visitChildren(c, vm, left, right)
}

func (v *pairVisitor) LeaveNodes(c PairContext, visitor interface{}, left, right interface{}, cont Cont) Cont {
	return v.leave(c, visitor, left, right, cont)
}

func (v *pairVisitor) ExtendContext(c PairContext, left, right interface{}) {
	v.extendContext(c, left, right)
}
