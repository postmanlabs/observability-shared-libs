package http_rest

import "github.com/akitasoftware/akita-libs/visitors"

// Describes which part of an HTTP request/response a value belongs to.
type HttpValueType int

const (
	UNKNOWN HttpValueType = iota
	PATH
	QUERY
	HEADER
	COOKIE
	BODY
)

func (t HttpValueType) String() string {
	return []string{"Unknown", "Path", "Query", "Header", "Cookie", "Body"}[t]
}

type HttpRestSpecVisitorContext interface {
	visitors.Context

	// Used by the visitor infrastructure to construct a REST path by replacing
	// parts of the AST path with names from DataMeta.
	AppendRestPath(string) HttpRestSpecVisitorContext

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

	GetValueType() HttpValueType

	// Like GetRestPath, except only including the portion about the argument
	// value.
	GetArgPath() []string

	// Like GetRestPath, except only including the portion about the argument
	// value.
	GetResponsePath() []string

	// Returns the endpoint path, e.g. the "/v1/users" part of
	// "localhost GET /v1/users".
	GetEndpointPath() string

	// Returns the host.
	GetHost() string

	setIsArg(bool)
	setValueType(HttpValueType)
	setTopLevelDataIndex(int)
}

type httpRestSpecVisitorContext struct {
	path     []string
	restPath []string

	// nil means we're not sure if this is an arg or response value.
	isArg *bool

	valueType HttpValueType

	// Index within restPath of the start of the section describing arg or
	// response.
	topDataIndex int

	restOperation string
}

func (c *httpRestSpecVisitorContext) AppendPath(s string) visitors.Context {
	rv := *c
	rv.path = append(c.GetPath(), s)
	return &rv
}

func (c *httpRestSpecVisitorContext) GetPath() []string {
	return c.path
}

func (c *httpRestSpecVisitorContext) AppendRestPath(s string) HttpRestSpecVisitorContext {
	rv := *c
	rv.restPath = append(c.GetRestPath(), s)
	return &rv
}

func (c *httpRestSpecVisitorContext) GetRestPath() []string {
	return c.restPath
}

func (c *httpRestSpecVisitorContext) GetRestOperation() string {
	return c.restOperation
}

func (c *httpRestSpecVisitorContext) setRestOperation(op string) {
	c.restOperation = op
}

func (c *httpRestSpecVisitorContext) IsArg() bool {
	if c.isArg != nil {
		return *c.isArg
	}
	return false
}

func (c *httpRestSpecVisitorContext) IsResponse() bool {
	if c.isArg != nil {
		return !*c.isArg
	}
	return false
}

func (c *httpRestSpecVisitorContext) GetValueType() HttpValueType {
	return c.valueType
}

func (c *httpRestSpecVisitorContext) GetArgPath() []string {
	if !c.IsArg() {
		return nil
	}
	return c.getTopLevelDataPath()
}

func (c *httpRestSpecVisitorContext) GetResponsePath() []string {
	if !c.IsResponse() {
		return nil
	}
	return c.getTopLevelDataPath()
}

func (c *httpRestSpecVisitorContext) GetEndpointPath() string {
	if len(c.restPath) < 3 {
		return ""
	}
	return c.restPath[2]
}

func (c *httpRestSpecVisitorContext) GetHost() string {
	if len(c.restPath) < 1 {
		return ""
	}
	return c.restPath[0]
}

func (c *httpRestSpecVisitorContext) getTopLevelDataPath() []string {
	p := c.GetRestPath()
	if len(p) > c.topDataIndex+1 {
		return p[c.topDataIndex+1:]
	}
	return nil
}

func (c *httpRestSpecVisitorContext) setIsArg(isArg bool) {
	c.isArg = &isArg
}

func (c *httpRestSpecVisitorContext) setValueType(vt HttpValueType) {
	c.valueType = vt
}

func (c *httpRestSpecVisitorContext) setTopLevelDataIndex(i int) {
	c.topDataIndex = i
}
