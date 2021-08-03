package http_rest

import (
	"reflect"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/visitors"
)

// Describes which part of an HTTP request/response a value belongs to.
type HttpValueType int

const (
	UNKNOWN HttpValueType = iota
	PATH
	QUERY
	HEADER
	COOKIE
	BODY
	AUTH
)

func (t HttpValueType) String() string {
	return []string{"Unknown", "Path", "Query", "Header", "Cookie", "Body", "Auth"}[t]
}

type SpecVisitorContext interface {
	visitors.Context

	// Identifies the REST field being visited when combined with IsArg,
	// IsResponse, and GetValueType. For fields outside of bodies and
	// authorization headers, the first component of the path is a FieldName
	// identifying the path argument, query variable, header, or cookie in which
	// the field is located.
	GetFieldPath() []FieldPathElement

	// Gets the REST path, which is similar to the AST path but using names
	// from DataMeta and MethodMeta objects where appropriate, as well as
	// hiding names of container data structures.
	//
	// For example, the AST path to a path parameter might be:
	//
	//   Methods 0 Args arg-headers-0 Value Primitive
	//
	// These are the names of the fields in the AST.  In contrast, the REST
	// path might be:
	//
	//   localhost:9000 GET /api/0/issues/{issue_id}/events/ Header Authorization
	//
	GetRestPath() []string

	// Returns the REST operation name of the method being traversed, if the
	// visitor is visiting a method or its descendent nodes.
	GetRestOperation() string
	setRestOperation(string)

	// Returns true if the message is a descendent of Method.Args.
	IsArg() bool

	// Returns true if the message is a descendent of Method.Responses.
	IsResponse() bool

	// Returns true if the message is a descendant of (or is itself) a Data
	// instance representing an optional value.
	IsOptional() bool

	GetValueType() HttpValueType

	GetHttpAuthType() *pb.HTTPAuth_HTTPAuthType

	// Like GetRestPath, except only including the portion about the argument
	// value.
	GetArgPath() []string

	// Like GetRestPath, except only including the portion about the argument
	// value.
	GetResponsePath() []string

	// Returns the endpoint path, e.g. the "/v1/users" part of
	// "localhost GET /v1/users".
	GetEndpointPath() string

	// This is nil if the message is not a descendant of Method.Responses.
	GetResponseCode() *string

	// This is nil if the message is not part of a body.
	GetContentType() *string

	// Returns the host.
	GetHost() string

	// Returns the innermost node being visited having the given type, and that
	// node's context.
	GetInnermostNode(reflect.Type) (interface{}, SpecVisitorContext)

	// Used by the visitor infrastructure to maintain the location of the field
	// being visited.
	appendFieldPath(FieldPathElement)

	// Used by the visitor infrastructure to construct a REST path by replacing
	// parts of the AST path with names from DataMeta.
	appendRestPath(string)

	setIsArg(bool)
	setIsOptional()
	setValueType(HttpValueType)
	setHttpAuthType(pb.HTTPAuth_HTTPAuthType)
	setTopLevelDataIndex(int)
	setResponseCode(string)
	setContentType(string)
}

type specVisitorContext struct {
	path            visitors.ContextPath
	lastPathElement visitors.ContextPathElement

	outer SpecVisitorContext

	fieldPath []FieldPathElement
	restPath  []string

	// nil means we're not sure if this is an arg or response value.
	isArg *bool

	isOptional bool

	valueType    HttpValueType
	httpAuthType *pb.HTTPAuth_HTTPAuthType

	responseCode *string
	contentType  *string

	// Index within restPath of the start of the section describing arg or
	// response.
	topDataIndex int

	restOperation string
}

var _ SpecVisitorContext = (*specVisitorContext)(nil)

func (c *specVisitorContext) EnterStruct(structNode interface{}, fieldName string) visitors.Context {
	return c.appendPath(visitors.ContextPathElement{
		AncestorNode: structNode,
		OutEdge:      visitors.NewStructFieldEdge(fieldName),
	})
}

func (c *specVisitorContext) EnterArray(arrayNode interface{}, elementIndex int) visitors.Context {
	return c.appendPath(visitors.ContextPathElement{
		AncestorNode: arrayNode,
		OutEdge:      visitors.NewArrayElementEdge(elementIndex),
	})
}

func (c *specVisitorContext) EnterMapValue(mapNode, mapKey interface{}) visitors.Context {
	return c.appendPath(visitors.ContextPathElement{
		AncestorNode: mapNode,
		OutEdge:      visitors.NewMapValueEdge(mapKey),
	})
}

func (c *specVisitorContext) appendPath(e visitors.ContextPathElement) *specVisitorContext {
	// TODO: can we also lazily create the path here?
	result := *c
	result.path = append(result.path, e)
	result.outer = c
	return &result
}

func (c *specVisitorContext) GetPath() visitors.ContextPath {
	return c.path
}

func (c *specVisitorContext) GetOuter() visitors.Context {
	return c.outer
}

