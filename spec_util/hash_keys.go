package spec_util

import (
	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/pkg/errors"

	"github.com/akitasoftware/akita-libs/spec_util/ir_hash"
	. "github.com/akitasoftware/akita-libs/visitors"
	"github.com/akitasoftware/akita-libs/visitors/http_rest"
)

type hashOneOfVisitor struct {
	http_rest.DefaultSpecVisitorImpl
	err error
}

var _ http_rest.DefaultSpecVisitor = (*hashOneOfVisitor)(nil)

func (vis *hashOneOfVisitor) LeaveData(self interface{}, c http_rest.SpecVisitorContext, p *pb.Data, cont Cont) Cont {
	oneOf := p.GetOneof()
	if oneOf == nil {
		return cont
	}

	options := p.GetOneof().Options
	keys := make([]string, 0, len(options))

	for k, _ := range options {
		keys = append(keys, k)
	}

	for _, k := range keys {
		v := options[k]
		h := ir_hash.HashDataToString(v)
		if k != h {
			delete(options, k)
			options[h] = v
		}
	}
	return cont
}

// Three maps in the IR use hashes of the values as keys (i.e. map[hash(v)] = v):
//  - Method.Args
//  - Method.Responses
//  - OneOf.Options
//
// This method traverses the spec, recomputes the hash of each value, and updates the map.
func RewriteHashKeys(spec *pb.APISpec) error {
	// Hash OneOf values in postorder, so that children are updated before computing the
	// new hash for the parent.
	v := &hashOneOfVisitor{}
	http_rest.Apply(v, spec)
	if v.err != nil {
		return errors.Wrap(v.err, "failed to compute hash of oneOf data")
	}

	// Hash Args and Responses for each method.
	for _, method := range spec.Methods {
		for _, m := range []map[string]*pb.Data{method.Args, method.Responses} {
			keys := make([]string, 0, len(m))
			for k, _ := range m {
				keys = append(keys, k)
			}
			for _, k := range keys {
				arg := m[k]
				h := ir_hash.HashDataToString(arg)
				if h != k {
					delete(m, k)
					m[h] = arg
				}
			}
		}
	}

	return nil
}
