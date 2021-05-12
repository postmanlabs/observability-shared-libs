package http_rest_diff

import (
	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	. "github.com/akitasoftware/akita-libs/visitors"
	"github.com/akitasoftware/akita-libs/visitors/http_rest"
)

// Determines whether diffing between two Data instances will produce an empty
// result.
func IsSameData(d1, d2 *pb.Data) bool {
	var comparator dataComparator
	http_rest.ApplyPair(&comparator, d1, d2)
	return !comparator.isDifferent
}

// A diff visitor for determining whether two Data instances are different.
type dataComparator struct {
	DefaultHttpRestSpecDiffVisitorImpl

	isDifferent bool
}

func (v *dataComparator) EnterDiff(self interface{}, ctx http_rest.HttpRestSpecPairVisitorContext, left, right interface{}) Cont {
	v.isDifferent = true
	return Abort
}
