package spec_util

import (
	"testing"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/akitasoftware/akita-libs/test"
)

func TestHashKeyRewrite(t *testing.T) {
	// We reuse test data from meld_test, because the expected witnesses
	// cover a range of expected IR objects.
	for _, testData := range tests {
		witness := test.LoadWitnessFromFileOrDile(testData.expectedWitnessFile)
		expected := test.LoadWitnessFromFileOrDile(testData.expectedWitnessFile)

		// Witnesses are just specs with a single method.
		err := RewriteHashKeys(&pb.APISpec{
			Methods: []*pb.Method{witness.Method},
		})
		assert.NoError(t, err, testData.name)
		assert.Equal(t, proto.MarshalTextString(expected), proto.MarshalTextString(witness), testData.name)
	}
}
