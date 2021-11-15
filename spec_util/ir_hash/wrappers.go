package ir_hash

import (
	"encoding/base64"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

func HashWitnessToString(w *pb.Witness) string {
	return base64.URLEncoding.EncodeToString(HashWitness(w))
}

func HashDataToString(d *pb.Data) string {
	return base64.URLEncoding.EncodeToString(HashData(d))
}

func HashDataMetaToString(d *pb.DataMeta) string {
	return base64.URLEncoding.EncodeToString(HashDataMeta(d))
}

// TODO: take any proto.Message or interface{} and hash it?
