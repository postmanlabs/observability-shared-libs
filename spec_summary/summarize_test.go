package spec_summary

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/test"
)

func TestSummarize(t *testing.T) {
	testCases := []struct {
		name     string
		specFile string
		filters  map[string][]string
		expected *Summary
	}{
		{
			name:     "summary without filters",
			specFile: "testdata/spec1.pb.txt",
			filters:  nil,
			expected: &Summary{
				Authentications: map[string]int{
					"BASIC": 2,
				},
				Directions: map[string]int{
					"request":  2,
					"response": 2,
				},
				HTTPMethods: map[string]int{
					"POST": 2,
				},
				Paths: map[string]int{
					"/v1/projects/{arg3}": 1,
					"/v1/users/{arg3}": 1,
				},
				Params: map[string]int{
					"X-My-Header": 2,
				},
				Properties: map[string]int{
					"top-level-prop":       2,
					"my-special-prop":      2,
					"other-top-level-prop": 2,
				},
				ResponseCodes: map[string]int{
					"200": 1,
					"201": 1,
				},
				Hosts: map[string]int{
					"example.com": 1,
					"other-example.com": 1,
				},
				DataFormats: map[string]int{
					"rfc3339": 2,
				},
				DataKinds: map[string]int{},
				DataTypes: map[string]int{
					"string": 2,
				},
			},
		},
		{
			name:     "summary with one filter",
			specFile: "testdata/spec1.pb.txt",
			filters:  map[string][]string{
				"hosts": {"example.com"},
			},
			expected: &Summary{
				Authentications: map[string]int{
					"BASIC": 1,
				},
				Directions: map[string]int{
					"request":  1,
					"response": 1,
				},
				HTTPMethods: map[string]int{
					"POST": 1,
				},
				Paths: map[string]int{
					"/v1/projects/{arg3}": 1,
					"/v1/users/{arg3}": 0,
				},
				Params: map[string]int{
					"X-My-Header": 1,
				},
				Properties: map[string]int{
					"top-level-prop":       1,
					"my-special-prop":      1,
					"other-top-level-prop": 1,
				},
				ResponseCodes: map[string]int{
					"200": 1,
					"201": 0,
				},
				Hosts: map[string]int{
					"example.com": 1,
					"other-example.com": 1,
				},
				DataFormats: map[string]int{
					"rfc3339": 1,
				},
				DataKinds: nil,
				DataTypes: map[string]int{
					"string": 1,
				},
			},
		},
	}

	for _, tc := range testCases {
		spec := test.LoadAPISpecFromFileOrDie(tc.specFile)
		assert.Equal(t, tc.expected, SummarizeWithFilters(spec, tc.filters), tc.name)
	}
}

func TestIntersect(t *testing.T) {
	m1 := test.LoadMethodFromFileOrDie("testdata/method1.pb.txt")
	m2 := test.LoadMethodFromFileOrDie("testdata/method1.pb.txt")

	setM1 := make(map[*pb.Method]struct{})
	setM1[m1] = struct{}{}

	setM2 := make(map[*pb.Method]struct{})
	setM2[m2] = struct{}{}

	setM12 := make(map[*pb.Method]struct{})
	setM12[m1] = struct{}{}
	setM12[m2] = struct{}{}

	emptyset := make(map[*pb.Method]struct{})

	assert.Equal(t, emptyset, intersect([]map[*pb.Method]struct{}{setM1, setM2}))
	assert.Equal(t, emptyset, intersect([]map[*pb.Method]struct{}{emptyset, setM2}))
	assert.Equal(t, emptyset, intersect([]map[*pb.Method]struct{}{setM1, emptyset}))
	assert.Equal(t, setM1, intersect([]map[*pb.Method]struct{}{setM1, setM12}))
	assert.Equal(t, setM1, intersect([]map[*pb.Method]struct{}{setM12, setM1}))
	assert.Equal(t, setM2, intersect([]map[*pb.Method]struct{}{setM2, setM12}))
	assert.Equal(t, setM2, intersect([]map[*pb.Method]struct{}{setM12, setM2}))
	assert.Equal(t, setM12, intersect([]map[*pb.Method]struct{}{setM12, setM12}))
}
