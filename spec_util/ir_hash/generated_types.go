package ir_hash

import (
	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	"github.com/OneOfOne/xxhash"
)

func HashInt32Value(node *wrapperspb.Int32Value) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Value != 0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Int32(node.Value))
	}
	return hash.Sum(nil)
}
func HashInt64Value(node *wrapperspb.Int64Value) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Value != 0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Int64(node.Value))
	}
	return hash.Sum(nil)
}
func HashUInt32Value(node *wrapperspb.UInt32Value) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Value != 0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Uint32(node.Value))
	}
	return hash.Sum(nil)
}
func HashUInt64Value(node *wrapperspb.UInt64Value) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Value != 0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Uint64(node.Value))
	}
	return hash.Sum(nil)
}
func HashFloatValue(node *wrapperspb.FloatValue) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Value != 0.0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Float32(node.Value))
	}
	return hash.Sum(nil)
}
func HashDoubleValue(node *wrapperspb.DoubleValue) []byte {
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
	if node.IsSensitive != false {
		hash.Write(intHashes[2])
		hash.Write(Hash_Bool(node.IsSensitive))
	}
	if node.FormatOption != nil {
		hash.Write(intHashes[3])
		hash.Write(HashFormatOption(node.FormatOption))
	}
	return hash.Sum(nil)
}
func HashAPISpec(node *pb.APISpec) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.Methods) != 0 {
		hash.Write(intHashes[1])
		listHash := xxhash.New64()
		listHash.Write([]byte("l"))
		for _, v := range node.Methods {
			listHash.Write(HashMethod(v))
		}
		hash.Write(listHash.Sum(nil))
	}
	if len(node.Tags) != 0 {
		hash.Write(intHashes[2])
		pairs := make ([]KeyValuePair, 0, len(node.Tags))
		for k, v := range node.Tags {
			pairs = append(pairs, KeyValuePair{Hash_Unicode(k), Hash_Unicode(v)})
		}
		hash.Write(Hash_KeyValues(pairs))
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
func HashData(node *pb.Data) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if val, ok := node.Value.(*pb.Data_Oneof); ok {
		hash.Write(intHashes[6])
		hash.Write(HashOneOf(val.Oneof))
	}
	if val, ok := node.Value.(*pb.Data_Optional); ok {
		hash.Write(intHashes[4])
		hash.Write(HashOptional(val.Optional))
	}
	if val, ok := node.Value.(*pb.Data_Primitive); ok {
		hash.Write(intHashes[1])
		hash.Write(HashPrimitive(val.Primitive))
	}
	if node.Meta != nil {
		hash.Write(intHashes[5])
		hash.Write(HashDataMeta(node.Meta))
	}
	if val, ok := node.Value.(*pb.Data_Struct); ok {
		hash.Write(intHashes[2])
		hash.Write(HashStruct(val.Struct))
	}
	if val, ok := node.Value.(*pb.Data_List); ok {
		hash.Write(intHashes[3])
		hash.Write(HashList(val.List))
	}
	if node.Nullable != false {
		hash.Write(intHashes[7])
		hash.Write(Hash_Bool(node.Nullable))
	}
	return hash.Sum(nil)
}
func HashDataMeta(node *pb.DataMeta) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if val, ok := node.Meta.(*pb.DataMeta_Grpc); ok {
		hash.Write(intHashes[1])
		hash.Write(HashGRPCMeta(val.Grpc))
	}
	if val, ok := node.Meta.(*pb.DataMeta_Http); ok {
		hash.Write(intHashes[2])
		hash.Write(HashHTTPMeta(val.Http))
	}
	return hash.Sum(nil)
}
func HashDataRef(node *pb.DataRef) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if val, ok := node.ValueRef.(*pb.DataRef_PrimitiveRef); ok {
		hash.Write(intHashes[1])
		hash.Write(HashPrimitiveRef(val.PrimitiveRef))
	}
	if val, ok := node.ValueRef.(*pb.DataRef_StructRef); ok {
		hash.Write(intHashes[2])
		hash.Write(HashStructRef(val.StructRef))
	}
	if val, ok := node.ValueRef.(*pb.DataRef_ListRef); ok {
		hash.Write(intHashes[3])
		hash.Write(HashListRef(val.ListRef))
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
	if node.Min != nil {
		hash.Write(intHashes[2])
		hash.Write(HashDoubleValue(node.Min))
	}
	if node.Max != nil {
		hash.Write(intHashes[3])
		hash.Write(HashDoubleValue(node.Max))
	}
	return hash.Sum(nil)
}
func HashExampleValue(node *pb.ExampleValue) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	return hash.Sum(nil)
}
func HashGRPCMeta(node *pb.GRPCMeta) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	return hash.Sum(nil)
}
func HashGRPCMethodMeta(node *pb.GRPCMethodMeta) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	return hash.Sum(nil)
}
func HashHTTPAuth(node *pb.HTTPAuth) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Type != 0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Int32(int32(node.Type)))
	}
	return hash.Sum(nil)
}
func HashHTTPBody(node *pb.HTTPBody) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.ContentType != 0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Int32(int32(node.ContentType)))
	}
	return hash.Sum(nil)
}
func HashHTTPCookie(node *pb.HTTPCookie) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Key != "" {
		hash.Write(intHashes[1])
		hash.Write(Hash_Unicode(node.Key))
	}
	return hash.Sum(nil)
}
func HashHTTPEmpty(node *pb.HTTPEmpty) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	return hash.Sum(nil)
}
func HashHTTPHeader(node *pb.HTTPHeader) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Key != "" {
		hash.Write(intHashes[1])
		hash.Write(Hash_Unicode(node.Key))
	}
	return hash.Sum(nil)
}
func HashHTTPMeta(node *pb.HTTPMeta) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if val, ok := node.Location.(*pb.HTTPMeta_Empty); ok {
		hash.Write(intHashes[6])
		hash.Write(HashHTTPEmpty(val.Empty))
	}
	if val, ok := node.Location.(*pb.HTTPMeta_Cookie); ok {
		hash.Write(intHashes[4])
		hash.Write(HashHTTPCookie(val.Cookie))
	}
	if val, ok := node.Location.(*pb.HTTPMeta_Path); ok {
		hash.Write(intHashes[1])
		hash.Write(HashHTTPPath(val.Path))
	}
	if val, ok := node.Location.(*pb.HTTPMeta_Body); ok {
		hash.Write(intHashes[5])
		hash.Write(HashHTTPBody(val.Body))
	}
	if val, ok := node.Location.(*pb.HTTPMeta_Query); ok {
		hash.Write(intHashes[2])
		hash.Write(HashHTTPQuery(val.Query))
	}
	if val, ok := node.Location.(*pb.HTTPMeta_Header); ok {
		hash.Write(intHashes[3])
		hash.Write(HashHTTPHeader(val.Header))
	}
	if val, ok := node.Location.(*pb.HTTPMeta_Auth); ok {
		hash.Write(intHashes[8])
		hash.Write(HashHTTPAuth(val.Auth))
	}
	if node.ResponseCode != 0 {
		hash.Write(intHashes[7])
		hash.Write(Hash_Int32(node.ResponseCode))
	}
	if val, ok := node.Location.(*pb.HTTPMeta_Multipart); ok {
		hash.Write(intHashes[9])
		hash.Write(HashHTTPMultipart(val.Multipart))
	}
	return hash.Sum(nil)
}
func HashHTTPMethodMeta(node *pb.HTTPMethodMeta) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Method != "" {
		hash.Write(intHashes[1])
		hash.Write(Hash_Unicode(node.Method))
	}
	if node.PathTemplate != "" {
		hash.Write(intHashes[2])
		hash.Write(Hash_Unicode(node.PathTemplate))
	}
	if node.Host != "" {
		hash.Write(intHashes[3])
		hash.Write(Hash_Unicode(node.Host))
	}
	return hash.Sum(nil)
}
func HashHTTPMultipart(node *pb.HTTPMultipart) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Type != "" {
		hash.Write(intHashes[1])
		hash.Write(Hash_Unicode(node.Type))
	}
	return hash.Sum(nil)
}
func HashHTTPPath(node *pb.HTTPPath) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Key != "" {
		hash.Write(intHashes[1])
		hash.Write(Hash_Unicode(node.Key))
	}
	return hash.Sum(nil)
}
func HashHTTPQuery(node *pb.HTTPQuery) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Key != "" {
		hash.Write(intHashes[1])
		hash.Write(Hash_Unicode(node.Key))
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
func HashIndexedDataRef(node *pb.IndexedDataRef) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Index != 0 {
		hash.Write(intHashes[1])
		hash.Write(Hash_Int32(node.Index))
	}
	if node.DataRef != nil {
		hash.Write(intHashes[2])
		hash.Write(HashDataRef(node.DataRef))
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
	if node.Min != nil {
		hash.Write(intHashes[2])
		hash.Write(HashInt32Value(node.Min))
	}
	if node.Max != nil {
		hash.Write(intHashes[3])
		hash.Write(HashInt32Value(node.Max))
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
	if node.Min != nil {
		hash.Write(intHashes[2])
		hash.Write(HashInt64Value(node.Min))
	}
	if node.Max != nil {
		hash.Write(intHashes[3])
		hash.Write(HashInt64Value(node.Max))
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
	if node.Min != nil {
		hash.Write(intHashes[2])
		hash.Write(HashFloatValue(node.Min))
	}
	if node.Max != nil {
		hash.Write(intHashes[3])
		hash.Write(HashFloatValue(node.Max))
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
func HashList(node *pb.List) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.Elems) != 0 {
		hash.Write(intHashes[1])
		listHash := xxhash.New64()
		listHash.Write([]byte("l"))
		for _, v := range node.Elems {
			listHash.Write(HashData(v))
		}
		hash.Write(listHash.Sum(nil))
	}
	return hash.Sum(nil)
}
func HashListRef(node *pb.ListRef) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if val, ok := node.Ref.(*pb.ListRef_FullList); ok {
		hash.Write(intHashes[1])
		hash.Write(HashListRef_FullListRef(val.FullList))
	}
	if val, ok := node.Ref.(*pb.ListRef_ElemRef); ok {
		hash.Write(intHashes[2])
		hash.Write(HashIndexedDataRef(val.ElemRef))
	}
	return hash.Sum(nil)
}
func HashListRef_FullListRef(node *pb.ListRef_FullListRef) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	return hash.Sum(nil)
}
func HashMapData(node *pb.MapData) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Key != nil {
		hash.Write(intHashes[1])
		hash.Write(HashData(node.Key))
	}
	if node.Value != nil {
		hash.Write(intHashes[2])
		hash.Write(HashData(node.Value))
	}
	return hash.Sum(nil)
}
func HashMethod(node *pb.Method) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Meta != nil {
		hash.Write(intHashes[4])
		hash.Write(HashMethodMeta(node.Meta))
	}
	if node.Id != nil {
		hash.Write(intHashes[1])
		hash.Write(HashMethodID(node.Id))
	}
	if len(node.Args) != 0 {
		hash.Write(intHashes[2])
		pairs := make ([]KeyValuePair, 0, len(node.Args))
		for k, v := range node.Args {
			pairs = append(pairs, KeyValuePair{Hash_Unicode(k), HashData(v)})
		}
		hash.Write(Hash_KeyValues(pairs))
	}
	if len(node.Responses) != 0 {
		hash.Write(intHashes[3])
		pairs := make ([]KeyValuePair, 0, len(node.Responses))
		for k, v := range node.Responses {
			pairs = append(pairs, KeyValuePair{Hash_Unicode(k), HashData(v)})
		}
		hash.Write(Hash_KeyValues(pairs))
	}
	return hash.Sum(nil)
}
func HashMethodID(node *pb.MethodID) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Name != "" {
		hash.Write(intHashes[1])
		hash.Write(Hash_Unicode(node.Name))
	}
	if node.ApiType != 0 {
		hash.Write(intHashes[2])
		hash.Write(Hash_Int32(int32(node.ApiType)))
	}
	return hash.Sum(nil)
}
func HashMethodMeta(node *pb.MethodMeta) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if val, ok := node.Meta.(*pb.MethodMeta_Grpc); ok {
		hash.Write(intHashes[1])
		hash.Write(HashGRPCMethodMeta(val.Grpc))
	}
	if val, ok := node.Meta.(*pb.MethodMeta_Http); ok {
		hash.Write(intHashes[2])
		hash.Write(HashHTTPMethodMeta(val.Http))
	}
	return hash.Sum(nil)
}
func HashNone(node *pb.None) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	return hash.Sum(nil)
}
func HashOneOf(node *pb.OneOf) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.Options) != 0 {
		hash.Write(intHashes[1])
		pairs := make ([]KeyValuePair, 0, len(node.Options))
		for k, v := range node.Options {
			pairs = append(pairs, KeyValuePair{Hash_Unicode(k), HashData(v)})
		}
		hash.Write(Hash_KeyValues(pairs))
	}
	if node.PotentialConflict != false {
		hash.Write(intHashes[2])
		hash.Write(Hash_Bool(node.PotentialConflict))
	}
	return hash.Sum(nil)
}
func HashOptional(node *pb.Optional) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if val, ok := node.Value.(*pb.Optional_Data); ok {
		hash.Write(intHashes[1])
		hash.Write(HashData(val.Data))
	}
	if val, ok := node.Value.(*pb.Optional_None); ok {
		hash.Write(intHashes[2])
		hash.Write(HashNone(val.None))
	}
	return hash.Sum(nil)
}
func HashPrimitive(node *pb.Primitive) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if val, ok := node.Value.(*pb.Primitive_Int64Value); ok {
		hash.Write(intHashes[6])
		hash.Write(HashInt64(val.Int64Value))
	}
	if val, ok := node.Value.(*pb.Primitive_StringValue); ok {
		hash.Write(intHashes[4])
		hash.Write(HashString(val.StringValue))
	}
	if node.TypeHint != "" {
		hash.Write(intHashes[1])
		hash.Write(Hash_Unicode(node.TypeHint))
	}
	if node.AkitaAnnotations != nil {
		hash.Write(intHashes[11])
		hash.Write(HashAkitaAnnotations(node.AkitaAnnotations))
	}
	if val, ok := node.Value.(*pb.Primitive_Int32Value); ok {
		hash.Write(intHashes[5])
		hash.Write(HashInt32(val.Int32Value))
	}
	if val, ok := node.Value.(*pb.Primitive_BoolValue); ok {
		hash.Write(intHashes[2])
		hash.Write(HashBool(val.BoolValue))
	}
	if len(node.Formats) != 0 {
		hash.Write(intHashes[13])
		pairs := make ([]KeyValuePair, 0, len(node.Formats))
		for k, v := range node.Formats {
			pairs = append(pairs, KeyValuePair{Hash_Unicode(k), Hash_Bool(v)})
		}
		hash.Write(Hash_KeyValues(pairs))
	}
	if val, ok := node.Value.(*pb.Primitive_FloatValue); ok {
		hash.Write(intHashes[10])
		hash.Write(HashFloat(val.FloatValue))
	}
	if node.FormatKind != "" {
		hash.Write(intHashes[14])
		hash.Write(Hash_Unicode(node.FormatKind))
	}
	if val, ok := node.Value.(*pb.Primitive_BytesValue); ok {
		hash.Write(intHashes[3])
		hash.Write(HashBytes(val.BytesValue))
	}
	if val, ok := node.Value.(*pb.Primitive_Uint64Value); ok {
		hash.Write(intHashes[8])
		hash.Write(HashUint64(val.Uint64Value))
	}
	if val, ok := node.Value.(*pb.Primitive_Uint32Value); ok {
		hash.Write(intHashes[7])
		hash.Write(HashUint32(val.Uint32Value))
	}
	if val, ok := node.Value.(*pb.Primitive_DoubleValue); ok {
		hash.Write(intHashes[9])
		hash.Write(HashDouble(val.DoubleValue))
	}
	if node.ContainsRandomValue != false {
		hash.Write(intHashes[12])
		hash.Write(Hash_Bool(node.ContainsRandomValue))
	}
	return hash.Sum(nil)
}
func HashPrimitiveRef(node *pb.PrimitiveRef) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if val, ok := node.Type.(*pb.PrimitiveRef_Uint32Type); ok {
		hash.Write(intHashes[6])
		hash.Write(HashUint32Type(val.Uint32Type))
	}
	if val, ok := node.Type.(*pb.PrimitiveRef_Int32Type); ok {
		hash.Write(intHashes[4])
		hash.Write(HashInt32Type(val.Int32Type))
	}
	if val, ok := node.Type.(*pb.PrimitiveRef_BoolType); ok {
		hash.Write(intHashes[1])
		hash.Write(HashBoolType(val.BoolType))
	}
	if val, ok := node.Type.(*pb.PrimitiveRef_Int64Type); ok {
		hash.Write(intHashes[5])
		hash.Write(HashInt64Type(val.Int64Type))
	}
	if val, ok := node.Type.(*pb.PrimitiveRef_BytesType); ok {
		hash.Write(intHashes[2])
		hash.Write(HashBytesType(val.BytesType))
	}
	if val, ok := node.Type.(*pb.PrimitiveRef_StringType); ok {
		hash.Write(intHashes[3])
		hash.Write(HashStringType(val.StringType))
	}
	if val, ok := node.Type.(*pb.PrimitiveRef_DoubleType); ok {
		hash.Write(intHashes[8])
		hash.Write(HashDoubleType(val.DoubleType))
	}
	if val, ok := node.Type.(*pb.PrimitiveRef_Uint64Type); ok {
		hash.Write(intHashes[7])
		hash.Write(HashUint64Type(val.Uint64Type))
	}
	if val, ok := node.Type.(*pb.PrimitiveRef_FloatType); ok {
		hash.Write(intHashes[9])
		hash.Write(HashFloatType(val.FloatType))
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
func HashStruct(node *pb.Struct) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if len(node.Fields) != 0 {
		hash.Write(intHashes[1])
		pairs := make ([]KeyValuePair, 0, len(node.Fields))
		for k, v := range node.Fields {
			pairs = append(pairs, KeyValuePair{Hash_Unicode(k), HashData(v)})
		}
		hash.Write(Hash_KeyValues(pairs))
	}
	if node.MapType != nil {
		hash.Write(intHashes[2])
		hash.Write(HashMapData(node.MapType))
	}
	return hash.Sum(nil)
}
func HashStructRef(node *pb.StructRef) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
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
	if node.Min != nil {
		hash.Write(intHashes[2])
		hash.Write(HashUInt32Value(node.Min))
	}
	if node.Max != nil {
		hash.Write(intHashes[3])
		hash.Write(HashUInt32Value(node.Max))
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
	if node.Min != nil {
		hash.Write(intHashes[2])
		hash.Write(HashUInt64Value(node.Min))
	}
	if node.Max != nil {
		hash.Write(intHashes[3])
		hash.Write(HashUInt64Value(node.Max))
	}
	return hash.Sum(nil)
}
func HashWitness(node *pb.Witness) []byte {
	hash := xxhash.New64()
	hash.Write([]byte("d"))
	if node.Method != nil {
		hash.Write(intHashes[1])
		hash.Write(HashMethod(node.Method))
	}
	return hash.Sum(nil)
}

var ProtobufFileHashes map[string][]byte = map[string][]byte{"method.proto": []byte{123, 92, 153, 152, 73, 68, 208, 226}, "witness.proto": []byte{42, 213, 185, 25, 124, 226, 76, 187}, "types.proto": []byte{98, 84, 34, 180, 249, 140, 214, 227}, "spec.proto": []byte{13, 101, 129, 126, 232, 252, 1, 146}}
