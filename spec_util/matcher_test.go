package spec_util

import (
	"testing"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

func singleMethodSpec(operation string, host string, template string) *pb.APISpec {
	return &pb.APISpec{
		Methods: []*pb.Method{
			testMethodWithHost(operation, host, template),
		},
	}
}

func testMethodWithHost(operation string, host string, template string) *pb.Method {
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
					Host:         host,
				},
			},
		},
	}
}

func TestMethodMatching(t *testing.T) {
	matchConcreteOpts := MethodMatchOptions{
		MatchOperation:     true,
		MatchHost:          true,
		MatchConcretePaths: true,
	}
	matchConcreteOptsNoOperation := MethodMatchOptions{
		MatchHost:          true,
		MatchConcretePaths: true,
	}
	matchTemplateOpts := MethodMatchOptions{
		MatchOperation:     true,
		MatchHost:          true,
		MatchPathTemplates: true,
	}

	testCases := []struct {
		Name            string
		MethodOperation string
		MethodTemplate  string
		TestOperation   string
		TestPath        string
		Options         MethodMatchOptions
		ExpectedMatch   bool
	}{
		{
			"single match",
			"GET", "/v1/{service}/foo",
			"GET", "/v1/abcdef/foo",
			matchConcreteOpts,
			true,
		},
		{
			"different operation",
			"POST", "/v1/{service}/foo",
			"GET", "/v1/abcdef/foo",
			matchConcreteOpts,
			false,
		},
		{
			"different operation (ignore operation)",
			"POST", "/v1/{service}/foo",
			"GET", "/v1/abcdef/foo",
			matchConcreteOptsNoOperation,
			true,
		},
		{
			"missing component",
			"GET", "/v1/{service}/foo",
			"GET", "/v1/abcdef",
			matchConcreteOpts,
			false,
		},
		{
			"too many components",
			"GET", "/v1/{service}/foo",
			"GET", "/v1/abc/def/foo",
			matchConcreteOpts,
			false,
		},
		{
			"multiple matches",
			"GET", "/v1/{abc}/{def}",
			"GET", "/v1/abc/def",
			matchConcreteOpts,
			true,
		},
		{
			"too few matches",
			"GET", "/v1/{abc}/{def}",
			"GET", "/v1/abcdef",
			matchConcreteOpts,
			false,
		},
		{
			"matches with non-alphabetic characters",
			"GET", "/v.1/{abc}/{def}",
			"GET", "/v.1/a~c/d-f",
			matchConcreteOpts,
			true,
		},
		{
			"non-matches with non-alphabetic characters",
			"GET", "/v.1/{abc}/{def}",
			"GET", "/vx1/a.c/d.f",
			matchConcreteOpts,
			false,
		},
		{
			"match template as concrete",
			"GET", "/v1/{service}/foo",
			"GET", "/v1/{service}/foo",
			matchConcreteOpts,
			false,
		},
		{
			"match template as concrete (different arg name)",
			"GET", "/v1/{service}/foo",
			"GET", "/v1/{arg}/foo",
			matchConcreteOpts,
			false,
		},
		{
			"match template as template",
			"GET", "/v1/{service}/foo",
			"GET", "/v1/{service}/foo",
			matchTemplateOpts,
			true,
		},
		{
			"match template as template (different arg name)",
			"GET", "/v1/{service}/foo",
			"GET", "/v1/{arg}/foo",
			matchTemplateOpts,
			true,
		},
	}
	host := "localhost:5000"

	for _, tc := range testCases {
		m, err := NewMethodMatcher(singleMethodSpec(tc.MethodOperation, host, tc.MethodTemplate))
		if err != nil {
			t.Fatal(err)
		}
		actual, matched := m.Lookup(tc.TestOperation, host, tc.TestPath, tc.Options)
		if tc.ExpectedMatch {
			if actual != tc.MethodTemplate {
				t.Errorf("in case %q, expected template match but got %q", tc.Name, actual)
			}
			if !matched {
				t.Errorf("in case %q, expected template to match, but no match was found", tc.Name)
			}
		} else {
			if actual != tc.TestPath {
				t.Errorf("in case %q, expected original path but got %q", tc.Name, actual)
			}
			if matched {
				t.Errorf("in case %q, expected template to not match, but a match was found", tc.Name)
			}
		}
	}
}

func TestMultipleMethodMatching(t *testing.T) {
	host := "localhost:5000"
	spec := &pb.APISpec{
		Methods: []*pb.Method{
			testMethodWithHost("GET", host, "/users/{arg2}"),
			testMethodWithHost("POST", host, "/users/{arg2}/files"),
			testMethodWithHost("GET", host, "/users/{arg2}/files"),
			testMethodWithHost("GET", host, "/users/{arg2}/files/{arg4}"),
		},
	}
	m, err := NewMethodMatcher(spec)
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		TestOperation   string
		TestPath        string
		ExpectedMatch   string
		ExpectedMatched bool
	}{
		{
			"GET",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d93ba",
			"/users/{arg2}",
			true,
		},
		{
			"POST",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d93ba/files",
			"/users/{arg2}/files",
			true,
		},
		{
			"GET",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d93ba/files",
			"/users/{arg2}/files",
			true,
		},
		{
			"GET",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d93ba/files/7b1ddce4-9d70-11eb-9870-0bc4cfc23f34",
			"/users/{arg2}/files/{arg4}",
			true,
		},
		{
			"POST",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d93ba/files/7b1ddce4-9d70-11eb-9870-0bc4cfc23f34",
			"/users/{arg2}/files/{arg4}",
			true,
		},
	}
	for _, tc := range testCases {
		actual, matched := m.LookupWithHost(tc.TestOperation, host, tc.TestPath)
		if actual != tc.ExpectedMatch {
			t.Errorf("expected %q but got %q for input %s %s", tc.ExpectedMatch, actual, tc.TestOperation, tc.TestPath)
		}
		if tc.ExpectedMatched && !matched {
			t.Errorf("expected a match for input %s %s, but got no match", tc.TestOperation, tc.TestPath)
		}
		if !tc.ExpectedMatched && matched {
			t.Errorf("expected no match for input %s %s, but got a match", tc.TestOperation, tc.TestPath)
		}
	}
}

