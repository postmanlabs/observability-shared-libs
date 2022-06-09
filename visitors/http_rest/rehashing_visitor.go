package http_rest

import (
	"github.com/akitasoftware/akita-ir/go/api_spec"

	"github.com/akitasoftware/akita-libs/spec_util/ir_hash"
	. "github.com/akitasoftware/akita-libs/visitors"
)

// An abstract contextless visitor for recomputing node hashes while walking
// back up the tree.
type RehashingContextlessSpecVisitor struct {
	DefaultContextlessSpecVisitorImpl
}

var _ SpecVisitor = (*RehashingContextlessSpecVisitor)(nil)

func (*RehashingContextlessSpecVisitor) LeaveMethod(self interface{}, _ SpecVisitorContext, method *api_spec.Method, cont Cont) Cont {
	method.Args = rehashDataMap(method.Args)
	method.Responses = rehashDataMap(method.Responses)
	return cont
}

func (*RehashingContextlessSpecVisitor) LeaveOneOf(self interface{}, _ SpecVisitorContext, oneOf *api_spec.OneOf, cont Cont) Cont {
	oneOf.Options = rehashDataMap(oneOf.Options)
	return cont
}

// An abstract visitor for recomputing node hashes while walking back up the
// tree.
type RehashingSpecVisitor struct {
	DefaultSpecVisitorImpl
}

var _ SpecVisitor = (*RehashingSpecVisitor)(nil)

func (*RehashingSpecVisitor) LeaveMethod(self interface{}, _ SpecVisitorContext, method *api_spec.Method, cont Cont) Cont {
	method.Args = rehashDataMap(method.Args)
	method.Responses = rehashDataMap(method.Responses)
	return cont
}

func (*RehashingSpecVisitor) LeaveOneOf(self interface{}, _ SpecVisitorContext, oneOf *api_spec.OneOf, cont Cont) Cont {
	oneOf.Options = rehashDataMap(oneOf.Options)
	return cont
}

func rehashDataMap(data map[string]*api_spec.Data) map[string]*api_spec.Data {
	result := make(map[string]*api_spec.Data, len(data))
	for _, elt := range data {
		h := ir_hash.HashDataToString(elt)
		result[h] = elt
	}
	return result
}