func (c *specVisitorContext) appendFieldPath(elt FieldPathElement) {
	c.fieldPath = append(c.fieldPath, elt)
}

func (c *specVisitorContext) GetFieldPath() []FieldPathElement {
	return c.fieldPath
}

func (c *specVisitorContext) appendRestPath(s string) {
	c.restPath = append(c.GetRestPath(), s)
}

func (c *specVisitorContext) GetRestPath() []string {
	return c.restPath
}

func (c *specVisitorContext) GetRestOperation() string {
	return c.restOperation
}

func (c *specVisitorContext) setRestOperation(op string) {
	c.restOperation = op
}

func (c *specVisitorContext) IsArg() bool {
	if c.isArg != nil {
		return *c.isArg
	}
	return false
}

func (c *specVisitorContext) IsResponse() bool {
	if c.isArg != nil {
		return !*c.isArg
	}
	return false
}

func (c *specVisitorContext) IsOptional() bool {
	return c.isOptional
}

func (c *specVisitorContext) GetValueType() HttpValueType {
	return c.valueType
}

func (c *specVisitorContext) GetHttpAuthType() *pb.HTTPAuth_HTTPAuthType {
	return c.httpAuthType
}

func (c *specVisitorContext) GetArgPath() []string {
	if !c.IsArg() {
		return nil
	}
	return c.getTopLevelDataPath()
}

func (c *specVisitorContext) GetResponsePath() []string {
	if !c.IsResponse() {
		return nil
	}
	return c.getTopLevelDataPath()
}

func (c *specVisitorContext) GetEndpointPath() string {
	if len(c.restPath) < 3 {
		return ""
	}
	return c.restPath[2]
}

func (c *specVisitorContext) GetResponseCode() *string {
	if !c.IsResponse() {
		return nil
	}
	return c.responseCode
}

func (c *specVisitorContext) GetContentType() *string {
	if c.GetValueType() != BODY {
		return nil
	}
	return c.contentType
}

func (c *specVisitorContext) GetHost() string {
	if len(c.restPath) < 1 {
		return ""
	}
	return c.restPath[0]
}

func (c *specVisitorContext) GetInnermostNode(typ reflect.Type) (interface{}, SpecVisitorContext) {
	for c != nil {
		node := c.path.GetLast().AncestorNode
		if reflect.TypeOf(node) == typ {
			return node, c.outer
		}
		c = c.outer.(*specVisitorContext)
	}

	return nil, nil
}

func (c *specVisitorContext) getTopLevelDataPath() []string {
	p := c.GetRestPath()
	if len(p) > c.topDataIndex+1 {
		return p[c.topDataIndex+1:]
	}
	return nil
}

func (c *specVisitorContext) setIsArg(isArg bool) {
	c.isArg = &isArg
}

func (c *specVisitorContext) setIsOptional() {
	c.isOptional = true
}

func (c *specVisitorContext) setValueType(vt HttpValueType) {
	c.valueType = vt
}

func (c *specVisitorContext) setResponseCode(code string) {
	c.responseCode = &code
}

func (c *specVisitorContext) setContentType(contentType string) {
	c.contentType = &contentType
}

func (c *specVisitorContext) setHttpAuthType(at pb.HTTPAuth_HTTPAuthType) {
	c.httpAuthType = &at
}

func (c *specVisitorContext) setTopLevelDataIndex(i int) {
	c.topDataIndex = i
}

// Preallocate a stack of specVisitorContexts and re-use them
// for a single visit.
//
// This optimization + lazily constructing the ContextPath takes
// appendPath from >25% of memory allocations to <5%.  Though
// both do still show up high in the profile.
type contextStack struct {
	Stack  []specVisitorContext
	Latest int
}

func NewPreallocatedVisitorContext() SpecVisitorContext {
	// TODO: can we make this self-tuning? Remember the 90th percentile depth
	// and preallocate that?
	cs := &contextStack{
		Stack:  make([]specVisitorContext, 10, 50),
		Latest: 0,
	}
	return stackVisitorContext{
		Preallocated: cs,
		Position:     0,
	}
}

type stackVisitorContext struct {
	Preallocated *contextStack
	Position     int
}

// Allocate a new context from ths stack.
// No explicit Pop(), but we can track whether an old context
// is re-used erroneously.
func (s *contextStack) Push(parent int) int {
	if parent > s.Latest {
		panic("push context from too deep in stack")
	}
	next := parent + 1
	if next == len(s.Stack) {
		s.Stack = append(s.Stack, specVisitorContext{})
	} else if next > len(s.Stack) {
		panic("push increases size by more than 1")
	}
	s.Latest = next
	return next
}

// This pointer is only valid until the slice is resized, so it should be used
// and then immediately discarded.
func (c stackVisitorContext) Delegate() *specVisitorContext {
	return &c.Preallocated.Stack[c.Position]
}

