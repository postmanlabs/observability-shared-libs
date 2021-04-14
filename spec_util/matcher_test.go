package spec_util

import (
	"testing"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

func singleMethodSpec(operation string, template string) *pb.APISpec {
	return &pb.APISpec{
		Methods: []*pb.Method{
			testMethod(operation, template),
		},
	}
}

func testMethod(operation string, template string) *pb.Method {
	return &pb.Method{
		Id: &pb.MethodID{
			Name:    "fake_name",
			ApiType: pb.ApiType_HTTP_REST,
		},
		Meta: &pb.MethodMeta{
			Meta: &pb.MethodMeta_Http{
				Http: &pb.HTTPMethodMeta{
					Method:       operation,
					PathTemplate: template,
					Host:         "",
				},
			},
		},
	}
}

func TestMethodMatching(t *testing.T) {
	testCases := []struct {
		Name            string
		MethodOperation string
		MethodTemplate  string
		TestOperation   string
		TestPath        string
		ExpectedMatch   bool
	}{
		{
			"single match",
			"GET", "/v1/{service}/foo",
			"GET", "/v1/abcdef/foo",
			true,
		},
		{
			"wrong operation",
			"POST", "/v1/{service}/foo",
			"GET", "/v1/abcdef/foo",
			false,
		},
		{
			"missing component",
			"GET", "/v1/{service}/foo",
			"GET", "/v1/abcdef",
			false,
		},
		{
			"too many components",
			"GET", "/v1/{service}/foo",
			"GET", "/v1/abc/def/foo",
			false,
		},
		{
			"multiple matches",
			"GET", "/v1/{abc}/{def}",
			"GET", "/v1/abc/def",
			true,
		},
		{
			"too few matches",
			"GET", "/v1/{abc}/{def}",
			"GET", "/v1/abcdef",
			false,
		},
		{
			"matches with non-alphabetic characters",
			"GET", "/v.1/{abc}/{def}",
			"GET", "/v.1/a~c/d-f",
			true,
		},
		{
			"non-matches with non-alphabetic characters",
			"GET", "/v.1/{abc}/{def}",
			"GET", "/vx1/a.c/d.f",
			false,
		},
	}

	for _, tc := range testCases {
		m, err := NewMethodMatcher(singleMethodSpec(tc.MethodOperation, tc.MethodTemplate))
		if err != nil {
			t.Fatal(err)
		}
		actual := m.Lookup(tc.TestOperation, tc.TestPath)
		if tc.ExpectedMatch {
			if actual != tc.MethodTemplate {
				t.Errorf("in case %q, expected template match but got %q", tc.Name, actual)
			}
		} else {
			if actual != tc.TestPath {
				t.Errorf("in case %q, expected original path but got %q", tc.Name, actual)
			}
		}
	}
}

func TestMultipleMethodMatching(t *testing.T) {
	spec := &pb.APISpec{
		Methods: []*pb.Method{
			testMethod("GET", "/users/{arg2}"),
			testMethod("POST", "/users/{arg2}/files"),
			testMethod("GET", "/users/{arg2}/files"),
			testMethod("GET", "/users/{arg2}/files/{arg4}"),
		},
	}
	m, err := NewMethodMatcher(spec)
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		TestOperation string
		TestPath      string
		ExpectedMatch string
	}{
		{
			"GET",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d93ba",
			"/users/{arg2}",
		},
		{
			"POST",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d93ba/files",
			"/users/{arg2}/files",
		},
		{
			"GET",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d93ba/files",
			"/users/{arg2}/files",
		},
		{
			"GET",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d93ba/files/7b1ddce4-9d70-11eb-9870-0bc4cfc23f34",
			"/users/{arg2}/files/{arg4}",
		},
		{
			"POST",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d93ba/files/7b1ddce4-9d70-11eb-9870-0bc4cfc23f34",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d93ba/files/7b1ddce4-9d70-11eb-9870-0bc4cfc23f34",
		},
	}
	for _, tc := range testCases {
		actual := m.Lookup(tc.TestOperation, tc.TestPath)
		if actual != tc.ExpectedMatch {
			t.Errorf("expected %q but got %q for input %s %s", tc.ExpectedMatch, actual, tc.TestOperation, tc.TestPath)
		}
	}
}
