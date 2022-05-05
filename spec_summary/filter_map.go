package spec_summary

import (
	. "github.com/akitasoftware/go-utils/sets"
)

// Maps filter kinds/values to the methods that match them.
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

// Compute a summary that reflects filters that have already been applied, as
// well as the set of methods that match the applied filters.
// The count for a given filter value is calculated as the number of
// methods that match it, assuming
// - no other values of the same filter are applied
// - all other filters are applied.
//
// For example, if the current filters are http_method=GET and response_code=200,
// then the count for response_code=404 is calculated as the number of methods
// with a 404 response code a GET http method.
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
			// Process directed filters.
			ms.Union(fm.filterMap.Get(filter, v))

			// Process nondirected filters.
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
	// well as all the other applied filters.
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
