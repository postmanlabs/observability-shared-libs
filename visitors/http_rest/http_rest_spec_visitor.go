package http_rest

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"

	"github.com/akitasoftware/akita-libs/visitors"
	"github.com/akitasoftware/akita-libs/visitors/go_ast"
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

// VisitorManager that lets you read each message in an APISpec, starting with the
// APISpec message itself.  Returning false stops the traversal.
type HttpRestSpecVisitor interface {
	VisitAPISpec(HttpRestSpecVisitorContext, *pb.APISpec) bool
	VisitMethod(HttpRestSpecVisitorContext, *pb.Method) bool
	VisitData(HttpRestSpecVisitorContext, *pb.Data) bool
	VisitPrimitive(HttpRestSpecVisitorContext, *pb.Primitive) bool
}

// Defines nops for all visitor methods in HttpRestSpecVisitor.
type DefaultHttpRestSpecVisitor struct{}

func (*DefaultHttpRestSpecVisitor) VisitAPISpec(c HttpRestSpecVisitorContext, spec *pb.APISpec) bool {
	return true
}

func (*DefaultHttpRestSpecVisitor) VisitMethod(c HttpRestSpecVisitorContext, m *pb.Method) bool {
	return true
}

func (*DefaultHttpRestSpecVisitor) VisitData(c HttpRestSpecVisitorContext, d *pb.Data) bool {
	return true
}

func (*DefaultHttpRestSpecVisitor) VisitPrimitive(c HttpRestSpecVisitorContext, d *pb.Primitive) bool {
	return true
}

