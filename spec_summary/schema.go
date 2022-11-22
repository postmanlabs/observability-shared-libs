package spec_summary

import (
	"strings"

	"github.com/pkg/errors"
)

// Summarizes a spec along different dimensions that can be used to filter for
// parts of the spec.
type DetailedSummary struct {
	Authentications map[FilterValue]int `json:"authentications"`
	Directions      map[FilterValue]int `json:"directions"`
	Hosts           map[FilterValue]int `json:"hosts"`
	HTTPMethods     map[FilterValue]int `json:"http_methods"`
	Paths           map[FilterValue]int `json:"paths"`
	Params          map[FilterValue]int `json:"params"`
	Properties      map[FilterValue]int `json:"properties"`
	ResponseCodes   map[FilterValue]int `json:"response_codes"`
	DataFormats     map[FilterValue]int `json:"data_formats"`
	DataKinds       map[FilterValue]int `json:"data_kinds"`
	DataTypes       map[FilterValue]int `json:"data_types"`
}

func NewDetailedSummary() *DetailedSummary {
	return &DetailedSummary{
		Authentications: make(map[FilterValue]int),
		Directions:      make(map[FilterValue]int),
		Hosts:           make(map[FilterValue]int),
		HTTPMethods:     make(map[FilterValue]int),
		Paths:           make(map[FilterValue]int),
		Params:          make(map[FilterValue]int),
		Properties:      make(map[FilterValue]int),
		ResponseCodes:   make(map[FilterValue]int),
		DataFormats:     make(map[FilterValue]int),
		DataKinds:       make(map[FilterValue]int),
		DataTypes:       make(map[FilterValue]int),
	}
}

// Summarizes a spec along different dimensions that can be used to filter for
// parts of the spec, distinguishing between requests and responses.
type SummaryByDirection struct {
	// Filters that discriminate by direction, i.e. whether the filter value
	// is in the request or response.
	DirectedFilters DirectedFilterCounts `json:"directed_filters"`

	// Filters that are independent of the request/response.
	NondirectedFilters NondirectedFilterCounts `json:"nondirected_filters"`
}

func NewSummaryByDirection() *SummaryByDirection {
	return &SummaryByDirection{
		DirectedFilters: DirectedFilterCounts{
			RequestDirection:  make(NondirectedFilterCounts),
			ResponseDirection: make(NondirectedFilterCounts),
		},
		NondirectedFilters: make(map[FilterKind]map[FilterValue]int),
	}
}

func (s *SummaryByDirection) ToSummary() *DetailedSummary {
	if s == nil {
		return nil
	}

	return &DetailedSummary{
		// Non-directional properties.
		Directions:  s.NondirectedFilters[DirectionFilter],
		Hosts:       s.NondirectedFilters[HostFilter],
		HTTPMethods: s.NondirectedFilters[HttpMethodFilter],
		Paths:       s.NondirectedFilters[PathFilter],

		// Directional properties.
		Authentications: s.DirectedFilters.GetCountsByValue(RequestDirection, AuthFilter),
		DataKinds:       s.DirectedFilters.MergeAcrossDirections(DataFormatKindFilter),
		DataFormats:     s.DirectedFilters.MergeAcrossDirections(DataFormatFilter),
		DataTypes:       s.DirectedFilters.MergeAcrossDirections(DataTypeFilter),
		Params:          s.DirectedFilters.MergeAcrossDirections(ParamFilter),
		Properties:      s.DirectedFilters.MergeAcrossDirections(PropertyFilter),
		ResponseCodes:   s.DirectedFilters.GetCountsByValue(ResponseDirection, ResponseCodeFilter),
	}
}

type NondirectedFilterCounts map[FilterKind]map[FilterValue]int

func (cs NondirectedFilterCounts) Increment(kind FilterKind, v FilterValue) {
	byVal, ok := cs[kind]
	if !ok {
		byVal = make(map[FilterValue]int)
		cs[kind] = byVal
	}

	byVal[v] += 1
}

func (cs NondirectedFilterCounts) Insert(kind FilterKind, v FilterValue, count int) {
	byVal, ok := cs[kind]
	if !ok {
		byVal = make(map[FilterValue]int)
		cs[kind] = byVal
	}

	byVal[v] = count
}

// Calls f on each kind–value pair. If f returns false, iteration stops
// immediately, and false is returned. Otherwise, returns true.
func (cs NondirectedFilterCounts) ForEach(f func(kind FilterKind, v FilterValue, count int) bool) bool {
	for kind, byVal := range cs {
		for v, count := range byVal {
			if !f(kind, v, count) {
				return false
			}
		}
	}

	return true
}

type DirectedFilterCounts map[Direction]NondirectedFilterCounts

func (cs DirectedFilterCounts) Increment(direction Direction, kind FilterKind, v FilterValue) {
	byKind, ok := cs[direction]
	if !ok {
		byKind = make(NondirectedFilterCounts)
		cs[direction] = byKind
	}

	byKind.Increment(kind, v)
}

func (cs DirectedFilterCounts) Insert(direction Direction, kind FilterKind, v FilterValue, count int) {
	byKind, ok := cs[direction]
	if !ok {
		byKind = make(NondirectedFilterCounts)
		cs[direction] = byKind
	}

	byKind.Insert(kind, v, count)
}

// Calls f on each (direction, kind, value) tuple. If f returns false, iteration
// stops immediately, and false is returned. Otherwise, returns true.
func (cs DirectedFilterCounts) ForEach(f func(direction Direction, kind FilterKind, v FilterValue, count int) bool) bool {
	for direction, byKind := range cs {
		df := func(kind FilterKind, v FilterValue, count int) bool {
			return f(direction, kind, v, count)
		}
		if cont := byKind.ForEach(df); !cont {
			return false
		}
	}
	return true
}

func (cs DirectedFilterCounts) GetCountsByValue(direction Direction, kind FilterKind) map[FilterValue]int {
	byKind, ok := cs[direction]
	if !ok {
		return nil
	}

	return byKind[kind]
}

func (cs DirectedFilterCounts) MergeAcrossDirections(kind FilterKind) map[FilterValue]int {
	merged := make(map[FilterValue]int)

	for _, byKind := range cs {
		if byVal, ok := byKind[kind]; ok {
			for v, count := range byVal {
				merged[v] += count
			}
		}
	}

	if len(merged) == 0 {
		return nil
	}
	return merged
}

type Direction string

const (
	RequestDirection  Direction = "REQUEST"
	ResponseDirection Direction = "RESPONSE"
)

func ParseDirection(d string) (Direction, error) {
	switch Direction(strings.ToUpper(d)) {
	case RequestDirection:
		return RequestDirection, nil
	case ResponseDirection:
		return ResponseDirection, nil
	default:
		return "UNKNOWN_DIRECTION", errors.Errorf("unknown direction: %s", d)
	}
}
