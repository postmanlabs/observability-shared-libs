package spec_util

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"math"
	"reflect"
	"strconv"
	"unicode/utf8"

	"github.com/OneOfOne/xxhash"
	"github.com/pkg/errors"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

func NewPrimitiveInt32(v int32) *pb.Primitive {
	return &pb.Primitive{
		Value: &pb.Primitive_Int32Value{
			Int32Value: &pb.Int32{Value: v},
		},
	}
}

func NewPrimitiveUint32(v uint32) *pb.Primitive {
	return &pb.Primitive{
		Value: &pb.Primitive_Uint32Value{
			Uint32Value: &pb.Uint32{Value: v},
		},
	}
}

func NewPrimitiveInt64(v int64) *pb.Primitive {
	return &pb.Primitive{
		Value: &pb.Primitive_Int64Value{
			Int64Value: &pb.Int64{Value: v},
		},
	}
}

func NewPrimitiveUint64(v uint64) *pb.Primitive {
	return &pb.Primitive{
		Value: &pb.Primitive_Uint64Value{
			Uint64Value: &pb.Uint64{Value: v},
		},
	}
}

func NewPrimitiveBool(v bool) *pb.Primitive {
	return &pb.Primitive{
		Value: &pb.Primitive_BoolValue{
			BoolValue: &pb.Bool{Value: v},
		},
	}
}

func NewPrimitiveFloat(v float32) *pb.Primitive {
	return &pb.Primitive{
		Value: &pb.Primitive_FloatValue{
			FloatValue: &pb.Float{Value: v},
		},
	}
}

func NewPrimitiveDouble(v float64) *pb.Primitive {
	return &pb.Primitive{
		Value: &pb.Primitive_DoubleValue{
			DoubleValue: &pb.Double{Value: v},
		},
	}
}

func NewPrimitiveString(v string) *pb.Primitive {
	return &pb.Primitive{
		Value: &pb.Primitive_StringValue{
			StringValue: &pb.String{Value: v},
		},
	}
}

func NewPrimitiveBytes(v []byte) *pb.Primitive {
	return &pb.Primitive{
		Value: &pb.Primitive_BytesValue{
			BytesValue: &pb.Bytes{Value: v},
		},
	}
}

// Interface for values that our Primitive protobuf can represent.
type PrimitiveValue interface {
	// Returns the zero value for this PrimitiveValue.
	Zero() PrimitiveValue

	// Returns the PrimitiveValue after obfuscating the original value while
	// keeping the type. For example, an int32 remains int32 after obfuscation
	// instead of becoming a []byte.
	Obfuscate() PrimitiveValue

	String() string

	ToProto() *pb.Primitive
	GoValue() interface{}

	// Make sure only spec_util package can generate PrimitiveValue.
	xxxSpecUtilPrimitiveValueImpl()
}

type primValueImpl struct {
	v interface{}
}

func (primValueImpl) xxxSpecUtilPrimitiveValueImpl() {}

// Returns the zero value for this Basic Value.
func (b primValueImpl) Zero() PrimitiveValue {
	return primValueImpl{v: reflect.Zero(reflect.TypeOf(b.v)).Interface()}
}

// DO NOT CHANGE without transition plan to support existing values hashed with
// the current implementation.
// Possible transition plan: record the implementation version in the
// akita_annotations section of the primitive proto.
func (b primValueImpl) Obfuscate() PrimitiveValue {
	h := sha256.New()
	h.Write([]byte("b0eadc4a-18a0-49ef-8da5-c840c1942647")) // fixed salt
	switch tv := b.v.(type) {
	case int32:
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], uint32(tv))
		h.Write(buf[:])
		return primValueImpl{v: int32(xxhash.Checksum32(h.Sum(nil)))}
	case uint32:
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], tv)
		h.Write(buf[:])
		return primValueImpl{v: xxhash.Checksum32(h.Sum(nil))}
	case int64:
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(tv))
		h.Write(buf[:])
		return primValueImpl{v: int64(xxhash.Checksum64(h.Sum(nil)))}
	case uint64:
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], tv)
		h.Write(buf[:])
		return primValueImpl{v: xxhash.Checksum64(h.Sum(nil))}
	case float32:
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], math.Float32bits(tv))
		h.Write(buf[:])
		return primValueImpl{v: math.Float32frombits(xxhash.Checksum32(h.Sum(nil)))}
	case float64:
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], math.Float64bits(tv))
		h.Write(buf[:])
		return primValueImpl{v: math.Float64frombits(xxhash.Checksum64(h.Sum(nil)))}
	case bool:
		// No point of obfuscating a bool.
		return b
	case string:
		h.Write([]byte(tv))
		// Use base64 encoding since when converting to protobuf, protobuf requires
		// the string to contain only UTF-8 characters.
		return primValueImpl{v: base64.URLEncoding.EncodeToString(h.Sum(nil))}
	case []byte:
		h.Write(tv)
		return primValueImpl{v: h.Sum(nil)}
	default:
		// This should never happen since only this package can generate
		// primValueImpl.
		panic(errors.Errorf("cannot obfuscate PrimitiveValue type %T", tv))
	}
}

