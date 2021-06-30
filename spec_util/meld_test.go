package spec_util

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/akitasoftware/akita-libs/test"
)

type testData struct {
	name                string
	witnessFiles        []string
	expectedWitnessFile string
}

var tests = []testData{
	{
		"no format, format",
		[]string{
			"testdata/meld/meld_no_data_formats.pb.txt",
			"testdata/meld/meld_data_formats_1.pb.txt",
		},
		"testdata/meld/meld_data_formats_1.pb.txt",
	},
	{
		"format, format",
		[]string{
			"testdata/meld/meld_data_formats_1.pb.txt",
			"testdata/meld/meld_data_formats_2.pb.txt",
		},
		"testdata/meld/meld_data_formats_3.pb.txt",
	},
	{
		"format, format with conflict",
		[]string{
			"testdata/meld/meld_conflict_1.pb.txt",
			"testdata/meld/meld_conflict_2.pb.txt",
		},
		"testdata/meld/meld_conflict_expected.pb.txt",
	},
	{
		"duplicate format dropped",
		[]string{
			"testdata/meld/meld_data_formats_1.pb.txt",
			"testdata/meld/meld_data_formats_2.pb.txt",
			"testdata/meld/meld_data_formats_1.pb.txt",
			"testdata/meld/meld_data_formats_2.pb.txt",
		},
		"testdata/meld/meld_data_formats_3.pb.txt",
	},
	{
		"duplicate format kind dropped",
		[]string{
			"testdata/meld/meld_data_kind_1.pb.txt",
			"testdata/meld/meld_data_kind_2.pb.txt",
			"testdata/meld/meld_data_kind_1.pb.txt",
			"testdata/meld/meld_data_kind_2.pb.txt",
		},
		"testdata/meld/meld_data_kind_expected.pb.txt",
	},
	{
		"meld into existing conflict",
		[]string{
			"testdata/meld/meld_with_existing_conflict_1.pb.txt",
			"testdata/meld/meld_with_existing_conflict_2.pb.txt",
		},
		"testdata/meld/meld_with_existing_conflict_expected.pb.txt",
	},
	{
		"turn conflict with none into optional - order 1",
		[]string{
			"testdata/meld/meld_suppress_none_conflict_1.pb.txt",
			"testdata/meld/meld_suppress_none_conflict_2.pb.txt",
		},
		"testdata/meld/meld_suppress_none_conflict_expected.pb.txt",
	},
	{
		// Make sure none is suppressed if it's not the first value that we process.
		"turn conflict with none into optional - order 2",
		[]string{
			"testdata/meld/meld_suppress_none_conflict_2.pb.txt",
			"testdata/meld/meld_suppress_none_conflict_1.pb.txt",
		},
		"testdata/meld/meld_suppress_none_conflict_expected.pb.txt",
	},
	{
		// Test meld(T, optional<T>) => optional<T>
		"meld optional and non-optional versions of the same type",
		[]string{
			"testdata/meld/meld_optional_required_1.pb.txt",
			"testdata/meld/meld_optional_required_2.pb.txt",
		},
		"testdata/meld/meld_optional_required_2.pb.txt",
	},
	{
		// meld(oneof(T1, T2), oneof(T1, T3)) => oneof(T1, T2, T3)
		"meld additive oneof",
		[]string{
			"testdata/meld/meld_additive_oneof_1.pb.txt",
			"testdata/meld/meld_additive_oneof_2.pb.txt",
		},
		"testdata/meld/meld_additive_oneof_expected.pb.txt",
	},
	{
		// meld(oneof(T1, T2), T3) => oneof(T1, T2, T3)
		"meld additive oneof with primitive",
		[]string{
			"testdata/meld/meld_oneof_with_primitive_1.pb.txt",
			"testdata/meld/meld_oneof_with_primitive_2.pb.txt",
		},
		"testdata/meld/meld_oneof_with_primitive_expected.pb.txt",
	},
	{
		"meld struct",
		[]string{
			"testdata/meld/meld_struct_1.pb.txt",
			"testdata/meld/meld_struct_2.pb.txt",
		},
		"testdata/meld/meld_struct_2.pb.txt",
	},
	{
		"meld list",
		[]string{
			"testdata/meld/meld_list_1.pb.txt",
			"testdata/meld/meld_list_2.pb.txt",
		},
		"testdata/meld/meld_list_2.pb.txt",
	},
	{
		"example, example",
		[]string{
			"testdata/meld/meld_examples_1.pb.txt",
			"testdata/meld/meld_examples_2.pb.txt",
		},
		"testdata/meld/meld_examples_3.pb.txt",
	},
	{
		"3 examples, 3 examples",
		[]string{
			"testdata/meld/meld_examples_big_1.pb.txt",
			"testdata/meld/meld_examples_big_2.pb.txt",
		},
		"testdata/meld/meld_examples_big_3.pb.txt",
	},
	{
		"1 example, 0 examples",
		[]string{
			"testdata/meld/meld_no_examples_1.pb.txt",
			"testdata/meld/meld_no_examples_2.pb.txt",
		},
		"testdata/meld/meld_no_examples_3.pb.txt",
	},
	{
		"optional field",
		[]string{
			"testdata/meld/meld_optional_1.pb.txt",
			"testdata/meld/meld_optional_2.pb.txt",
		},
		"testdata/meld/meld_optional_expected.pb.txt",
	},
	// Test melding non-4xx with 4xx.
	{
		"4xx example, example",
		[]string{
			"testdata/meld/meld_examples_1.pb.txt",
			"testdata/meld/meld_examples_4xx_2.pb.txt",
		},
		"testdata/meld/meld_non_4xx_with_4xx_expected.pb.txt",
	},
	// Test melding 4xx with 4xx.
	{
		"4xx example, 4xx example",
		[]string{
			"testdata/meld/meld_examples_4xx_1.pb.txt",
			"testdata/meld/meld_examples_4xx_2.pb.txt",
		},
		"testdata/meld/meld_4xx_expected.pb.txt",
	},
	// Test melding request-only with 4xx. We should get the request from the first, paired with the response from the second.
	{
		"no response, 4xx example",
		[]string{
			"testdata/meld/meld_no_response.pb.txt",
			"testdata/meld/meld_examples_4xx_2.pb.txt",
		},
		"testdata/meld/meld_no_response_4xx_expected.pb.txt",
	},
	// Test melding request-only with 4xx with 4xx and non-4xx. We should get the requests from the first and third, paired with both responses.
	{
		"no response, 4xx example, full non-4xx",
		[]string{
			"testdata/meld/meld_no_response.pb.txt",
			"testdata/meld/meld_examples_4xx_2.pb.txt",
			"testdata/meld/meld_examples_2.pb.txt",
		},
		"testdata/meld/meld_no_response_4xx_non_4xx_expected.pb.txt",
	},
}

