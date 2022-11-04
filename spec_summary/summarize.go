package spec_summary

import (
	"fmt"
	"reflect"
	"strings"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/spec_util"
	. "github.com/akitasoftware/akita-libs/visitors"
	vis "github.com/akitasoftware/akita-libs/visitors/http_rest"
	"github.com/akitasoftware/go-utils/sets"
	"github.com/akitasoftware/go-utils/slices"
	"github.com/golang/glog"
)

// See SummarizeWithFilters.
func Summarize(spec *pb.APISpec) *Summary {
	return SummarizeWithFilters(spec, nil)
}

// Produce a summary such that the count for each summary value reflects the
// number of endpoints that would be present in the spec if that value were
// applied as a filter, while considering other existing filters.
//
// For example, suppose filters were { response_codes: [404] }.  If the summary
// included HTTPMethods: {"GET": 2}, it would mean that there are two GET
// methods with 404 response codes.
func SummarizeWithFilters(spec *pb.APISpec, filters Filters) *Summary {
	return SummarizeByDirectionWithFilters(spec, filters).ToSummary()
}

// As Summarize, but distinguishing by direction in the response.
func SummarizeByDirection(spec *pb.APISpec) *SummaryByDirection {
	return SummarizeByDirectionWithFilters(spec, nil)
}

// As SummarizeWithFilters, but distinguishing by direction in the response.
func SummarizeByDirectionWithFilters(spec *pb.APISpec, filters Filters) *SummaryByDirection {
	v := specSummaryVisitor{
		methodSummary:    NewSummaryByDirection(),
		summary:          NewSummaryByDirection(),
		filtersToMethods: NewFiltersToMethods[*pb.Method](),
	}
	vis.Apply(&v, spec)

	if len(filters) == 0 {
		return v.summary
	}

	summary, _ := v.filtersToMethods.SummarizeWithFilters(filters)
	return summary
}

type specSummaryVisitor struct {
	vis.DefaultSpecVisitorImpl

	// Count occurrences within a single method.
	methodSummary *SummaryByDirection

	// Count the number of methods in which each term occurs.
	summary *SummaryByDirection

	// Reverse mapping from filters to methods that match them.
	filtersToMethods *FiltersToMethods[*pb.Method]
}

var _ vis.DefaultSpecVisitor = (*specSummaryVisitor)(nil)

func (v *specSummaryVisitor) LeaveMethod(self interface{}, _ vis.SpecVisitorContext, m *pb.Method, cont Cont) Cont {
	if meta := spec_util.HTTPMetaFromMethod(m); meta != nil {
		methodName := strings.ToUpper(meta.GetMethod())
		v.summary.NondirectedFilters.Increment(HttpMethodFilter, methodName)
		v.filtersToMethods.InsertNondirectionalFilter("http_methods", methodName, m, sets.NewSet(m))

		v.summary.NondirectedFilters.Increment(PathFilter, meta.GetPathTemplate())
		v.filtersToMethods.InsertNondirectionalFilter("paths", meta.GetPathTemplate(), m, sets.NewSet(m))

		v.summary.NondirectedFilters.Increment(HostFilter, meta.GetHost())
		v.filtersToMethods.InsertNondirectionalFilter("hosts", meta.GetHost(), m, sets.NewSet(m))
	}

	// If this method has no authentications, increment Authentications["None"].
	if v.methodSummary.DirectedFilters.GetCountsByValue(RequestDirection, AuthFilter) == nil {
		v.summary.DirectedFilters.Increment(RequestDirection, AuthFilter, "None")
		v.filtersToMethods.InsertDirectionalFilter(RequestDirection, AuthFilter, "None", m, sets.NewSet(m))
	}

	// For each term that occurs at least once in this method, increment the
	// summary count by one.
	v.methodSummary.NondirectedFilters.ForEach(func(kind FilterKind, value FilterValue, count int) bool {
		if count > 0 {
			v.summary.NondirectedFilters.Increment(kind, value)
			v.filtersToMethods.InsertNondirectionalFilter(kind, value, m, sets.NewSet(m))
		}
		return true
	})
	v.methodSummary.DirectedFilters.ForEach(func(direction Direction, kind FilterKind, value FilterValue, count int) bool {
		if count > 0 {
			v.summary.DirectedFilters.Increment(direction, kind, value)
			v.filtersToMethods.InsertDirectionalFilter(direction, kind, value, m, sets.NewSet(m))
		}
		return true
	})

	// Clear the method-level summary.
	v.methodSummary = NewSummaryByDirection()

	return cont
}