func (b primValueImpl) String() string {
	switch tv := b.v.(type) {
	case int32:
		return strconv.FormatInt(int64(tv), 10)
	case uint32:
		return strconv.FormatUint(uint64(tv), 10)
	case int64:
		return strconv.FormatInt(tv, 10)
	case uint64:
		return strconv.FormatUint(tv, 10)
	case float32:
		return strconv.FormatFloat(float64(tv), 'E', -1, 32)
	case float64:
		return strconv.FormatFloat(tv, 'E', -1, 64)
	case bool:
		return strconv.FormatBool(tv)
	case string:
		return tv
	case []byte:
		return string(tv)
	default:
		// This should never happen since only this package can generate
		// primValueImpl.
		panic(errors.Errorf("cannot convert primitive value of type %T to string", tv))
	}
}

func (b primValueImpl) ToProto() *pb.Primitive {
	switch tv := b.v.(type) {
	case int32:
		return NewPrimitiveInt32(tv)
	case uint32:
		return NewPrimitiveUint32(tv)
	case int64:
		return NewPrimitiveInt64(tv)
	case uint64:
		return NewPrimitiveUint64(tv)
	case float32:
		return NewPrimitiveFloat(tv)
	case float64:
		return NewPrimitiveDouble(tv)
	case bool:
		return NewPrimitiveBool(tv)
	case string:
		return NewPrimitiveString(tv)
	case []byte:
		return NewPrimitiveBytes(tv)
	default:
		// This should never happen since only this package can generate
		// primValueImpl.
		panic(errors.Errorf("cannot convert value of type %T to primitive", tv))
	}
}

func (b primValueImpl) GoValue() interface{} {
	return b.v
}

func CategorizeString(str string) PrimitiveValue {
	// Prefer int64 over uint64 since the former can represent both positive and
	// negative integers, thereby creating fewer spurious diffs from integer
	// signedness. We also skip over float32 since float64 can represent
	// everything.
	if v, err := strconv.ParseInt(str, 10, 64); err == nil {
		return primValueImpl{v: v}
	} else if v, err := strconv.ParseUint(str, 10, 64); err == nil {
		return primValueImpl{v: v}
	} else if v, err := strconv.ParseFloat(str, 64); err == nil {
		return primValueImpl{v: v}
	} else if v, err := strconv.ParseBool(str); err == nil {
		return primValueImpl{v: v}
	}
	return nonUtf8StringWorkaround(str)
}

// A workaround to handle non-UTF-8 strings. Protobuf string can only represent
// UTF-8 values, so we treat strings containing invalid UTF-8 runes as bytes.
// https://app.clubhouse.io/akita-software/story/1427
func nonUtf8StringWorkaround(str string) PrimitiveValue {
	if utf8.ValidString(str) {
		return primValueImpl{v: str}
	}
	return primValueImpl{v: []byte(str)}
}

type InterpretStrings bool

const (
	// Indicates that strings should be interpreted as numbers or boolean values
	// wherever possible.
	INTERPRET_STRINGS InterpretStrings = true

	// Indicates that strings should remain strings.
	NO_INTERPRET_STRINGS InterpretStrings = false
)

func ToPrimitiveValue(v interface{}, interpretStrings InterpretStrings) (PrimitiveValue, error) {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Int:
		v = int64(v.(int))
	case reflect.Uint:
		v = uint64(v.(uint))
	case reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64, reflect.Float32, reflect.Float64:
		// Do nothing
	case reflect.Bool:
		// Do nothing
	case reflect.String:
		if interpretStrings {
			return CategorizeString(v.(string)), nil
		}
		return nonUtf8StringWorkaround(v.(string)), nil
	case reflect.Slice:
		if bs, ok := v.([]byte); ok {
			v = bs
		} else {
			return nil, errors.Errorf("unsupported primitive value type %T", v)
		}
	default:
		return nil, errors.Errorf("unsupported primitive value type %T", v)
	}
	return primValueImpl{v: v}, nil
}

func PrimitiveValueFromProto(p *pb.Primitive) (PrimitiveValue, error) {
	var val interface{}
	switch v := p.GetValue().(type) {
	case *pb.Primitive_BoolValue:
		val = v.BoolValue.GetValue()
	case *pb.Primitive_BytesValue:
		val = v.BytesValue.GetValue()
	case *pb.Primitive_StringValue:
		val = v.StringValue.GetValue()
	case *pb.Primitive_Int32Value:
		val = v.Int32Value.GetValue()
	case *pb.Primitive_Int64Value:
		val = v.Int64Value.GetValue()
	case *pb.Primitive_Uint32Value:
		val = v.Uint32Value.GetValue()
	case *pb.Primitive_Uint64Value:
		val = v.Uint64Value.GetValue()
	case *pb.Primitive_DoubleValue:
		val = v.DoubleValue.GetValue()
	case *pb.Primitive_FloatValue:
		val = v.FloatValue.GetValue()
	default:
		return nil, errors.Errorf("unsupported primtive protobuf type %T", v)
	}
	return primValueImpl{v: val}, nil
}