func TestMeldWithFormats(t *testing.T) {
	for _, testData := range tests {
		expected := test.LoadWitnessFromFileOrDile(testData.expectedWitnessFile).Method

		// test right merged to left
		{
			result := NewMeldedMethod(test.LoadWitnessFromFileOrDile(testData.witnessFiles[0]).Method)
			for i := 1; i < len(testData.witnessFiles); i++ {
				newWitness := test.LoadWitnessFromFileOrDile(testData.witnessFiles[i])
				assert.NoError(t, result.Meld(NewMeldedMethod(newWitness.Method)))
			}
			if diff := cmp.Diff(expected, result.GetMethod(), cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("[%s] right merged to left\n%v", testData.name, diff)
				continue
			}
		}

		// test left merged to right
		{
			l := len(testData.witnessFiles)
			result := NewMeldedMethod(test.LoadWitnessFromFileOrDile(testData.witnessFiles[l-1]).Method)
			for i := l - 2; i >= 0; i-- {
				newWitness := test.LoadWitnessFromFileOrDile(testData.witnessFiles[i])
				assert.NoError(t, result.Meld(NewMeldedMethod(newWitness.Method)))
			}
			if diff := cmp.Diff(expected, result.GetMethod(), cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("[%s] left merged to right\n%v", testData.name, diff)
				continue
			}
		}
	}
}