func (c stackVisitorContext) EnterStruct(structNode interface{}, fieldName string) visitors.Context {
	return c.appendPath(visitors.ContextPathElement{
		AncestorNode: structNode,
		OutEdge:      visitors.NewStructFieldEdge(fieldName),
	})
}

func (c stackVisitorContext) EnterArray(arrayNode interface{}, elementIndex int) visitors.Context {
	return c.appendPath(visitors.ContextPathElement{
		AncestorNode: arrayNode,
		OutEdge:      visitors.NewArrayElementEdge(elementIndex),
	})
}

func (c stackVisitorContext) EnterMapValue(mapNode, mapKey interface{}) visitors.Context {
	return c.appendPath(visitors.ContextPathElement{
		AncestorNode: mapNode,
		OutEdge:      visitors.NewMapValueEdge(mapKey),
	})
}

func (c stackVisitorContext) appendPath(e visitors.ContextPathElement) stackVisitorContext {
	// Allocate the next-deepest element on the stack
	newPosition := c.Preallocated.Push(c.Position)
	newContext := stackVisitorContext{
		Preallocated: c.Preallocated,
		Position:     newPosition,
	}

	// Copy the current values.
	// TODO: can we do this lazily as well? for example, keep two pointers, one for the
	// values modified below, and another for the mutable values?
	d := newContext.Delegate()
	*d = *(c.Delegate())

	// Construct the whole path lazily, just store the last element.
	d.lastPathElement = e
	d.outer = c
	return newContext
}

func (c stackVisitorContext) GetPath() visitors.ContextPath {
	// Element 0 has no path element; every other context in the stack does.
	result := make([]visitors.ContextPathElement, c.Position)
	for i := 1; i <= c.Position; i++ {
		result[i-1] = c.Preallocated.Stack[i].lastPathElement
	}
	return result
}

func (c stackVisitorContext) GetOuter() visitors.Context {
	return c.Delegate().GetOuter()
}

func (c stackVisitorContext) appendFieldPath(elt FieldPathElement) {
	c.Delegate().appendFieldPath(elt)
}

func (c stackVisitorContext) GetFieldPath() []FieldPathElement {
	return c.Delegate().GetFieldPath()
}

func (c stackVisitorContext) appendRestPath(s string) {
	c.Delegate().appendRestPath(s)
}

func (c stackVisitorContext) GetRestPath() []string {
	return c.Delegate().GetRestPath()
}

func (c stackVisitorContext) GetRestOperation() string {
	return c.Delegate().GetRestOperation()
}

func (c stackVisitorContext) setRestOperation(op string) {
	c.Delegate().setRestOperation(op)
}

func (c stackVisitorContext) IsArg() bool {
	return c.Delegate().IsArg()
}

func (c stackVisitorContext) IsResponse() bool {
	return c.Delegate().IsResponse()
}

func (c stackVisitorContext) IsOptional() bool {
	return c.Delegate().IsOptional()
}

func (c stackVisitorContext) GetValueType() HttpValueType {
	return c.Delegate().GetValueType()
}

func (c stackVisitorContext) GetHttpAuthType() *pb.HTTPAuth_HTTPAuthType {
	return c.Delegate().GetHttpAuthType()
}

func (c stackVisitorContext) GetArgPath() []string {
	return c.Delegate().GetArgPath()
}

func (c stackVisitorContext) GetResponsePath() []string {
	return c.Delegate().GetResponsePath()
}

func (c stackVisitorContext) GetEndpointPath() string {
	return c.Delegate().GetEndpointPath()
}

func (c stackVisitorContext) GetResponseCode() *string {
	return c.Delegate().GetResponseCode()
}

func (c stackVisitorContext) GetContentType() *string {
	return c.Delegate().GetContentType()
}

func (c stackVisitorContext) GetHost() string {
	return c.Delegate().GetHost()
}

func (c stackVisitorContext) GetInnermostNode(typ reflect.Type) (interface{}, SpecVisitorContext) {
	return c.Delegate().GetInnermostNode(typ)
}

func (c stackVisitorContext) getTopLevelDataPath() []string {
	return c.Delegate().getTopLevelDataPath()
}

func (c stackVisitorContext) setIsArg(isArg bool) {
	c.Delegate().setIsArg(isArg)
}

func (c stackVisitorContext) setIsOptional() {
	c.Delegate().setIsOptional()
}

func (c stackVisitorContext) setValueType(vt HttpValueType) {
	c.Delegate().setValueType(vt)
}

func (c stackVisitorContext) setResponseCode(code string) {
	c.Delegate().setResponseCode(code)
}

func (c stackVisitorContext) setContentType(contentType string) {
	c.Delegate().setContentType(contentType)
}

func (c stackVisitorContext) setHttpAuthType(at pb.HTTPAuth_HTTPAuthType) {
	c.Delegate().setHttpAuthType(at)
}

func (c stackVisitorContext) setTopLevelDataIndex(i int) {
	c.Delegate().setTopLevelDataIndex(i)
}
