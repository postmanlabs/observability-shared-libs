package visitors

type PairContext interface {
	// Returns a new Context after appending to the paths.
	AppendPaths(left string, right string) PairContext

	// Returns the paths through the structures being traversed.
	// See Context.GetPath().
	GetPaths() ([]string, []string)
}

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

	ExtendContext(c PairContext, leftNode, rightNode interface{}) PairContext
}

func NewPairVisitorManager(
	c PairContext,
	v interface{},
	enter func(c PairContext, visitor interface{}, left, right interface{}) Cont,
	visitChildren func(c PairContext, vm PairVisitorManager, left, right interface{}) Cont,
	leave func(c PairContext, visitor interface{}, left, right interface{}, cont Cont) Cont,
	extendContext func(c PairContext, left, right interface{}) PairContext,
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

func (c *pairContext) AppendPaths(left string, right string) PairContext {
	return &pairContext{
		left:  c.left.AppendPath(left),
		right: c.right.AppendPath(right),
	}
}

func (c *pairContext) GetPaths() ([]string, []string) {
	return c.left.GetPath(), c.right.GetPath()
}

type pairVisitor struct {
	context       PairContext
	visitor       interface{}
	enter         func(c PairContext, visitor interface{}, left, right interface{}) Cont
	visitChildren func(c PairContext, vm PairVisitorManager, left, right interface{}) Cont
	leave         func(c PairContext, visitor interface{}, left, right interface{}, cont Cont) Cont
	extendContext func(c PairContext, left, right interface{}) PairContext
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

func (v *pairVisitor) ExtendContext(c PairContext, left, right interface{}) PairContext {
	return v.extendContext(c, left, right)
}
