package spec_util

import (
	"testing"

	"github.com/akitasoftware/akita-libs/test"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
)

func TestInferMaps(t *testing.T) {
	type testData struct {
		name                string
		witnessFile         string
		expectedWitnessFile string
	}

	var tests = []testData{
		{
			"map_1",
			"testdata/meld/infer_map_1.pb.txt",
			"testdata/meld/infer_map_1_expected.pb.txt",
		},
	}

	for _, testData := range tests {
		witness := test.LoadWitnessFromFileOrDile(testData.witnessFile).Method
		expected := test.LoadWitnessFromFileOrDile(testData.expectedWitnessFile).Method

		InferMapsInMethod(witness)

		if diff := cmp.Diff(expected, witness, cmp.Comparer(proto.Equal)); diff != "" {
			t.Errorf("%s\n%v", testData.name, diff)
			continue
		}
	}
}
