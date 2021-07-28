package pbhash

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akitasoftware/akita-libs/test"
	protohash "github.com/akitasoftware/objecthash-proto"
)

func TestHash(t *testing.T) {
	// Should get the same result if we remove hashmap assumptions.
	slowOptions := []protohash.Option{}
	for _, opt := range options {
		if _, isAssumeHashMap := opt.(protohash.IsAssumeHashMap); !isAssumeHashMap {
			slowOptions = append(slowOptions, opt)
		}
	}

	slowHasher := protohash.NewHasher(slowOptions...)

	spec := test.LoadAPISpecFromFileOrDie("../spec_util/testdata/generalize_witnesses/spec.pb.txt")

	// Sanity-check the test case. The args and responses should be mapped by
	// their hashes.
	for _, m := range spec.GetMethods() {
		for k, v := range m.Args {
			k2, err := hashProto(v, slowHasher)
			assert.Nil(t, err, "Error hashing arg")
			assert.Equal(t, k, k2)
		}
		for k, v := range m.Responses {
			k2, err := hashProto(v, slowHasher)
			assert.Nil(t, err, "Error hashing response")
			assert.Equal(t, k, k2)
		}
	}

	// Check the overall hash of the entire method.
	{
		h1, err := HashProto(spec)
		assert.Nil(t, err, "Error hashing spec")

		h2, err := hashProto(spec, slowHasher)
		assert.Nil(t, err, "Error hashing spec with slow hasher")

		assert.Equal(t, h2, h1)
	}
}
