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

type Context interface {
	// Returns a new Context with s appended to the path.
	AppendPath(s string) Context

	// Returns the path through the structure being traversed, including
	// the indexes in lists and the keys in maps being traversed.
	//
	// For example, the path when visiting the second input argument to
	// the first parameter of an APISpec would be:
	//
	//   Methods 0 Args arg-headers-0 Value Primitive
	//   ^       ^ ^    ^             ^     ^
	//   |       | |    |             |     \-- "primitive" field of Data object value being visited
	//   |       | |    |             \-- "value" oneof field of the Data object
	//   |       | |    \-- key of the args map
	//   |       | \-- "args" field of the Method object
	//   |       \-- first element of the APISpec.methods list
	//   \-- "methods" field of APISpec
	//
	GetPath() []string
}

func NewContext() Context {
	return new(context)
}

// A visitor is made up of a context (which may extend the Context defined
// above), an arbitrary visitor object, and an Apply function that takes the
// context, the visitor object, and a term to visit.  It returns the context
// passed in (possibly with modifications) as well as a boolean value
// indicating whether to continue the traversal (false means stop).
//
// Typically, the Apply method will use the context and the term to figure
// out which method of the visitor object to call, and then call it with
// the context and the term.
//
// Factoring out the dispatch logic into Apply means the logic for figuring
// out which visitor method to call is implemented once for a given data
// structure, and custom visitors for that data structure can simply overload
// the methods on the vistor object they care about.  See
// http_rest_spec_visitor for an example.
//
// ExtendContext creates a new context for a given term without visiting the
// term.  This makes it possible to create the correct context for children
// when applying a postorder traversal.
type VisitorManager interface {
	Context() Context
	Visitor() interface{}
	Apply(c Context, visitor interface{}, term interface{}) bool
	ExtendContext(c Context, visitor interface{}, term interface{}) Context
}

func NewVisitorManager(
	c Context,
	v interface{},
	apply func(c Context, visitor interface{}, term interface{}) bool,
	extendContext func(c Context, visitor interface{}, term interface{}) Context,
) VisitorManager {
	rv := visitor{
		context:       c,
		visitor:       v,
		apply:         apply,
		extendContext: extendContext,
	}
	return &rv
}

type context struct {
	path []string
}

func (c *context) AppendPath(s string) Context {
	return &context{path: append(c.path, s)}
}

func (c *context) GetPath() []string {
	return c.path
}

type visitor struct {
	context       Context
	visitor       interface{}
	apply         func(c Context, visitor interface{}, term interface{}) bool
	extendContext func(c Context, visitor interface{}, term interface{}) Context
}

func (v *visitor) Context() Context {
	return v.context
}

func (v *visitor) Visitor() interface{} {
	return v.visitor
}

func (v *visitor) Apply(c Context, visitor interface{}, term interface{}) bool {
	return v.apply(c, visitor, term)
}

func (v *visitor) ExtendContext(c Context, visitor interface{}, term interface{}) Context {
	return v.extendContext(c, visitor, term)
}
