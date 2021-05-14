package http_rest

import "github.com/akitasoftware/akita-libs/visitors"

// Basically a pair of HttpRestSpecVisitorContexts. See that class for more
// information.
type SpecPairVisitorContext interface {
	visitors.PairContext

	ExtendLeftContext(leftNode interface{})
	ExtendRightContext(rightNode interface{})

	GetLeftContext() SpecVisitorContext
	GetRightContext() SpecVisitorContext
	SplitContext() (SpecVisitorContext, SpecVisitorContext)

	AppendRestPaths(string, string) SpecPairVisitorContext
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

type specPairVisitorContext struct {
	left  *httpRestSpecVisitorContext
	right *httpRestSpecVisitorContext
}

func newSpecPairVisitorContext() SpecPairVisitorContext {
	return &specPairVisitorContext{
		left:  &httpRestSpecVisitorContext{},
		right: &httpRestSpecVisitorContext{},
	}
}

func (c *specPairVisitorContext) ExtendLeftContext(leftNode interface{}) {
	c.left = extendContext(c.left, leftNode).(*httpRestSpecVisitorContext)
}

func (c *specPairVisitorContext) ExtendRightContext(rightNode interface{}) {
	c.right = extendContext(c.right, rightNode).(*httpRestSpecVisitorContext)
}

func (c *specPairVisitorContext) GetLeftContext() SpecVisitorContext {
	return c.left
}

func (c *specPairVisitorContext) GetRightContext() SpecVisitorContext {
	return c.right
}

func (c *specPairVisitorContext) SplitContext() (SpecVisitorContext, SpecVisitorContext) {
	return c.left, c.right
}

func (c *specPairVisitorContext) AppendPaths(left, right string) visitors.PairContext {
	rv := *c
	rv.left = c.left.AppendPath(left).(*httpRestSpecVisitorContext)
	rv.right = c.right.AppendPath(right).(*httpRestSpecVisitorContext)
	return &rv
}

func (c *specPairVisitorContext) GetPaths() ([]string, []string) {
	return c.left.GetPath(), c.right.GetPath()
}

func (c *specPairVisitorContext) AppendRestPaths(left, right string) SpecPairVisitorContext {
	rv := *c
	rv.left = c.left.AppendRestPath(left).(*httpRestSpecVisitorContext)
	rv.right = c.right.AppendRestPath(right).(*httpRestSpecVisitorContext)
	return &rv
}

func (c *specPairVisitorContext) GetRestPaths() ([]string, []string) {
	return c.left.GetRestPath(), c.right.GetRestPath()
}

func (c *specPairVisitorContext) GetRestOperations() (string, string) {
	return c.left.GetRestOperation(), c.right.GetRestOperation()
}

func (c *specPairVisitorContext) setRestOperations(left, right string) {
	c.left.setRestOperation(left)
	c.right.setRestOperation(right)
}

func (c *specPairVisitorContext) IsArg() (bool, bool) {
	return c.left.IsArg(), c.right.IsArg()
}

func (c *specPairVisitorContext) IsResponse() (bool, bool) {
	return c.left.IsResponse(), c.right.IsResponse()
}

func (c *specPairVisitorContext) GetValueTypes() (HttpValueType, HttpValueType) {
	return c.left.GetValueType(), c.right.GetValueType()
}

func (c *specPairVisitorContext) GetArgPaths() ([]string, []string) {
	return c.left.GetArgPath(), c.right.GetArgPath()
}

func (c *specPairVisitorContext) GetResponsePaths() ([]string, []string) {
	return c.left.GetResponsePath(), c.right.GetResponsePath()
}

func (c *specPairVisitorContext) GetEndpointPaths() (string, string) {
	return c.left.GetEndpointPath(), c.right.GetEndpointPath()
}

func (c *specPairVisitorContext) GetHosts() (string, string) {
	return c.left.GetHost(), c.right.GetHost()
}

func (c *specPairVisitorContext) setIsArg(left, right bool) {
	c.left.setIsArg(left)
	c.right.setIsArg(right)
}

func (c *specPairVisitorContext) setValueType(left, right HttpValueType) {
	c.left.setValueType(left)
	c.right.setValueType(right)
}

func (c *specPairVisitorContext) setTopLevelDataIndex(left, right int) {
	c.left.setTopLevelDataIndex(left)
	c.right.setTopLevelDataIndex(right)
}
