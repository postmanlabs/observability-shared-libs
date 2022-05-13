package spec_summary

import (
	"testing"

	"github.com/akitasoftware/akita-libs/akid"
	"github.com/akitasoftware/go-utils/sets"
	"github.com/stretchr/testify/assert"
)

func TestFiltersToMethods(t *testing.T) {
	fm := NewFiltersToMethods[string]()

	m1 := akid.GenerateAPIMethodID().GetUUID().String()
	m2 := akid.GenerateAPIMethodID().GetUUID().String()

	fm.InsertNondirectionalFilter(HostFilter, "example.com", m1)
	fm.InsertNondirectionalFilter(HostFilter, "example.com", m2)
	fm.InsertDirectionalFilter(RequestDirection, AuthFilter, "None", m1)
	fm.InsertDirectionalFilter(RequestDirection, AuthFilter, "None", m2)

	fm.filterMap.Insert(HttpMethodFilter, "GET", m1)
	fm.filterMap.Insert(HttpMethodFilter, "POST", m2)

	// No filters.
	directedSummary, methods := fm.SummarizeWithFilters(nil)

	assert.Equal(t, sets.NewSet(m1, m2), methods)
	assert.Equal(t, 2, directedSummary.NondirectedFilters[HostFilter]["example.com"], "directed: example")
	assert.Equal(t, 1, directedSummary.NondirectedFilters[HttpMethodFilter]["GET"], "directed: get")
	assert.Equal(t, 2, directedSummary.DirectedFilters[RequestDirection][AuthFilter]["None"], "directed: auth")

	summary := directedSummary.ToSummary()

	assert.Equal(t, 2, summary.Hosts["example.com"], "example")
	assert.Equal(t, 1, summary.HTTPMethods["GET"], "get")

	// With filters.
	directedSummary, methods = fm.SummarizeWithFilters(Filters{
		HttpMethodFilter: {"GET"},
	})

	assert.Equal(t, sets.NewSet(m1), methods)
	assert.Equal(t, 1, directedSummary.NondirectedFilters[HostFilter]["example.com"], "directed: example")
	assert.Equal(t, 1, directedSummary.NondirectedFilters[HttpMethodFilter]["GET"], "directed: get")
	assert.Equal(t, 0, directedSummary.NondirectedFilters[HttpMethodFilter]["PUT"], "directed: put")
	assert.Equal(t, 1, directedSummary.DirectedFilters[RequestDirection][AuthFilter]["None"], "directed: auth")

	summary = directedSummary.ToSummary()

	assert.Equal(t, 1, summary.Hosts["example.com"], "example")
	assert.Equal(t, 1, summary.HTTPMethods["GET"], "get")
	assert.Equal(t, 0, summary.HTTPMethods["PUT"], "put")
}
