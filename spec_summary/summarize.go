package spec_summary

import (
	"fmt"
	"strings"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/spec_util"
	. "github.com/akitasoftware/akita-libs/visitors"
	vis "github.com/akitasoftware/akita-libs/visitors/http_rest"
)

type FilterMap map[string]map[string]map[*pb.Method]struct{}

func (fm FilterMap) insert(filterKind string, filterValue string, method *pb.Method) {
	methodsByFilterValue, ok := fm[filterKind]
	if !ok {
		methodsByFilterValue = make(map[string]map[*pb.Method]struct{})
		fm[filterKind] = methodsByFilterValue
	}

	methods, ok := methodsByFilterValue[filterValue]
	if !ok {
		methods = make(map[*pb.Method]struct{})
		methodsByFilterValue[filterValue] = methods
	}

	methods[method] = struct{}{}
}

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
func SummarizeWithFilters(spec *pb.APISpec, filters map[string][]string) *Summary {
	v := specSummaryVisitor{
		methodSummary: &Summary{
			Authentications: make(map[string]int),
			HTTPMethods:     make(map[string]int),
			Paths:           make(map[string]int),
			Params:          make(map[string]int),
			Properties:      make(map[string]int),
			ResponseCodes:   make(map[string]int),
			DataFormats:     make(map[string]int),
			DataKinds:       make(map[string]int),
			DataTypes:       make(map[string]int),
		},
		summary: &Summary{
			Authentications: make(map[string]int),
			HTTPMethods:     make(map[string]int),
			Paths:           make(map[string]int),
			Params:          make(map[string]int),
			Properties:      make(map[string]int),
			ResponseCodes:   make(map[string]int),
			DataFormats:     make(map[string]int),
			DataKinds:       make(map[string]int),
			DataTypes:       make(map[string]int),
		},
		filtersToMethods: make(map[string]map[string]map[*pb.Method]struct{}),
	}
	vis.Apply(&v, spec)

	// If there are no known filters, return the default count.
	if filters == nil {
		return v.summary
	}
	knownFilterKeys := map[string]struct{}{
		"authentications": {},
		"http_methods":    {},
		"paths":           {},
		"params":          {},
		"properties":      {},
		"response_codes":  {},
		"data_formats":    {},
		"data_kinds":      {},
		"data_types":      {},
	}
	knownFiltersPresent := false
	for filterKey, _ := range filters {
		if _, ok := knownFilterKeys[filterKey]; ok {
			knownFiltersPresent = true
		}
	}
	if !knownFiltersPresent {
		return v.summary
	}

	// The count for a given filter value is calculated as the number of
	// methods that match it, assuming
	// - no other values of the same filter are applied
	// - all other filters are applied.
	//
	// For example, if the current filters are http_method=GET and response_code=200,
	// then the count for response_code=404 is calculated as the number of methods
	// with a 404 response code a GET http method.

	counts := make(map[string]map[string]int, len(v.filtersToMethods))

	allMethods := make(map[*pb.Method]struct{})
	for _, methodsByFilterVal := range v.filtersToMethods {
		for _, methods := range methodsByFilterVal {
			for m, _ := range methods {
				allMethods[m] = struct{}{}
			}
		}
	}

	for filterKind, methodsByFilterVal := range v.filtersToMethods {
		// Get set of all methods that match all other filters.
		methodSets := []map[*pb.Method]struct{}{allMethods}
		for otherFilterKind, otherMethodsByFilterVal := range v.filtersToMethods {
			if filterKind == otherFilterKind {
				continue
			}

			appliedFilterValues, ok := filters[otherFilterKind]

			// If no filters are being applied for this filter kind, then there are no
			// restrictions on the set of methods.
			if !ok {
				continue
			}

			// Otherwise, collect the methods for the filter values being applied.
			methodSet := make(map[*pb.Method]struct{})
			for _, appliedFilterVal := range appliedFilterValues {
				if methods, ok := otherMethodsByFilterVal[appliedFilterVal]; ok {
					for m, _ := range methods {
						methodSet[m] = struct{}{}
					}
				}
			}
			methodSets = append(methodSets, methodSet)
		}

		otherMethods := intersect(methodSets)

		// For each filter value, get the intersection of its methods with
		// otherMethods.  The size of the intersection is the count for the
		// filter value.
		for filterVal, methods := range methodsByFilterVal {
			commonMethods := intersect([]map[*pb.Method]struct{}{otherMethods, methods})

			countsByFilterVal, ok := counts[filterKind]
			if !ok {
				countsByFilterVal = make(map[string]int)
				counts[filterKind] = countsByFilterVal
			}
			countsByFilterVal[filterVal] = len(commonMethods)
		}
	}

	summary := Summary{}
	for filterKind, countsByFilterVal := range counts {
		switch filterKind {
		case "authentications":
			summary.Authentications = countsByFilterVal
		case "data_kinds":
			summary.DataKinds = countsByFilterVal
		case "data_formats":
			summary.DataFormats = countsByFilterVal
		case "data_types":
			summary.DataTypes = countsByFilterVal
		case "http_methods":
			summary.HTTPMethods = countsByFilterVal
		case "params":
			summary.Params = countsByFilterVal
		case "paths":
			summary.Paths = countsByFilterVal
		case "properties":
			summary.Properties = countsByFilterVal
		case "response_codes":
			summary.ResponseCodes = countsByFilterVal
		}
	}

	return &summary
}

