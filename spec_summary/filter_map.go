package spec_summary

import (
	. "github.com/akitasoftware/go-utils/sets"
)

// Maps all supported filter kinds to the methods that match them. The methods
// for each filter kind are indexed by the filter's value for each method.
//
// For example, if the supported filter kinds are operation and path, then the
// method set {"GET /", "PUT /"} is represented as follows:
//
//   operation -> "GET" -> "GET /"
//                "PUT" -> "PUT /"
//   path -> "/" -> "GET /", "PUT /"
type FiltersToMethods[MethodID comparable] struct {
	// For non-directional filters.
	filterMap FilterMap[MethodID]

	// For directional filters.
	filterMapByDirection FilterMapByDirection[MethodID]

	// All methods.
	allMethods Set[MethodID]
}

func NewFiltersToMethods[MethodID comparable]() *FiltersToMethods[MethodID] {
	return &FiltersToMethods[MethodID]{
		filterMap:            make(FilterMap[MethodID]),
		filterMapByDirection: make(FilterMapByDirection[MethodID]),
		allMethods:           make(Set[MethodID]),
	}
}

func (fm *FiltersToMethods[MethodID]) InsertNondirectionalFilter(filter FilterKind, value string, method MethodID) {
	if fm.filterMap == nil {
		fm.filterMap = make(FilterMap[MethodID])
	}
	fm.filterMap.Insert(filter, value, method)

	if fm.allMethods == nil {
		fm.allMethods = make(Set[MethodID])
	}
	fm.allMethods.Insert(method)
}

func (fm *FiltersToMethods[MethodID]) InsertDirectionalFilter(direction Direction, filter FilterKind, value string, method MethodID) {
	if fm.filterMapByDirection == nil {
		fm.filterMapByDirection = make(FilterMapByDirection[MethodID])
	}
	fm.filterMapByDirection.Insert(direction, filter, value, method)

	if fm.allMethods == nil {
		fm.allMethods = make(Set[MethodID])
	}
	fm.allMethods.Insert(method)
}

