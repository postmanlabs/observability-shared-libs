package http_rest

import "github.com/akitasoftware/akita-libs/visitors"

// Basically a pair of HttpRestSpecVisitorContexts. See that class for more
// information.
type HttpRestSpecPairVisitorContext interface {
	visitors.PairContext

	ExtendLeftContext(leftNode interface{})
	ExtendRightContext(rightNode interface{})

	AppendRestPaths(string, string) HttpRestSpecPairVisitorContext
	GetRestPaths() ([]string, []string)
	GetRestOperations() (string, string)
	setRestOperations(string, string)
	IsArg() (bool, bool)
	IsResponse() (bool, bool)
	GetValueTypes() (HttpValueType, HttpValueType)
	GetArgPaths() ([]string, []string)
	GetResponsePaths() ([]string, []string)
	GetEndpointPaths() (string, string)
	GetHosts() (string, string)
	setIsArg(bool, bool)
	setValueType(HttpValueType, HttpValueType)
	setTopLevelDataIndex(int, int)
}

type httpRestSpecPairVisitorContext struct {
	left  *httpRestSpecVisitorContext
	right *httpRestSpecVisitorContext
}

func newHttpRestSpecPairVisitorContext() HttpRestSpecPairVisitorContext {
	return &httpRestSpecPairVisitorContext{
		left:  &httpRestSpecVisitorContext{},
		right: &httpRestSpecVisitorContext{},
	}
}

func (c *httpRestSpecPairVisitorContext) ExtendLeftContext(leftNode interface{}) {
	c.left = extendContext(c.left, leftNode).(*httpRestSpecVisitorContext)
}

func (c *httpRestSpecPairVisitorContext) ExtendRightContext(rightNode interface{}) {
	c.right = extendContext(c.right, rightNode).(*httpRestSpecVisitorContext)
}

func (c *httpRestSpecPairVisitorContext) AppendPaths(left, right string) visitors.PairContext {
	rv := *c
	rv.left = c.left.AppendPath(left).(*httpRestSpecVisitorContext)
	rv.right = c.right.AppendPath(right).(*httpRestSpecVisitorContext)
	return &rv
}

func (c *httpRestSpecPairVisitorContext) GetPaths() ([]string, []string) {
	return c.left.GetPath(), c.right.GetPath()
}

func (c *httpRestSpecPairVisitorContext) AppendRestPaths(left, right string) HttpRestSpecPairVisitorContext {
	rv := *c
	rv.left = c.left.AppendRestPath(left).(*httpRestSpecVisitorContext)
	rv.right = c.right.AppendRestPath(right).(*httpRestSpecVisitorContext)
	return &rv
}

func (c *httpRestSpecPairVisitorContext) GetRestPaths() ([]string, []string) {
	return c.left.GetRestPath(), c.right.GetRestPath()
}

func (c *httpRestSpecPairVisitorContext) GetRestOperations() (string, string) {
	return c.left.GetRestOperation(), c.right.GetRestOperation()
}

func (c *httpRestSpecPairVisitorContext) setRestOperations(left, right string) {
	c.left.setRestOperation(left)
	c.right.setRestOperation(right)
}

func (c *httpRestSpecPairVisitorContext) IsArg() (bool, bool) {
	return c.left.IsArg(), c.right.IsArg()
}

func (c *httpRestSpecPairVisitorContext) IsResponse() (bool, bool) {
	return c.left.IsResponse(), c.right.IsResponse()
}

func (c *httpRestSpecPairVisitorContext) GetValueTypes() (HttpValueType, HttpValueType) {
	return c.left.GetValueType(), c.right.GetValueType()
}

func (c *httpRestSpecPairVisitorContext) GetArgPaths() ([]string, []string) {
	return c.left.GetArgPath(), c.right.GetArgPath()
}

func (c *httpRestSpecPairVisitorContext) GetResponsePaths() ([]string, []string) {
	return c.left.GetResponsePath(), c.right.GetResponsePath()
}

func (c *httpRestSpecPairVisitorContext) GetEndpointPaths() (string, string) {
	return c.left.GetEndpointPath(), c.right.GetEndpointPath()
}

func (c *httpRestSpecPairVisitorContext) GetHosts() (string, string) {
	return c.left.GetHost(), c.right.GetHost()
}

func (c *httpRestSpecPairVisitorContext) setIsArg(left, right bool) {
	c.left.setIsArg(left)
	c.right.setIsArg(right)
}

func (c *httpRestSpecPairVisitorContext) setValueType(left, right HttpValueType) {
	c.left.setValueType(left)
	c.right.setValueType(right)
}

func (c *httpRestSpecPairVisitorContext) setTopLevelDataIndex(left, right int) {
	c.left.setTopLevelDataIndex(left)
	c.right.setTopLevelDataIndex(right)
}