func (v *specSummaryVisitor) LeaveData(self interface{}, context vis.SpecVisitorContext, d *pb.Data, cont Cont) Cont {
	direction := RequestDirection
	if context.IsResponse() {
		direction = ResponseDirection
	}

	// Handle auth vs params vs properties.
	if meta := spec_util.HTTPAuthFromData(d); meta != nil {
		v.methodSummary.DirectedFilters.Increment(direction, AuthFilter, meta.Type.String())
	} else if meta := spec_util.HTTPPathFromData(d); meta != nil {
		v.methodSummary.DirectedFilters.Increment(direction, ParamFilter, meta.Key)
	} else if meta := spec_util.HTTPQueryFromData(d); meta != nil {
		v.methodSummary.DirectedFilters.Increment(direction, ParamFilter, meta.Key)
	} else if meta := spec_util.HTTPHeaderFromData(d); meta != nil {
		v.methodSummary.DirectedFilters.Increment(direction, ParamFilter, meta.Key)
	} else if meta := spec_util.HTTPCookieFromData(d); meta != nil {
		v.methodSummary.DirectedFilters.Increment(direction, ParamFilter, meta.Key)
	} else {
		if d == nil {
			glog.Errorf("[SPEC_SUMMARY_VISITOR_BUG] Context of nil data: %v", strings.Join(slices.Map(context.GetPath(), func(e ContextPathElement) string {
				return fmt.Sprintf("[%s] %s", reflect.TypeOf(e.AncestorNode), e.OutEdge)
			}), " . "))
		}
		if s, ok := d.GetValue().(*pb.Data_Struct); ok && s != nil {
			for k := range s.Struct.GetFields() {
				v.methodSummary.DirectedFilters.Increment(direction, PropertyFilter, k)
			}
		}
	}

	if context.IsResponse() {
		v.methodSummary.NondirectedFilters.Increment(DirectionFilter, "response")
	} else {
		v.methodSummary.NondirectedFilters.Increment(DirectionFilter, "request")
	}

	// Handle response codes.
	if meta := spec_util.HTTPMetaFromData(d); meta != nil {
		if meta.GetResponseCode() != 0 { // response code 0 means it's a request
			v.methodSummary.DirectedFilters.Increment(direction, ResponseCodeFilter, fmt.Sprintf("%d", meta.GetResponseCode()))
		}
	}

	return cont
}

func (v *specSummaryVisitor) LeavePrimitive(self interface{}, context vis.SpecVisitorContext, p *pb.Primitive, cont Cont) Cont {
	direction := RequestDirection
	if context.IsResponse() {
		direction = ResponseDirection
	}

	for f := range p.GetFormats() {
		v.methodSummary.DirectedFilters.Increment(direction, DataFormatFilter, f)
	}

	if k := p.GetFormatKind(); k != "" {
		v.methodSummary.DirectedFilters.Increment(direction, DataFormatKindFilter, k)
	}

	v.methodSummary.DirectedFilters.Increment(direction, DataTypeFilter, spec_util.TypeOfPrimitive(p))

	return cont
}

func intersect(methodSets ...map[*pb.Method]struct{}) map[*pb.Method]struct{} {
	result := make(map[*pb.Method]struct{})
	if len(methodSets) == 0 {
		return result
	}

	isFirst := true
	for _, methods := range methodSets {
		if isFirst {
			// Initialize result with contents of first map.
			for m, _ := range methods {
				result[m] = struct{}{}
			}
			isFirst = false
		} else {
			// Remove methods in result not in each other filter.
			for m, _ := range result {
				if _, ok := methods[m]; !ok {
					delete(result, m)
				}
			}
		}
	}

	return result
}
