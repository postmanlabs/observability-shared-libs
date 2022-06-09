package spec_util

import (
	pb "github.com/akitasoftware/akita-ir/go/api_spec"

	"github.com/akitasoftware/akita-libs/visitors/http_rest"
)

// Three maps in the IR use hashes of the values as keys (i.e. map[hash(v)] = v):
//  - Method.Args
//  - Method.Responses
//  - OneOf.Options
//
// This method traverses the spec, recomputes the hash of each value, and updates the map.
func RewriteHashKeys(spec *pb.APISpec) error {
	// Hash OneOf values in postorder, so that children are updated before computing the
	// new hash for the parent.
	v := &http_rest.RehashingContextlessSpecVisitor{}
	http_rest.Apply(v, spec)

	return nil
}