func TestHostMatching(t *testing.T) {
	spec := &pb.APISpec{
		Methods: []*pb.Method{
			testMethodWithHost("GET", "api-server", "/users/{arg2}/files"),
			testMethodWithHost("GET", "api-server", "/users/{arg2}"),
			testMethodWithHost("GET", "api-server:8000", "/users/{xyz}/files"),
		},
	}
	m, err := NewMethodMatcher(spec)
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		TestOperation   string
		TestHost        string
		TestPath        string
		ExpectedMatch   string
		ExpectedMatched bool
	}{
		{
			"GET",
			"localhost",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9111",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9111",
			false,
		},
		{
			"GET",
			"api-server",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9222",
			"/users/{arg2}",
			true,
		},
		{
			"GET",
			"api-server",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9333/files",
			"/users/{arg2}/files",
			true,
		},
		{
			"GET",
			"api-server",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9444/other",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9444/other",
			false,
		},
		{
			"GET",
			"api-server:8000",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9555/files",
			"/users/{xyz}/files",
			true,
		},
		{
			// this case now falls back to the GET path
			"POST",
			"api-server:8000",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9555/files",
			"/users/{xyz}/files",
			true,
		},
		{
			"GET",
			"api-server:8000",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9666",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9666",
			false,
		},
	}

	for _, tc := range testCases {
		actual, matched := m.LookupWithHost(tc.TestOperation, tc.TestHost, tc.TestPath)
		if actual != tc.ExpectedMatch {
			t.Errorf("expected %q but got %q for input %s %s%s", tc.ExpectedMatch, actual, tc.TestOperation, tc.TestHost, tc.TestPath)
		}
		if tc.ExpectedMatched && !matched {
			t.Errorf("expected a match for input %s %s%s, but got no match", tc.TestOperation, tc.TestHost, tc.TestPath)
		}
		if !tc.ExpectedMatched && matched {
			t.Errorf("expected no match for input %s %s%s, but got a match", tc.TestOperation, tc.TestHost, tc.TestPath)
		}
	}
}

func TestMoreSpecificMatching(t *testing.T) {
	spec := &pb.APISpec{
		Methods: []*pb.Method{
			testMethodWithHost("GET", "api-server", "/users/{arg2}/files/{arg4}"),
			testMethodWithHost("GET", "api-server", "/users/admin/files/{arg4}"),
			testMethodWithHost("GET", "api-server", "/users/admin/files/foo"),
			testMethodWithHost("GET", "api-server", "/users/{arg2}/files/bar"),
			testMethodWithHost("GET", "api-server", "/users/{arg2}/{arg3}/{arg4}"),
		},
	}
	m, err := NewMethodMatcher(spec)
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		TestOperation   string
		TestHost        string
		TestPath        string
		ExpectedMatch   string
		ExpectedMatched bool
	}{
		{
			"GET",
			"api-server",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9111",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9111",
			false,
		},
		{
			"GET",
			"api-server",
			"/users/2b9046ac-6112-11eb-ae07-3e22fb0d9111/files/abcdef",
			"/users/{arg2}/files/{arg4}",
			true,
		},
		{
			"GET",
			"api-server",
			"/users/admin/files/abcdef",
			"/users/admin/files/{arg4}",
			true,
		},
		{
			"GET",
			"api-server",
			"/users/admin/files/foo",
			"/users/admin/files/foo",
			true,
		},
		{
			"GET",
			"api-server",
			"/users/mark/directories/bar",
			"/users/{arg2}/{arg3}/{arg4}",
			true,
		},
		{
			"GET",
			"api-server",
			"/users/mark/files/bar",
			"/users/{arg2}/files/bar",
			true,
		},
		{
			"GET",
			"api-server",
			"/users/mark/files/foo",
			"/users/{arg2}/files/{arg4}",
			true,
		},
	}

	for _, tc := range testCases {
		actual, matched := m.LookupWithHost(tc.TestOperation, tc.TestHost, tc.TestPath)
		if actual != tc.ExpectedMatch {
			t.Errorf("expected %q but got %q for input %s %s%s", tc.ExpectedMatch, actual, tc.TestOperation, tc.TestHost, tc.TestPath)
		}
		if tc.ExpectedMatched && !matched {
			t.Errorf("expected a match for input %s %s%s, but got no match", tc.TestOperation, tc.TestHost, tc.TestPath)
		}
		if !tc.ExpectedMatched && matched {
			t.Errorf("expected no match for input %s %s%s, but got a match", tc.TestOperation, tc.TestHost, tc.TestPath)
		}
	}
}