// Returns a summary of the effect of adding an additional filter to a set of
// filters that have already been applied. For filter kinds that are already
// applied, the summary indicates how many methods would be added to the result
// set if an additional value were added for that filter kind. For filter kinds
// that are not yet been applied, the summary indicates how many methods would
// remain in the result set if a filter with that kind were added to the filter
// set.
//
// For example, if the current filters are http_methods=GET and
// response_codes=200, then the count for response_codes=404 will be the number
// of methods with a 404 response code and a GET http method, whereas the count
// for paths=/ will be the number of methods with a 200 response code, a GET
// http method, and path "/".
func (fm *FiltersToMethods[MethodID]) SummarizeWithFilters(appliedFilters Filters) (*SummaryByDirection, Set[MethodID]) {
	// Remove any unknown filters.
	knownFilters := ParseFiltersIgnoreErrors(appliedFilters)

	// For each filter kind in applied filters, precompute the union of method
	// sets of its values.  These are the methods that match just this filter.
	methodsByAppliedFilterKind := make(map[FilterKind]Set[MethodID], len(knownFilters))
	allFilterMethodSets := make([]Set[MethodID], 0, len(knownFilters))
	for filter, values := range knownFilters {
		ms := make(Set[MethodID], len(values))
		for _, v := range values {
			// Process non-directional filters.
			ms.Union(fm.filterMap.Get(filter, v))

			// Process directional filters.
			for _, direction := range []Direction{RequestDirection, ResponseDirection} {
				ms.Union(fm.filterMapByDirection.Get(direction, filter, v))
			}
		}
		methodsByAppliedFilterKind[filter] = ms
		allFilterMethodSets = append(allFilterMethodSets, ms)
	}

	// For each filter kind, precompute the intersection of the joined method
	// sets of all other filter kinds.  These are the methods that match all
	// the other applied filters.
	methodsIntersected := make(map[FilterKind]Set[MethodID], len(knownFilters))
	for filter := range knownFilters {
		toIntersect := make([]Set[MethodID], 0, len(knownFilters))
		for otherFilter := range knownFilters {
			if otherFilter == filter {
				continue
			}
			toIntersect = append(toIntersect, methodsByAppliedFilterKind[otherFilter])
		}

		// If there are other filters, use the intersection.
		if len(toIntersect) > 0 {
			methodsIntersected[filter] = Intersect(toIntersect...)
		} else {
			// If there are no other filters, then no methods are filtered out;
			// use all methods.
			methodsIntersected[filter] = fm.allMethods
		}
	}

	// Precompute the set of all methods after filters are applied.  If there
	// are no filters, this is the set of all methods.
	allFilteredMethods := fm.allMethods
	if len(knownFilters) > 0 {
		allFilteredMethods = Intersect(allFilterMethodSets...)
	}

	// Compute the summary by intersecting the method set for each filter
	// kind/value with the intersection of the method sets of all other
	// filter kinds.  This is the set of methods that match this filter as
	// well as all the applied filters having a different kind.
	summary := NewSummaryByDirection()

	// Process nondirected filters.
	for filter, byValue := range fm.filterMap {
		methodsMatchingOtherFilters, ok := methodsIntersected[filter]
		if !ok {
			// This filter is not currently applied, so intersect with
			// allFilteredMethods.
			methodsMatchingOtherFilters = allFilteredMethods
		}

		for value, methods := range byValue {
			summary.NondirectedFilters.Insert(filter, value, len(Intersect(methods, methodsMatchingOtherFilters)))
		}
	}

	// Process directed filters.
	for direction, byKind := range fm.filterMapByDirection {
		for filter, byValue := range byKind {
			methodsMatchingOtherFilters, ok := methodsIntersected[filter]
			if !ok {
				// This filter is not currently applied, so intersect with
				// allFilteredMethods.
				methodsMatchingOtherFilters = allFilteredMethods
			}

			for value, methods := range byValue {
				summary.DirectedFilters.Insert(direction, filter, value, len(Intersect(methods, methodsMatchingOtherFilters)))
			}
		}
	}

	return summary, allFilteredMethods
}

// Filter kind -> filter value -> method set.
type FilterMap[MethodID comparable] map[FilterKind]map[FilterValue]Set[MethodID]

func (fm FilterMap[MethodID]) Insert(filterKind FilterKind, filterValue string, method MethodID) {
	methodsByFilterValue, ok := fm[filterKind]
	if !ok {
		methodsByFilterValue = make(map[string]Set[MethodID])
		fm[filterKind] = methodsByFilterValue
	}

	methods, ok := methodsByFilterValue[filterValue]
	if !ok {
		methods = make(Set[MethodID])
		methodsByFilterValue[filterValue] = methods
	}

	methods.Insert(method)
}

// Returns the method set for kind, value or nil if the pair doesn't exist
// in the filter map.
func (fm FilterMap[MethodID]) Get(kind FilterKind, value FilterValue) Set[MethodID] {
	if byValue, ok := fm[kind]; ok {
		return byValue[value]
	}
	return nil
}

type FilterMapByDirection[MethodID comparable] map[Direction]FilterMap[MethodID]

func (fmd FilterMapByDirection[MethodID]) Get(direction Direction, kind FilterKind, value FilterValue) Set[MethodID] {
	byKind, ok := fmd[direction]
	if !ok {
		return nil
	}

	byVal, ok := byKind[kind]
	if !ok {
		return nil
	}

	return byVal[value]
}

func (fmd FilterMapByDirection[MethodID]) Insert(direction Direction, kind FilterKind, value string, method MethodID) {
	// Get or create the FilterMap for direction.
	fm, ok := fmd[direction]
	if !ok {
		fm = make(FilterMap[MethodID])
		fmd[direction] = fm
	}

	fm.Insert(kind, value, method)
}
