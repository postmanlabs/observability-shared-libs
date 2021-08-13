package ir_hash

import (
	"bytes"
	"fmt"

	"github.com/OneOfOne/xxhash"
	"github.com/golang/protobuf/proto"
)

// Verify that all the files used to generate the hash functions have not changed.
// If you are looking at these panics, it's because gen.go has not been run after the last time the
// protobuf generator was run on akita-ir, so the stored file descriptor mismatches.
//
// IMPORTANT: update the list of objects in gen.go with any new messages.  Then run "make"
// in this directory, and commit the updated generated_types.go.
func init() {
	for fn, expected := range ProtobufFileHashes {
		fdgzip := proto.FileDescriptor(fn)
		if fdgzip == nil {
			panic(fmt.Sprintf("Protobuf file descriptor not found for %q. Rerun gen.go after updating IR.", fn))
		}
		h := xxhash.New64()
		h.Write([]byte(fdgzip))
		actual := h.Sum(nil)

		if bytes.Compare(expected, actual) != 0 {
			panic(fmt.Sprintf("Protobuf file description mismatch for %q, expected %v got %v. Rerurn gen.go after updating IR.",
				fn, expected, actual))
		}
	}
}
