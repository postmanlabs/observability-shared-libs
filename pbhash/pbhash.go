package pbhash

import (
	"encoding/base64"

	protohash "github.com/akitasoftware/objecthash-proto"
	"github.com/golang/protobuf/proto"
)

var (
	ph = protohash.NewHasher(
		protohash.BasicHashFunction(protohash.XXHASH64),

		// Example values can fluctuate between runs, so we ignore them.
		protohash.IgnoreFieldName("example_values"),

		// Do not include latency when deduplicating by hash
		// for a learning sessions
		protohash.IgnoreFieldName("processing_latency"),

		// XXX: MethodId.name currently holds the OpenAPI3 operationId, which
		// is optional and only used for documentation.  In the future,
		// we should move that to metadata and either use a meaningful method name
		// or remove the field.
		protohash.IgnoreFieldName("name"),
	)
)

// Hashes the proto using objecthash-proto with XXHash64 as the basic function.
// Returns the hash in base64 URL encoded form so it can be used as strings in
// protobuf messages (which require UTF-8 encoding).
func HashProto(m proto.Message) (string, error) {
	h, err := ph.HashProto(m)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(h), nil
}
