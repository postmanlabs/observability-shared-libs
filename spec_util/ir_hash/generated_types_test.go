package ir_hash

import (
	"encoding/base64"
	"testing"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/pbhash"
	"github.com/akitasoftware/akita-libs/test"
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
			"unsigned integer",
			&pb.Primitive{
				Value: &pb.Primitive_Int64Value{
					Int64Value: &pb.Int64{
						Value: 5,
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

func TestEnum(t *testing.T) {
	p := &pb.HTTPBody{
		ContentType: pb.HTTPBody_JSON,
	}
	hash1, err := pbhash.HashProto(p)
	if err != nil {
		t.Fatal(err)
	}

	hash2 := base64.URLEncoding.EncodeToString(HashHTTPBody(p))

	if hash1 != hash2 {
		t.Errorf("Hashes are unequal, %v != %v", hash1, hash2)
	}
}

func TestTwoFields(t *testing.T) {
	h := &pb.HTTPMeta{
		Location: &pb.HTTPMeta_Body{
			Body: &pb.HTTPBody{
				ContentType: pb.HTTPBody_JSON,
			},
		},
		ResponseCode: 200,
	}

	hash1, err := pbhash.HashProto(h)
	if err != nil {
		t.Fatal(err)
	}
	hash2 := base64.URLEncoding.EncodeToString(HashHTTPMeta(h))

	if hash1 != hash2 {
		t.Errorf("Hashes are unequal, %v != %v", hash1, hash2)
	}

}

func makeInt(x int64) *pb.Data {
	return &pb.Data{
		Value: &pb.Data_Primitive{
			Primitive: &pb.Primitive{
				Value: &pb.Primitive_Int64Value{
					Int64Value: &pb.Int64{
						Value: x,
					},
				},
			},
		},
	}
}

func makeString(x string) *pb.Data {
	return &pb.Data{
		Value: &pb.Data_Primitive{
			Primitive: &pb.Primitive{
				Value: &pb.Primitive_StringValue{
					StringValue: &pb.String{
						Value: x,
					},
				},
			},
		},
	}
}

func TestRecursiveData(t *testing.T) {
	p := &pb.Data{
		Value: &pb.Data_Struct{
			Struct: &pb.Struct{
				Fields: map[string]*pb.Data{
					"key1": &pb.Data{
						Value: &pb.Data_Struct{
							Struct: &pb.Struct{
								Fields: map[string]*pb.Data{
									"aaa": makeInt(23),
									"bbb": makeInt(45),
								},
							},
						},
					},
					"key2": &pb.Data{
						Value: &pb.Data_Struct{
							Struct: &pb.Struct{
								Fields: map[string]*pb.Data{
									"ccc": makeInt(56),
								},
							},
						},
					},
				},
			},
		},
	}

	hash1, err := pbhash.HashProto(p)
	if err != nil {
		t.Fatal(err)
	}

	hash2 := base64.URLEncoding.EncodeToString(HashData(p))

	if hash1 != hash2 {
		t.Errorf("Hashes are unequal, %v != %v", hash1, hash2)
	}
}

func TestData(t *testing.T) {
	testCases := []struct {
		Name string
		D    *pb.Data
	}{
		{
			"single-entry struct",
			&pb.Data{
				Value: &pb.Data_Struct{
					Struct: &pb.Struct{
						Fields: map[string]*pb.Data{
							"name": makeString("file_1"),
						},
					},
				},
			},
		},
		{
			"single-entry struct with metadata",
			&pb.Data{
				Value: &pb.Data_Struct{
					Struct: &pb.Struct{
						Fields: map[string]*pb.Data{
							"name": makeString("file_1"),
						},
					},
				},
				Meta: &pb.DataMeta{
					Meta: &pb.DataMeta_Http{
						Http: &pb.HTTPMeta{
							Location: &pb.HTTPMeta_Body{
								Body: &pb.HTTPBody{
									ContentType: pb.HTTPBody_JSON,
								},
							},
							ResponseCode: 200,
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Log(tc.Name)
		hash1, err := pbhash.HashProto(tc.D)
		if err != nil {
			t.Fatal(err)
		}

		hash2 := base64.URLEncoding.EncodeToString(HashData(tc.D))

		if hash1 != hash2 {
			t.Errorf("Hashes are unequal, %v != %v", hash1, hash2)
		}
	}
}

func TestIgnoredField(t *testing.T) {
	m1 := &pb.HTTPMethodMeta{
		Method:            "POST",
		PathTemplate:      "/v1/foo",
		Host:              "localhost:8000",
		ProcessingLatency: 32.0,
	}
	m2 := &pb.HTTPMethodMeta{
		Method:            "POST",
		PathTemplate:      "/v1/foo",
		Host:              "localhost:8000",
		ProcessingLatency: 0.0,
	}
	m3 := &pb.HTTPMethodMeta{
		Method:            "POST",
		PathTemplate:      "/v1/foo",
		Host:              "localhost:8000",
		ProcessingLatency: 1.1,
	}

	h1a, _ := pbhash.HashProto(m1)
	h2a, _ := pbhash.HashProto(m2)
	h3a, _ := pbhash.HashProto(m3)

	h1b := base64.URLEncoding.EncodeToString(HashHTTPMethodMeta(m1))
	h2b := base64.URLEncoding.EncodeToString(HashHTTPMethodMeta(m2))
	h3b := base64.URLEncoding.EncodeToString(HashHTTPMethodMeta(m3))

	if h1a != h2a || h2a != h3a {
		t.Fatalf("Hashes are not identical using pbhash: %v %v %v", h1a, h2a, h3a)
	}

	if h1b != h2b || h2b != h3b {
		t.Errorf("Hashes are not identical using gen hash: %v %v %v", h1b, h2b, h3b)
	}

	if h1a != h2a {
		t.Errorf("Hashes are not equal: %v %v %v != %v %v %v", h1a, h2a, h3a, h1b, h2b, h3b)
	}

}

func TestWitnesses(t *testing.T) {
	witnessFiles := []string{
		"../testdata/meld/meld_no_data_formats.pb.txt",
		"../testdata/meld/meld_examples_1.pb.txt",
		"../testdata/meld/meld_examples_2.pb.txt",
		"../testdata/meld/meld_examples_3.pb.txt",
		// "../testdata/meld/meld_oneof_with_primitive_1.pb.txt",  // bad key
		"../testdata/generalize_witnesses/gitlab.1.concrete.pb.txt",
		"../testdata/generalize_witnesses/gitlab.1.generalized.pb.txt",
		"../testdata/generalize_witnesses/gitlab.2.concrete.pb.txt",
		"../testdata/generalize_witnesses/gitlab.2.generalized.pb.txt",
		"../testdata/generalize_witnesses/gitlab.3.concrete.pb.txt",
		"../testdata/generalize_witnesses/gitlab.3.generalized.pb.txt",
		"../testdata/generalize_witnesses/gitlab.4.concrete.pb.txt",
		"../testdata/generalize_witnesses/gitlab.4.generalized.pb.txt",
		"../testdata/generalize_witnesses/gitlab.5.concrete.pb.txt",
		"../testdata/generalize_witnesses/gitlab.5.generalized.pb.txt",
		"../testdata/generalize_witnesses/gitlab.6.concrete.pb.txt",
		"../testdata/generalize_witnesses/gitlab.6.generalized.pb.txt",
	}

	for _, wFile := range witnessFiles {
		t.Logf("Loading %v", wFile)
		witness := test.LoadWitnessFromFileOrDile(wFile)

		hash1, err := pbhash.HashProto(witness)
		if err != nil {
			t.Fatal(err)
		}

		hash2 := base64.URLEncoding.EncodeToString(HashWitness(witness))
		t.Logf("Hash: %v", hash1)

		if hash1 != hash2 {
			t.Errorf("Witness hashes are unequal, %v != %v", hash1, hash2)

			hash3, err := pbhash.HashProto(witness.Method)
			if err != nil {
				t.Fatal(err)
			}

			hash4 := base64.URLEncoding.EncodeToString(HashMethod(witness.Method))

			if hash3 != hash4 {
				t.Errorf("Method hashes are unequal, %v != %v", hash3, hash4)
			}

			for k, v := range witness.Method.Args {
				expected, _ := pbhash.HashProto(v)
				actual := base64.URLEncoding.EncodeToString(HashData(v))
				t.Logf("arg %v expected %v actual %v", k, expected, actual)
			}

			for k, v := range witness.Method.Responses {
				expected, _ := pbhash.HashProto(v)
				actual := base64.URLEncoding.EncodeToString(HashData(v))
				t.Logf("response %v expected %v actual %v", k, expected, actual)
			}
		}
	}
}

func TestSpec(t *testing.T) {
	spec := test.LoadAPISpecFromFileOrDie("../testdata/generalize_witnesses/spec.pb.txt")

	hash1, err := pbhash.HashProto(spec)
	if err != nil {
		t.Fatal(err)
	}

	hash2 := base64.URLEncoding.EncodeToString(HashAPISpec(spec))

	if hash1 != hash2 {
		t.Errorf("Spec hashes are unequal, %v != %v", hash1, hash2)
	}
}

func BenchmarkWitnessHash(b *testing.B) {
	witness := test.LoadWitnessFromFileOrDile("../testdata/generalize_witnesses/gitlab.1.concrete.pb.txt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = base64.URLEncoding.EncodeToString(HashWitness(witness))
	}
}

func BenchmarkWitnessOldHash(b *testing.B) {
	witness := test.LoadWitnessFromFileOrDile("../testdata/generalize_witnesses/gitlab.1.concrete.pb.txt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pbhash.HashProto(witness)
	}
}
