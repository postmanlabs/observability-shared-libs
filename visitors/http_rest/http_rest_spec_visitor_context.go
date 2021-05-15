package http_rest

import (
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

	// Returns the host.
	GetHost() string

	// Returns the innermost Data instance being visited and its context.
	GetInnermostData() (*pb.Data, SpecVisitorContext)

	// Used by the visitor infrastructure to construct a REST path by replacing
	// parts of the AST path with names from DataMeta.
	appendRestPath(string)

	setIsArg(bool)
	setIsOptional()
	setValueType(HttpValueType)
	setHttpAuthType(pb.HTTPAuth_HTTPAuthType)
	setTopLevelDataIndex(int)
	setResponseCode(string)
}

type specVisitorContext struct {
	path  visitors.ContextPath
	outer SpecVisitorContext

	restPath []string

	// nil means we're not sure if this is an arg or response value.
	isArg *bool

	isOptional bool

	valueType    HttpValueType
	httpAuthType *pb.HTTPAuth_HTTPAuthType

	responseCode *string

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

func (c *specVisitorContext) GetHost() string {
	if len(c.restPath) < 1 {
		return ""
	}
	return c.restPath[0]
}

func (c *specVisitorContext) GetInnermostData() (*pb.Data, SpecVisitorContext) {
	for c != nil {
		if data, ok := c.path.GetLast().AncestorNode.(*pb.Data); ok {
			return data, c.outer
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

func (c *specVisitorContext) setHttpAuthType(at pb.HTTPAuth_HTTPAuthType) {
	c.httpAuthType = &at
}

func (c *specVisitorContext) setTopLevelDataIndex(i int) {
	c.topDataIndex = i
}
