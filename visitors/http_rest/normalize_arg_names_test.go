package http_rest

import (
	"testing"

	"github.com/akitasoftware/akita-libs/test"
)

func TestNormalizeNames(t *testing.T) {
	spec := test.LoadAPISpecFromFileOrDie("../testdata/sentry_ir_spec.pb.txt")

	method := spec.Methods[5]
	methodMeta := method.GetMeta().GetHttp()

	normalizedNames, err := GetNormalizedArgNames(method.Args, methodMeta)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"Authorization": "arg-headers-0",
		"4":             "arg-path-0",
		"5":             "arg-path-1",
		"(body)":        "arg-body-0",
	}
	for k, v := range normalizedNames {
		e := expected[k.String()]
		if v != e {
			t.Errorf("Mismatch, expected %v -> %q, got %q", k, e, v)
		}
	}
}
