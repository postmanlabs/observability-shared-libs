package spec_summary

import "github.com/pkg/errors"

type FilterValue = string

type FilterKind string
type FilterValues []FilterValue
type Filters map[FilterKind]FilterValues

const (
	AuthFilter           FilterKind = "authentications"
	DirectionFilter      FilterKind = "directions"
	HostFilter           FilterKind = "hosts"
	HttpMethodFilter     FilterKind = "http_methods"
	PathFilter           FilterKind = "paths"
	ParamFilter          FilterKind = "params"
	PropertyFilter       FilterKind = "properties"
	ResponseCodeFilter   FilterKind = "response_codes"
	DataFormatFilter     FilterKind = "data_formats"
	DataFormatKindFilter FilterKind = "data_kinds"
	DataTypeFilter       FilterKind = "data_types"
	UnknownFilter        FilterKind = "unknown"
)

func ParseFilterKind(kind string) (FilterKind, error) {
	switch FilterKind(kind) {
	case AuthFilter:
		return AuthFilter, nil
	case DirectionFilter:
		return DirectionFilter, nil
	case HostFilter:
		return HostFilter, nil
	case HttpMethodFilter:
		return HttpMethodFilter, nil
	case PathFilter:
		return PathFilter, nil
	case ParamFilter:
		return ParamFilter, nil
	case PropertyFilter:
		return PropertyFilter, nil
	case ResponseCodeFilter:
		return ResponseCodeFilter, nil
	case DataFormatFilter:
		return DataFormatFilter, nil
	case DataFormatKindFilter:
		return DataFormatKindFilter, nil
	case DataTypeFilter:
		return DataTypeFilter, nil
	default:
		return UnknownFilter, errors.Errorf("unknown filter: %s", kind)
	}
}

// Return a new map of filters containing known filter kinds.  Ignore
// keys that are not known filter kinds.
func ParseFiltersIgnoreErrors(filters map[string][]string) Filters {
	knownFilters := make(Filters, len(filters))

	for filter, values := range filters {
		// Ignore unknown filters.
		if kind, err := ParseFilterKind(filter); err == nil {
			knownFilters[kind] = FilterValues(values)
		}
	}

	return knownFilters
}
