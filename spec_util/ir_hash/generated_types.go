package ir_hash

import (
	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/OneOfOne/xxhash"
)

func HashInt32Value(node *wrappers.Int32Value) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Value != 0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Int32(node.Value))
	}
	return hash.Sum(nil)
}
func HashInt64Value(node *wrappers.Int64Value) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Value != 0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Int64(node.Value))
	}
	return hash.Sum(nil)
}
func HashUInt32Value(node *wrappers.UInt32Value) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Value != 0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Uint32(node.Value))
	}
	return hash.Sum(nil)
}
func HashUInt64Value(node *wrappers.UInt64Value) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Value != 0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Uint64(node.Value))
	}
	return hash.Sum(nil)
}
func HashFloatValue(node *wrappers.FloatValue) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Value != 0.0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Float32(node.Value))
	}
	return hash.Sum(nil)
}
func HashDoubleValue(node *wrappers.DoubleValue) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Value != 0.0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Float64(node.Value))
	}
	return hash.Sum(nil)
}
func HashAkitaAnnotations(node *pb.AkitaAnnotations) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.IsFree != false {
		hash.Write(intHashes[1])
		hash.Write(Hash_Bool(node.IsFree))
	}
	if node.FormatOption != nil {
		hash.Write(intHashes[3])
		hash.Write(HashFormatOption(node.FormatOption))
	}
	if node.IsSensitive != false {
		hash.Write(intHashes[2])
		hash.Write(Hash_Bool(node.IsSensitive))
	}
	return hash.Sum(nil)
}
func HashBool(node *pb.Bool) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Type != nil {
		hash.Write(intHashes[1])
		hash.Write(HashBoolType(node.Type))
	}
	if node.Value != false {
		hash.Write(intHashes[2])
		hash.Write(Hash_Bool(node.Value))
	}
	return hash.Sum(nil)
}
func HashBoolType(node *pb.BoolType) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.FixedValues) != 0 {
		hash.Write(intHashes[1])
		listHash := xxhash.New64()
		listHash.Write([]byte("l"))
		for _, v := range node.FixedValues {
			listHash.Write(Hash_Bool(v))
		}
		hash.Write(listHash.Sum(nil))
	}
	return hash.Sum(nil)
}
func HashBytes(node *pb.Bytes) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Type != nil {
		hash.Write(intHashes[1])
		hash.Write(HashBytesType(node.Type))
	}
	if node.Value != nil {
		hash.Write(intHashes[2])
		hash.Write(Hash_Bytes(node.Value))
	}
	return hash.Sum(nil)
}
func HashBytesType(node *pb.BytesType) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.FixedValues) != 0 {
		hash.Write(intHashes[1])
		listHash := xxhash.New64()
		listHash.Write([]byte("l"))
		for _, v := range node.FixedValues {
			listHash.Write(Hash_Bytes(v))
		}
		hash.Write(listHash.Sum(nil))
	}
	return hash.Sum(nil)
}
func HashDouble(node *pb.Double) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Type != nil {
		hash.Write(intHashes[1])
		hash.Write(HashDoubleType(node.Type))
	}
	if node.Value != 0.0 {
		hash.Write(intHashes[2])
		hash.Write(Hash_Float64(node.Value))
	}
	return hash.Sum(nil)
}
func HashDoubleType(node *pb.DoubleType) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.FixedValues) != 0 {
		hash.Write(intHashes[1])
		listHash := xxhash.New64()
		listHash.Write([]byte("l"))
		for _, v := range node.FixedValues {
			listHash.Write(Hash_Float64(v))
		}
		hash.Write(listHash.Sum(nil))
	}
	if node.Max != nil {
		hash.Write(intHashes[3])
		hash.Write(HashDoubleValue(node.Max))
	}
	if node.Min != nil {
		hash.Write(intHashes[2])
		hash.Write(HashDoubleValue(node.Min))
	}
	return hash.Sum(nil)
}
func HashInt32(node *pb.Int32) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Type != nil {
		hash.Write(intHashes[1])
		hash.Write(HashInt32Type(node.Type))
	}
	if node.Value != 0 {
		hash.Write(intHashes[2])
		hash.Write(Hash_Int32(node.Value))
	}
	return hash.Sum(nil)
}
func HashInt32Type(node *pb.Int32Type) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.FixedValues) != 0 {
		hash.Write(intHashes[1])
		listHash := xxhash.New64()
		listHash.Write([]byte("l"))
		for _, v := range node.FixedValues {
			listHash.Write(Hash_Int32(v))
		}
		hash.Write(listHash.Sum(nil))
	}
	if node.Max != nil {
		hash.Write(intHashes[3])
		hash.Write(HashInt32Value(node.Max))
	}
	if node.Min != nil {
		hash.Write(intHashes[2])
		hash.Write(HashInt32Value(node.Min))
	}
	return hash.Sum(nil)
}
func HashInt64(node *pb.Int64) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Type != nil {
		hash.Write(intHashes[1])
		hash.Write(HashInt64Type(node.Type))
	}
	if node.Value != 0 {
		hash.Write(intHashes[2])
		hash.Write(Hash_Int64(node.Value))
	}
	return hash.Sum(nil)
}
func HashInt64Type(node *pb.Int64Type) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.FixedValues) != 0 {
		hash.Write(intHashes[1])
		listHash := xxhash.New64()
		listHash.Write([]byte("l"))
		for _, v := range node.FixedValues {
			listHash.Write(Hash_Int64(v))
		}
		hash.Write(listHash.Sum(nil))
	}
	if node.Max != nil {
		hash.Write(intHashes[3])
		hash.Write(HashInt64Value(node.Max))
	}
	if node.Min != nil {
		hash.Write(intHashes[2])
		hash.Write(HashInt64Value(node.Min))
	}
	return hash.Sum(nil)
}
func HashFloat(node *pb.Float) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Type != nil {
		hash.Write(intHashes[1])
		hash.Write(HashFloatType(node.Type))
	}
	if node.Value != 0.0 {
		hash.Write(intHashes[2])
		hash.Write(Hash_Float32(node.Value))
	}
	return hash.Sum(nil)
}
func HashFloatType(node *pb.FloatType) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.FixedValues) != 0 {
		hash.Write(intHashes[1])
		listHash := xxhash.New64()
		listHash.Write([]byte("l"))
		for _, v := range node.FixedValues {
			listHash.Write(Hash_Float32(v))
		}
		hash.Write(listHash.Sum(nil))
	}
	if node.Max != nil {
		hash.Write(intHashes[3])
		hash.Write(HashFloatValue(node.Max))
	}
	if node.Min != nil {
		hash.Write(intHashes[2])
		hash.Write(HashFloatValue(node.Min))
	}
	return hash.Sum(nil)
}
func HashFormatOption(node *pb.FormatOption) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if val, ok := node.Format.(*pb.FormatOption_StringFormat); ok {
		hash.Write(intHashes[1])
		hash.Write(Hash_Unicode(val.StringFormat))
	}
	return hash.Sum(nil)
}
func HashPrimitive(node *pb.Primitive) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if val, ok := node.Value.(*pb.Primitive_Uint32Value); ok {
		hash.Write(intHashes[7])
		hash.Write(HashUint32(val.Uint32Value))
	}
	if node.TypeHint != "" {
		hash.Write(intHashes[1])
		hash.Write(Hash_Unicode(node.TypeHint))
	}
	if len(node.Formats) != 0 {
		hash.Write(intHashes[13])
		pairs := make ([]KeyValuePair, 0, len(node.Formats))
		for k, v := range node.Formats {
			pairs = append(pairs, KeyValuePair{Hash_Unicode(k), Hash_Bool(v)})
		}
		hash.Write(Hash_KeyValues(pairs))
	}
	if val, ok := node.Value.(*pb.Primitive_Int64Value); ok {
		hash.Write(intHashes[6])
		hash.Write(HashInt64(val.Int64Value))
	}
	if val, ok := node.Value.(*pb.Primitive_StringValue); ok {
		hash.Write(intHashes[4])
		hash.Write(HashString(val.StringValue))
	}
	if node.FormatKind != "" {
		hash.Write(intHashes[14])
		hash.Write(Hash_Unicode(node.FormatKind))
	}
	if val, ok := node.Value.(*pb.Primitive_DoubleValue); ok {
		hash.Write(intHashes[9])
		hash.Write(HashDouble(val.DoubleValue))
	}
	if val, ok := node.Value.(*pb.Primitive_BytesValue); ok {
		hash.Write(intHashes[3])
		hash.Write(HashBytes(val.BytesValue))
	}
	if node.AkitaAnnotations != nil {
		hash.Write(intHashes[11])
		hash.Write(HashAkitaAnnotations(node.AkitaAnnotations))
	}
	if val, ok := node.Value.(*pb.Primitive_FloatValue); ok {
		hash.Write(intHashes[10])
		hash.Write(HashFloat(val.FloatValue))
	}
	if node.ContainsRandomValue != false {
		hash.Write(intHashes[12])
		hash.Write(Hash_Bool(node.ContainsRandomValue))
	}
	if val, ok := node.Value.(*pb.Primitive_Uint64Value); ok {
		hash.Write(intHashes[8])
		hash.Write(HashUint64(val.Uint64Value))
	}
	if val, ok := node.Value.(*pb.Primitive_Int32Value); ok {
		hash.Write(intHashes[5])
		hash.Write(HashInt32(val.Int32Value))
	}
	if val, ok := node.Value.(*pb.Primitive_BoolValue); ok {
		hash.Write(intHashes[2])
		hash.Write(HashBool(val.BoolValue))
	}
	return hash.Sum(nil)
}
func HashStringType(node *pb.StringType) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.FixedValues) != 0 {
		hash.Write(intHashes[1])
		listHash := xxhash.New64()
		listHash.Write([]byte("l"))
		for _, v := range node.FixedValues {
			listHash.Write(Hash_Unicode(v))
		}
		hash.Write(listHash.Sum(nil))
	}
	if node.Regex != "" {
		hash.Write(intHashes[2])
		hash.Write(Hash_Unicode(node.Regex))
	}
	return hash.Sum(nil)
}
func HashString(node *pb.String) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Type != nil {
		hash.Write(intHashes[1])
		hash.Write(HashStringType(node.Type))
	}
	if node.Value != "" {
		hash.Write(intHashes[2])
		hash.Write(Hash_Unicode(node.Value))
	}
	return hash.Sum(nil)
}
func HashUint32(node *pb.Uint32) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Type != nil {
		hash.Write(intHashes[1])
		hash.Write(HashUint32Type(node.Type))
	}
	if node.Value != 0 {
		hash.Write(intHashes[2])
		hash.Write(Hash_Uint32(node.Value))
	}
	return hash.Sum(nil)
}
func HashUint32Type(node *pb.Uint32Type) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.FixedValues) != 0 {
		hash.Write(intHashes[1])
		listHash := xxhash.New64()
		listHash.Write([]byte("l"))
		for _, v := range node.FixedValues {
			listHash.Write(Hash_Uint32(v))
		}
		hash.Write(listHash.Sum(nil))
	}
	if node.Max != nil {
		hash.Write(intHashes[3])
		hash.Write(HashUInt32Value(node.Max))
	}
	if node.Min != nil {
		hash.Write(intHashes[2])
		hash.Write(HashUInt32Value(node.Min))
	}
	return hash.Sum(nil)
}
func HashUint64(node *pb.Uint64) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Type != nil {
		hash.Write(intHashes[1])
		hash.Write(HashUint64Type(node.Type))
	}
	if node.Value != 0 {
		hash.Write(intHashes[2])
		hash.Write(Hash_Uint64(node.Value))
	}
	return hash.Sum(nil)
}
func HashUint64Type(node *pb.Uint64Type) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.FixedValues) != 0 {
		hash.Write(intHashes[1])
		listHash := xxhash.New64()
		listHash.Write([]byte("l"))
		for _, v := range node.FixedValues {
			listHash.Write(Hash_Uint64(v))
		}
		hash.Write(listHash.Sum(nil))
	}
	if node.Max != nil {
		hash.Write(intHashes[3])
		hash.Write(HashUInt64Value(node.Max))
	}
	if node.Min != nil {
		hash.Write(intHashes[2])
		hash.Write(HashUInt64Value(node.Min))
	}
	return hash.Sum(nil)
}
