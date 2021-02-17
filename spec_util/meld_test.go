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
		"meld additive oneof",
		[]string{
			"testdata/meld/meld_oneof_with_primitive_1.pb.txt",
			"testdata/meld/meld_oneof_with_primitive_2.pb.txt",
		},
		"testdata/meld/meld_oneof_with_primitive_expected.pb.txt",
	},
}

func TestMeldWithFormats(t *testing.T) {
	for _, testData := range tests {
		expected := test.LoadWitnessFromFileOrDile(testData.expectedWitnessFile)

		// test right merged to left
		{
			result := test.LoadWitnessFromFileOrDile(testData.witnessFiles[0])
			for i := 1; i < len(testData.witnessFiles); i++ {
				newWitness := test.LoadWitnessFromFileOrDile(testData.witnessFiles[i])
				assert.NoError(t, MeldMethod(result.Method, newWitness.Method))
			}
			if diff := cmp.Diff(expected, result, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("[%s] right merged to left\n%v", testData.name, diff)
				continue
			}
		}

		// test left merged to right
		{
			l := len(testData.witnessFiles)
			result := test.LoadWitnessFromFileOrDile(testData.witnessFiles[l-1])
			for i := l - 2; i >= 0; i-- {
				newWitness := test.LoadWitnessFromFileOrDile(testData.witnessFiles[i])
				assert.NoError(t, MeldMethod(result.Method, newWitness.Method))
			}
			if diff := cmp.Diff(expected, result, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("[%s] left merged to right\n%v", testData.name, diff)
				continue
			}
		}
	}
}