type specSummaryVisitor struct {
	vis.DefaultSpecVisitorImpl

	// Count occurrences within a single method.
	methodSummary *Summary

	// Count the number of methods in which each term occurs.
	summary *Summary

	// Reverse mapping from filters to methods that match them.
	filtersToMethods FilterMap
}

var _ vis.DefaultSpecVisitor = (*specSummaryVisitor)(nil)

func (v *specSummaryVisitor) LeaveMethod(self interface{}, _ vis.SpecVisitorContext, m *pb.Method, cont Cont) Cont {
	if meta := spec_util.HTTPMetaFromMethod(m); meta != nil {
		methodName := strings.ToUpper(meta.GetMethod())
		v.summary.HTTPMethods[methodName] += 1
		v.filtersToMethods.insert("http_methods", methodName, m)

		v.summary.Paths[meta.GetPathTemplate()] += 1
		v.filtersToMethods.insert("paths", meta.GetPathTemplate(), m)
	}

	// If this method has no authentications, increment Authentications["None"].
	if len(v.methodSummary.Authentications) == 0 {
		v.summary.Authentications["None"] += 1
		v.filtersToMethods.insert("authentications", "None", m)
	}

	// For each term that occurs at least once in this method, increment the
	// summary count by one and clear the method-level summary.
	summaryPairs := []struct {
		dst  map[string]int
		src  map[string]int
		kind string
	}{
		{dst: v.summary.Authentications, src: v.methodSummary.Authentications, kind: "authentications"},
		{dst: v.summary.HTTPMethods, src: v.methodSummary.HTTPMethods, kind: "http_methods"},
		{dst: v.summary.Paths, src: v.methodSummary.Paths, kind: "paths"},
		{dst: v.summary.Params, src: v.methodSummary.Params, kind: "params"},
		{dst: v.summary.Properties, src: v.methodSummary.Properties, kind: "properties"},
		{dst: v.summary.ResponseCodes, src: v.methodSummary.ResponseCodes, kind: "response_codes"},
		{dst: v.summary.DataFormats, src: v.methodSummary.DataFormats, kind: "data_formats"},
		{dst: v.summary.DataKinds, src: v.methodSummary.DataKinds, kind: "data_kinds"},
		{dst: v.summary.DataTypes, src: v.methodSummary.DataTypes, kind: "data_types"},
	}
	for _, summaryPair := range summaryPairs {
		for key, count := range summaryPair.src {
			if count > 0 {
				summaryPair.dst[key] += 1
				v.filtersToMethods.insert(summaryPair.kind, key, m)
			}
			delete(summaryPair.src, key)
		}
	}

	return cont
}

func (v *specSummaryVisitor) LeaveData(self interface{}, _ vis.SpecVisitorContext, d *pb.Data, cont Cont) Cont {
	// Handle auth vs params vs properties.
	if meta := spec_util.HTTPAuthFromData(d); meta != nil {
		// For proprietary headers, use the header value, otherwise
		// use the type.
		switch meta.Type {
		case pb.HTTPAuth_PROPRIETARY_HEADER:
			v.methodSummary.Authentications[meta.ProprietaryHeader] += 1
		default:
			v.methodSummary.Authentications[meta.Type.String()] += 1
		}
	} else if meta := spec_util.HTTPPathFromData(d); meta != nil {
		v.methodSummary.Params[meta.Key] += 1
	} else if meta := spec_util.HTTPQueryFromData(d); meta != nil {
		v.methodSummary.Params[meta.Key] += 1
	} else if meta := spec_util.HTTPHeaderFromData(d); meta != nil {
		v.methodSummary.Params[meta.Key] += 1
	} else if meta := spec_util.HTTPCookieFromData(d); meta != nil {
		v.methodSummary.Params[meta.Key] += 1
	} else {
		if s, ok := d.Value.(*pb.Data_Struct); ok {
			for k := range s.Struct.GetFields() {
				v.methodSummary.Properties[k] += 1
			}
		}
	}

	// Handle response codes.
	if meta := spec_util.HTTPMetaFromData(d); meta != nil {
		if meta.GetResponseCode() != 0 { // response code 0 means it's a request
			v.methodSummary.ResponseCodes[fmt.Sprintf("%d", meta.GetResponseCode())] += 1
		}
	}

	return cont
}

func (v *specSummaryVisitor) LeavePrimitive(self interface{}, _ vis.SpecVisitorContext, p *pb.Primitive, cont Cont) Cont {
	for f := range p.GetFormats() {
		v.methodSummary.DataFormats[f] += 1
	}
	if k := p.GetFormatKind(); k != "" {
		v.methodSummary.DataKinds[k] += 1
	}
	v.methodSummary.DataTypes[spec_util.TypeOfPrimitive(p)] += 1
	return cont
}

func intersect(methodSets []map[*pb.Method]struct{}) map[*pb.Method]struct{} {
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
