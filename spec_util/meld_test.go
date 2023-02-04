package spec_util

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"

	"github.com/akitasoftware/akita-libs/test"
)

type testData struct {
	name                string
	witnessFiles        []string
	opts                MeldOptions
	expectedWitnessFile string
}

var tests = []testData{
	{
		"no format, format",
		[]string{
			"testdata/meld/meld_no_data_formats.pb.txt",
			"testdata/meld/meld_data_formats_1.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_no_data_formats.pb.txt",
	},
	{
		"format, format",
		[]string{
			"testdata/meld/meld_data_formats_1.pb.txt",
			"testdata/meld/meld_data_formats_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_data_formats_3.pb.txt",
	},
	{
		"format, format with conflict",
		[]string{
			"testdata/meld/meld_conflict_1.pb.txt",
			"testdata/meld/meld_conflict_2.pb.txt",
		},
		MeldOptions{},
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
		MeldOptions{},
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
		MeldOptions{},
		"testdata/meld/meld_data_kind_expected.pb.txt",
	},
	{
		"meld into existing conflict",
		[]string{
			"testdata/meld/meld_with_existing_conflict_1.pb.txt",
			"testdata/meld/meld_with_existing_conflict_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_with_existing_conflict_expected.pb.txt",
	},
	{
		"turn conflict with none into nullable - order 1",
		[]string{
			"testdata/meld/meld_suppress_none_conflict_1.pb.txt",
			"testdata/meld/meld_suppress_none_conflict_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_suppress_none_conflict_expected.pb.txt",
	},
	{
		// Make sure none is suppressed if it's not the first value that we process.
		"turn conflict with none into optional - order 2",
		[]string{
			"testdata/meld/meld_suppress_none_conflict_2.pb.txt",
			"testdata/meld/meld_suppress_none_conflict_1.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_suppress_none_conflict_expected.pb.txt",
	},
	{
		// Test meld(T, optional<T>) => optional<T>
		"meld optional and non-optional versions of the same type",
		[]string{
			"testdata/meld/meld_optional_required_1.pb.txt",
			"testdata/meld/meld_optional_required_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_optional_required_2.pb.txt",
	},
	{
		// meld(oneof(T1, T2), oneof(T1, T3)) => oneof(T1, T2, T3)
		"meld additive oneof",
		[]string{
			"testdata/meld/meld_additive_oneof_1.pb.txt",
			"testdata/meld/meld_additive_oneof_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_additive_oneof_expected.pb.txt",
	},
	{
		// meld(oneof(T1, T2), T3) => oneof(T1, T2, T3)
		"meld additive oneof with primitive",
		[]string{
			"testdata/meld/meld_oneof_with_primitive_1.pb.txt",
			"testdata/meld/meld_oneof_with_primitive_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_oneof_with_primitive_expected.pb.txt",
	},
	{
		"melding oneof with itself is idempotent",
		[]string{
			"testdata/meld/meld_oneof_with_oneof_1.pb.txt",
			"testdata/meld/meld_oneof_with_oneof_1.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_oneof_with_oneof_1.pb.txt",
	},
	{
		// meld(oneof(list<L1>, S1), oneof(list<L2>, S2))
		//   => oneof(list<meld(L1, L2)>, meld(S1, S2))
		"meld oneof with oneof",
		[]string{
			"testdata/meld/meld_oneof_with_oneof_1.pb.txt",
			"testdata/meld/meld_oneof_with_oneof_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_oneof_with_oneof_expected.pb.txt",
	},
	{
		"meld struct",
		[]string{
			"testdata/meld/meld_struct_1.pb.txt",
			"testdata/meld/meld_struct_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_struct_2.pb.txt",
	},
	{
		"meld list",
		[]string{
			"testdata/meld/meld_list_1.pb.txt",
			"testdata/meld/meld_list_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_list_2.pb.txt",
	},
	{
		"example, example",
		[]string{
			"testdata/meld/meld_examples_1.pb.txt",
			"testdata/meld/meld_examples_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_examples_3.pb.txt",
	},
	{
		"3 examples, 3 examples",
		[]string{
			"testdata/meld/meld_examples_big_1.pb.txt",
			"testdata/meld/meld_examples_big_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_examples_big_3.pb.txt",
	},
	{
		"1 example, 0 examples",
		[]string{
			"testdata/meld/meld_no_examples_1.pb.txt",
			"testdata/meld/meld_no_examples_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_no_examples_3.pb.txt",
	},
	{
		"optional field",
		[]string{
			"testdata/meld/meld_optional_1.pb.txt",
			"testdata/meld/meld_optional_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_optional_expected.pb.txt",
	},
	// Test melding non-4xx with 4xx.
	{
		"4xx example, example",
		[]string{
			"testdata/meld/meld_examples_1.pb.txt",
			"testdata/meld/meld_examples_4xx_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_non_4xx_with_4xx_expected.pb.txt",
	},
	// Test melding 4xx with 4xx.
	{
		"4xx example, 4xx example",
		[]string{
			"testdata/meld/meld_examples_4xx_1.pb.txt",
			"testdata/meld/meld_examples_4xx_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_4xx_expected.pb.txt",
	},
	// Test melding request-only with 4xx. We should get the request from the first, paired with the response from the second.
	{
		"no response, 4xx example",
		[]string{
			"testdata/meld/meld_no_response.pb.txt",
			"testdata/meld/meld_examples_4xx_2.pb.txt",
		},
		MeldOptions{},
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
		MeldOptions{},
		"testdata/meld/meld_no_response_4xx_non_4xx_expected.pb.txt",
	},
	// Test conversion of structs into maps.
	{
		"map_1, map_2",
		[]string{
			"testdata/meld/meld_map_1.pb.txt",
			"testdata/meld/meld_map_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_map_1_map_2_expected.pb.txt",
	},
	// Test conversion of structs into maps with an optional none
	{
		"struct to map with none type",
		[]string{
			"testdata/meld/meld_map_1.pb.txt",
			"testdata/meld/meld_map_3.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_map_1_map_3_expected.pb.txt",
	},
	{
		"structs with number fields",
		[]string{
			"testdata/meld/meld_map_4.pb.txt",
			"testdata/meld/meld_map_5.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_map_4_map_5_expected.pb.txt",
	},
	{
		// Test meld(T, nullable<T>) => nullable<T>
		"meld nullable and non-nullable versions of the same type",
		[]string{
			"testdata/meld/meld_nullable_1.pb.txt",
			"testdata/meld/meld_nullable_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_nullable_expected.pb.txt",
	},
	{
		// Test meld(None, optional<T>) => nullable<optional<T>>
		"meld none and optional string should be nullable optional string",
		[]string{
			"testdata/meld/meld_optional_none_1.pb.txt",
			"testdata/meld/meld_optional_none_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_optional_none_expected.pb.txt",
	},
	{
		"test meld(oneof<T1, T2>, nullable<T1>) => nullable<oneof<T1, T2>>",
		[]string{
			"testdata/meld/meld_oneof_with_nullable_primitive_1.pb.txt",
			"testdata/meld/meld_oneof_with_nullable_primitive_2.pb.txt",
		},
		MeldOptions{},
		"testdata/meld/meld_oneof_with_nullable_primitive_expected.pb.txt",
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
				assert.NoError(t, result.Meld(NewMeldedMethod(newWitness.Method), testData.opts))
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
				assert.NoError(t, result.Meld(NewMeldedMethod(newWitness.Method), testData.opts))
			}
			if diff := cmp.Diff(expected, result.GetMethod(), cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("[%s] left merged to right\n%v", testData.name, diff)
				continue
			}
		}
	}
}

func wrapPrim(p *pb.Primitive) *pb.Data {
	return &pb.Data{Value: &pb.Data_Primitive{Primitive: p}}
}

func wrapOneOf(o *pb.OneOf) *pb.Data {
	return &pb.Data{Value: &pb.Data_Oneof{Oneof: o}}
}

var (
	cmpWitnessOptions = []cmp.Option{
		cmp.Comparer(proto.Equal),
		cmpopts.SortSlices(witnessLess),
		cmpopts.EquateEmpty(),
	}
)

func witnessLess(w1, w2 *pb.Witness) bool {
	return proto.MarshalTextString(w1) < proto.MarshalTextString(w2)
}

func TestMeldPrimitives(t *testing.T) {
	testCases := []struct {
		name     string
		left     *pb.Primitive
		right    *pb.Primitive
		expected *pb.Data
	}{
		{
			name: "merge is idempotent - int32/int32",
			left: &pb.Primitive{
				Value: &pb.Primitive_Int32Value{Int32Value: &pb.Int32{}},
			},
			right: &pb.Primitive{
				Value: &pb.Primitive_Int32Value{Int32Value: &pb.Int32{}},
			},
			expected: wrapPrim(&pb.Primitive{
				Value: &pb.Primitive_Int32Value{Int32Value: &pb.Int32{}},
			}),
		},
		{
			name: "merge is idempotent - timestamp/timestamp",
			left: &pb.Primitive{
				Value:      &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
				FormatKind: "datetime",
				Formats: map[string]bool{
					"DateTimeSecondsSinceEpoch": true,
				},
			},
			right: &pb.Primitive{
				Value:      &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
				FormatKind: "datetime",
				Formats: map[string]bool{
					"DateTimeSecondsSinceEpoch": true,
				},
			},
			expected: wrapPrim(&pb.Primitive{
				Value:      &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
				FormatKind: "datetime",
				Formats: map[string]bool{
					"DateTimeSecondsSinceEpoch": true,
				},
			}),
		},
		{
			name: "merge unions data formats of the same kind",
			left: &pb.Primitive{
				Value:      &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
				FormatKind: "datetime",
				Formats: map[string]bool{
					"DateTimeSecondsSinceEpoch": true,
				},
			},
			right: &pb.Primitive{
				Value:      &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
				FormatKind: "datetime",
				Formats: map[string]bool{
					"DateTimeNanosecondsSinceEpoch": true,
				},
			},
			expected: wrapPrim(&pb.Primitive{
				Value:      &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
				FormatKind: "datetime",
				Formats: map[string]bool{
					"DateTimeSecondsSinceEpoch":     true,
					"DateTimeNanosecondsSinceEpoch": true,
				},
			}),
		},
		{
			name: "merge operation joins up the type lattice - int64/int64 + formats",
			left: &pb.Primitive{
				Value:      &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
				FormatKind: "datetime",
				Formats: map[string]bool{
					"timestamp": true,
				},
			},
			right: &pb.Primitive{
				Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
			},
			expected: wrapPrim(&pb.Primitive{
				Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
			}),
		},
		{
			name: "merge operation joins up the type lattice - int64 + formats/int64",
			left: &pb.Primitive{
				Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
			},
			right: &pb.Primitive{
				Value:      &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
				FormatKind: "datetime",
				Formats: map[string]bool{
					"timestamp": true,
				},
			},
			expected: wrapPrim(&pb.Primitive{
				Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
			}),
		},
		{
			name: "merge operation joins up the type lattice - int32/int64",
			left: &pb.Primitive{
				Value: &pb.Primitive_Int32Value{Int32Value: &pb.Int32{}},
			},
			right: &pb.Primitive{
				Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
			},
			expected: wrapPrim(&pb.Primitive{
				Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
			}),
		},
		{
			name: "merge operation joins up the type lattice - uint32/int32",
			left: &pb.Primitive{
				Value: &pb.Primitive_Uint32Value{Uint32Value: &pb.Uint32{}},
			},
			right: &pb.Primitive{
				Value: &pb.Primitive_Int32Value{Int32Value: &pb.Int32{}},
			},
			expected: wrapPrim(&pb.Primitive{
				Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
			}),
		},
		{
			name: "merge operation joins up the type lattice - uint32/float",
			left: &pb.Primitive{
				Value: &pb.Primitive_Uint32Value{Uint32Value: &pb.Uint32{}},
			},
			right: &pb.Primitive{
				Value: &pb.Primitive_FloatValue{FloatValue: &pb.Float{}},
			},
			expected: wrapPrim(&pb.Primitive{
				Value: &pb.Primitive_FloatValue{FloatValue: &pb.Float{}},
			}),
		},
		{
			name: "merge operation joins up the type lattice - float/uint32",
			left: &pb.Primitive{
				Value: &pb.Primitive_FloatValue{FloatValue: &pb.Float{}},
			},
			right: &pb.Primitive{
				Value: &pb.Primitive_Uint32Value{Uint32Value: &pb.Uint32{}},
			},
			expected: wrapPrim(&pb.Primitive{
				Value: &pb.Primitive_FloatValue{FloatValue: &pb.Float{}},
			}),
		},
		{
			name: "merge operation joins up the type lattice - uin64/double",
			left: &pb.Primitive{
				Value: &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}},
			},
			right: &pb.Primitive{
				Value: &pb.Primitive_DoubleValue{DoubleValue: &pb.Double{}},
			},
			expected: wrapPrim(&pb.Primitive{
				Value: &pb.Primitive_DoubleValue{DoubleValue: &pb.Double{}},
			}),
		},
		{
			name: "merge operation conflicts on unjoinable types - int64/uint64",
			left: &pb.Primitive{
				Value: &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}},
			},
			right: &pb.Primitive{
				Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
			},
			expected: wrapOneOf(&pb.OneOf{
				Options: map[string]*pb.Data{
					"d616v2O0iE8=": wrapPrim(&pb.Primitive{
						Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
					}),
					"va5tP-fnZF8=": wrapPrim(&pb.Primitive{
						Value: &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}},
					}),
				},
				PotentialConflict: true,
			}),
		},
		{
			name: "merging oneof with primitive - merge primitive into existing variant",
			left: &pb.Primitive{
				Value: &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}},
			},
			right: &pb.Primitive{
				Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
			},
			expected: wrapOneOf(&pb.OneOf{
				Options: map[string]*pb.Data{
					"d616v2O0iE8=": wrapPrim(&pb.Primitive{
						Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
					}),
					"va5tP-fnZF8=": wrapPrim(&pb.Primitive{
						Value: &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}},
					}),
				},
				PotentialConflict: true,
			}),
		},
	}

	for _, tc := range testCases {
		leftDst := wrapPrim(proto.Clone(tc.left).(*pb.Primitive))
		err := MeldData(leftDst, wrapPrim(tc.right))
		assert.NoError(t, err, "[%s] failed to meld", tc.name)

		if diff := cmp.Diff(tc.expected, leftDst, cmpWitnessOptions...); diff != "" {
			t.Errorf("[%s] (right to left) found diff:\n %v", tc.name, diff)
		}

		rightDst := wrapPrim(proto.Clone(tc.right).(*pb.Primitive))
		err = MeldData(rightDst, wrapPrim(tc.left))
		assert.NoError(t, err, "[%s] failed to meld", tc.name)

		if diff := cmp.Diff(tc.expected, rightDst, cmpWitnessOptions...); diff != "" {
			t.Errorf("[%s] (left to right) found diff:\n %v", tc.name, diff)
		}
	}
}

func TestMeldData(t *testing.T) {
	testCases := []struct {
		name     string
		left     *pb.Data
		right    *pb.Data
		expected *pb.Data
	}{
		{
			name: "merging oneof with primitive",
			left: wrapOneOf(&pb.OneOf{
				Options: map[string]*pb.Data{
					"d616v2O0iE8=": wrapPrim(&pb.Primitive{
						Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
					}),
					"va5tP-fnZF8=": wrapPrim(&pb.Primitive{
						Value: &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}},
					}),
				},
				PotentialConflict: true,
			}),
			right: wrapPrim(&pb.Primitive{
				Value:      &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
				FormatKind: "datetime",
				Formats:    map[string]bool{"timestamp": true},
			}),
			expected: wrapOneOf(&pb.OneOf{
				Options: map[string]*pb.Data{
					"d616v2O0iE8=": wrapPrim(&pb.Primitive{
						Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}},
					}),
					"va5tP-fnZF8=": wrapPrim(&pb.Primitive{
						Value: &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}},
					}),
				},
				PotentialConflict: true,
			}),
		},
		{
			name: "merging oneof with unformatted primitive - resulting in a primitive",
			left: wrapOneOf(&pb.OneOf{
				Options: map[string]*pb.Data{
					"1": wrapPrim(&pb.Primitive{
						Value:      &pb.Primitive_Uint32Value{Uint32Value: &pb.Uint32{}},
						FormatKind: "datetime",
						Formats:    map[string]bool{"timestamp_seconds_since_epoch": true},
					}),
					"2": wrapPrim(&pb.Primitive{
						Value:      &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}},
						FormatKind: "unique_id",
						Formats:    map[string]bool{"integer_id": true},
					}),
				},
				PotentialConflict: true,
			}),
			right: wrapPrim(&pb.Primitive{
				Value: &pb.Primitive_Uint32Value{Uint32Value: &pb.Uint32{}},
			}),
			expected: wrapPrim(&pb.Primitive{
				Value: &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}},
			}),
		},
		{
			name: "merging oneof with unformatted primitive - resulting in a oneof",
			left: wrapOneOf(&pb.OneOf{
				Options: map[string]*pb.Data{
					"1": wrapPrim(&pb.Primitive{
						Value:      &pb.Primitive_Uint32Value{Uint32Value: &pb.Uint32{}},
						FormatKind: "datetime",
						Formats:    map[string]bool{"timestamp_seconds_since_epoch": true},
					}),
					"2": wrapPrim(&pb.Primitive{
						Value:      &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}},
						FormatKind: "unique_id",
						Formats:    map[string]bool{"integer_id": true},
					}),
					"3": wrapPrim(&pb.Primitive{
						Value:      &pb.Primitive_StringValue{StringValue: &pb.String{}},
						FormatKind: "currency_name",
						Formats:    map[string]bool{"currency_abbreviation": true},
					}),
				},
				PotentialConflict: true,
			}),
			right: wrapPrim(&pb.Primitive{
				Value: &pb.Primitive_Uint32Value{Uint32Value: &pb.Uint32{}},
			}),
			expected: wrapOneOf(&pb.OneOf{
				Options: map[string]*pb.Data{
					"va5tP-fnZF8=": wrapPrim(&pb.Primitive{
						Value: &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}},
					}),
					"3": wrapPrim(&pb.Primitive{
						Value:      &pb.Primitive_StringValue{StringValue: &pb.String{}},
						FormatKind: "currency_name",
						Formats:    map[string]bool{"currency_abbreviation": true},
					}),
				},
				PotentialConflict: true,
			}),
		},
	}

	for _, tc := range testCases {
		leftDst := proto.Clone(tc.left).(*pb.Data)
		right := proto.Clone(tc.right).(*pb.Data)
		err := MeldData(leftDst, right)
		assert.NoError(t, err, "[%s] failed to meld", tc.name)

		if diff := cmp.Diff(tc.expected, leftDst, cmpWitnessOptions...); diff != "" {
			t.Errorf("[%s] (right to left) found diff:\n %v", tc.name, diff)
		}

		rightDst := proto.Clone(tc.right).(*pb.Data)
		left := proto.Clone(tc.left).(*pb.Data)
		err = MeldData(rightDst, left)
		assert.NoError(t, err, "[%s] failed to meld", tc.name)

		if diff := cmp.Diff(tc.expected, rightDst, cmpWitnessOptions...); diff != "" {
			t.Errorf("[%s] (left to right) found diff:\n %v", tc.name, diff)
		}
	}
}
