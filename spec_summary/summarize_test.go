package spec_summary

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/test"
)

func TestSummarize(t *testing.T) {
	expected := &Summary{
		Authentications: map[string]int{
			"BASIC": 1,
		},
		HTTPMethods: map[string]int{
			"POST": 1,
		},
		Paths: map[string]int{
			"/v1/projects/{arg3}": 1,
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
		},
		DataFormats: map[string]int{
			"rfc3339": 1,
		},
		DataKinds: map[string]int{},
		DataTypes: map[string]int{
			"string": 1,
		},
	}

	m1 := test.LoadMethodFromFileOrDie("testdata/method1.pb.txt")
	assert.Equal(t, expected, Summarize(&pb.APISpec{Methods: []*pb.Method{m1}}))
}
