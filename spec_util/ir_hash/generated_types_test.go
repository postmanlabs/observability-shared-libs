package ir_hash

import (
	"encoding/base64"
	"testing"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/pbhash"
)

func TestPrimitives(t *testing.T) {
	testCases := []struct {
		Name string
		P    *pb.Primitive
	}{
		{
			"bool primitive with type",
			&pb.Primitive{
				Value: &pb.Primitive_BoolValue{
					BoolValue: &pb.Bool{
						Type: &pb.BoolType{
							FixedValues: []bool{true, false},
						},
						Value: true,
					},
				},
			},
		},
		{
			"bool primitive without type",
			&pb.Primitive{
				Value: &pb.Primitive_BoolValue{
					BoolValue: &pb.Bool{
						Value: true,
					},
				},
			},
		},
		{
			"bytes primitive without type",
			&pb.Primitive{
				Value: &pb.Primitive_BytesValue{
					BytesValue: &pb.Bytes{
						Value: []byte{1, 2, 3, 4},
					},
				},
			},
		},
		/* errors with pbhash???
		{
			"bytes primitive with type",
			&pb.Primitive{
				Value: &pb.Primitive_BytesValue{
					BytesValue: &pb.Bytes{
						Type: &pb.BytesType{
							FixedValues: [][]byte{
								{1, 2, 3, 4},
								{0, 0, 0, 0},
							},
						},
						Value: []byte{1, 2, 3, 4},
					},
				},
			},
		},
		*/
		{
			"string primitive without type",
			&pb.Primitive{
				Value: &pb.Primitive_StringValue{
					StringValue: &pb.String{
						Value: "abcdef",
					},
				},
			},
		},
		{
			"string primitive, empty string",
			&pb.Primitive{
				Value: &pb.Primitive_StringValue{
					StringValue: &pb.String{
						Value: "",
					},
				},
			},
		},
		{
			"signed integer without type",
			&pb.Primitive{
				Value: &pb.Primitive_Int32Value{
					Int32Value: &pb.Int32{
						Value: -3,
					},
				},
			},
		},
		{
			"float primitive without type",
			&pb.Primitive{
				Value: &pb.Primitive_FloatValue{
					FloatValue: &pb.Float{
						Value: -3.14,
					},
				},
			},
		},
		{
			"double primitive without type",
			&pb.Primitive{
				Value: &pb.Primitive_DoubleValue{
					DoubleValue: &pb.Double{
						Value: 1.1e-100,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Log(tc.Name)
		hash1, err := pbhash.HashProto(tc.P)
		if err != nil {
			t.Fatal(err)
		}

		hash2 := base64.URLEncoding.EncodeToString(HashPrimitive(tc.P))

		if hash1 != hash2 {
			t.Errorf("Hashes are unequal, %v != %v", hash1, hash2)
		}
	}
}