func extendContext(cin visitors.Context, rin interface{}, x interface{}) visitors.Context {
	r, ok := rin.(HttpRestSpecVisitor)
	c, ok := cin.(HttpRestSpecVisitorContext)
	rc := cin
	if !ok {
		panic(fmt.Sprintf("HttpRestSpecVisitor.Visit expected HttpRestSpecVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}

	// Dispatch on type and path.
	switch x.(type) {
	case pb.APISpec, pb.Method, pb.Data, pb.Primitive:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		rc = extendContext(c, r, &x)
	case *pb.Method:
		m, _ := x.(*pb.Method)

		// Update the RestPath in the context
		meta := m.GetMeta().GetHttp()
		if meta != nil {
			c.setRestOperation(meta.GetMethod())
			c = c.AppendRestPath(meta.GetHost()).AppendRestPath(meta.GetMethod()).AppendRestPath(meta.GetPathTemplate())
			rc = c
		}
	case *pb.Data:
		d, _ := x.(*pb.Data)

		// Update the RestPath in the context
		// HTTPMeta is only valid for the top-level Data object.
		if d.GetMeta() != nil && d.GetMeta().GetHttp() != nil {
			c.setTopLevelDataIndex(len(c.GetRestPath()) - 1)

			meta := d.GetMeta().GetHttp()
			switch rc := meta.GetResponseCode(); rc {
			case 0: // arg
				c.setIsArg(true)
				c = c.AppendRestPath("Arg")
			default:
				c.setIsArg(false)
				c = c.AppendRestPath("Response")
				if rc == -1 {
					c = c.AppendRestPath("default")
				} else {
					c = c.AppendRestPath(strconv.Itoa(int(rc)))
				}
			}

			var valueKey string
			if x := meta.GetPath(); x != nil {
				c.setValueType(PATH)
				valueKey = x.GetKey()
			} else if x := meta.GetQuery(); x != nil {
				c.setValueType(QUERY)
				valueKey = x.GetKey()
			} else if x := meta.GetHeader(); x != nil {
				c.setValueType(HEADER)
				valueKey = x.GetKey()
			} else if x := meta.GetCookie(); x != nil {
				c.setValueType(COOKIE)
				valueKey = x.GetKey()
			} else if x := meta.GetBody(); x != nil {
				c.setValueType(BODY)
				valueKey = x.GetContentType().String()
			}

			c = c.AppendRestPath(c.GetValueType().String())
			c = c.AppendRestPath(valueKey)

			// Do nothing for HTTPEmpty
		} else {
			astPath := c.GetPath()
			c = c.AppendRestPath(astPath[len(astPath)-1])
		}
		rc = c
	}

	return rc
}

func visit(cin visitors.Context, rin interface{}, x interface{}) bool {
	r, ok := rin.(HttpRestSpecVisitor)
	c, ok := extendContext(cin, rin, x).(HttpRestSpecVisitorContext)
	if !ok {
		panic(fmt.Sprintf("HttpRestSpecVisitor.Visit expected HttpRestSpecVisitorContext, got %s",
			reflect.TypeOf(cin)))
	}
	keepGoing := true

	// Dispatch on type and path.
	switch x.(type) {
	case pb.APISpec, pb.Method, pb.Data, pb.Primitive:
		// For simplicity, ensure we're operating on a pointer to any complex
		// structure.
		keepGoing = visit(c, r, &x)
	case *pb.APISpec:
		s, _ := x.(*pb.APISpec)
		keepGoing = r.VisitAPISpec(c, s)
	case *pb.Method:
		m, _ := x.(*pb.Method)
		keepGoing = r.VisitMethod(c, m)
	case *pb.Data:
		d, _ := x.(*pb.Data)
		keepGoing = r.VisitData(c, d)
	case *pb.Primitive:
		p, _ := x.(*pb.Primitive)
		keepGoing = r.VisitPrimitive(c, p)
	default:
		// Just keep going if we don't understand the type.
	}

	return keepGoing
}

// Visits m with v.  Returns false if the visitor aborts traversal (by
// returning false).  Order is either PREORDER or POSTORDER.
func Apply(order go_ast.TraversalOrder, v HttpRestSpecVisitor, m interface{}) bool {
	c := new(httpRestSpecVisitorContext)
	vis := visitors.NewVisitorManager(c, v, visit, extendContext)
	return go_ast.Apply(order, vis, m)
}

func GetPrimitiveType(p *pb.Primitive) reflect.Type {
	if t := p.GetBoolValue(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetBytesValue(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetStringValue(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetInt32Value(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetInt64Value(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetUint32Value(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetUint64Value(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetDoubleValue(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else if t := p.GetFloatValue(); t != nil {
		return reflect.TypeOf(t).Elem()
	} else {
		panic("Unknown primitive type")
	}
}

func GetPrimitiveValue(p *pb.Primitive) string {
	if t := p.GetBoolValue(); t != nil {
		return strconv.FormatBool(t.Value)
	} else if t := p.GetBytesValue(); t != nil {
		return string(t.Value)
	} else if t := p.GetStringValue(); t != nil {
		return t.Value
	} else if t := p.GetInt32Value(); t != nil {
		return strconv.Itoa(int(t.Value))
	} else if t := p.GetInt64Value(); t != nil {
		return strconv.Itoa(int(t.Value))
	} else if t := p.GetUint32Value(); t != nil {
		return strconv.FormatUint(uint64(t.Value), 10)
	} else if t := p.GetUint64Value(); t != nil {
		return strconv.FormatUint(t.Value, 10)
	} else if t := p.GetDoubleValue(); t != nil {
		return fmt.Sprintf("%f", t.Value)
	} else if t := p.GetFloatValue(); t != nil {
		return fmt.Sprintf("%f", t.Value)
	} else {
		panic("Unknown primitive type")
	}
}

type PrintVisitor struct {
	DefaultHttpRestSpecVisitor
}

func (*PrintVisitor) VisitData(ctx HttpRestSpecVisitorContext, d *pb.Data) bool {
	fmt.Printf("%s %s\n", strings.Join(ctx.GetRestPath(), "."), d)
	return true
}

func (*PrintVisitor) VisitPrimitive(ctx HttpRestSpecVisitorContext, p *pb.Primitive) bool {
	fmt.Printf("%s %s (%s)\n", strings.Join(ctx.GetRestPath(), "."), GetPrimitiveValue(p), GetPrimitiveType(p))
	return true
}
