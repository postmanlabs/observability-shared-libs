package spec_util

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/go-utils/optionals"
	"github.com/akitasoftware/go-utils/sets"

	"github.com/akitasoftware/akita-libs/spec_util/ir_hash"
)

type dataAndHash struct {
	hash string
	data *pb.Data
}

// Meld top-level args or responses map, where the keys are the hashes of the
// data. This means that we need to compare DataMeta to determine if we should
// meld two Data.
func meldTopLevelDataMap(dst, src map[string]*pb.Data, opts MeldOptions) (numCookiesDiscarded int, err error) {
	// The names of all cookies in dst.
	cookies := sets.NewSet[string]()

	// The names of all cookies that are discarded.
	cookiesDiscarded := sets.NewSet[string]()

	// Index the destination's Data objects by the hash of their DataMeta objects.
	// Count the number of cookies already in the destination. If it is over the
	// limit, discard the excess cookies.
	dstByMetaHash := map[string]dataAndHash{}
	for k, d := range dst {
		if d.Meta == nil {
			return numCookiesDiscarded, fmt.Errorf("missing Meta in top-level dst Data %q", k)
		}

		// Track the names of the cookies in dst. If dst has too many cookies,
		// discard the extras.
		if cookie := d.GetMeta().GetHttp().GetCookie(); cookie != nil {
			cookieName := cookie.Key
			isNewCookie := !cookies.Contains(cookieName)
			if isNewCookie {
				if maxNumCookies, exists := opts.MaxNumCookies.Get(); exists && len(cookies) >= maxNumCookies {
					// Dest has too many cookies. Remove and skip this cookie.
					cookiesDiscarded.Insert(cookieName)
					delete(dst, k)
					continue
				}

				cookies.Insert(cookieName)
			}
		}

		h := ir_hash.HashDataMetaToString(d.Meta)
		dstByMetaHash[h] = dataAndHash{hash: k, data: d}
	}

	results := make(map[string]*pb.Data, len(dstByMetaHash))
	for k, s := range src {
		if s.Meta == nil {
			return numCookiesDiscarded, fmt.Errorf("missing Meta in top-level src Data %q", k)
		}
		h := ir_hash.HashDataMetaToString(s.Meta)

		if d, ok := dstByMetaHash[h]; ok {
			// d and s have the same DataMeta, meaning that they are referring to the
			// same HTTP field. Meld them.
			if err := MeldData(d.data, s); err != nil {
				return len(cookiesDiscarded), err
			}

			// Rehash because the proto has changed.
			dh := ir_hash.HashDataToString(d.data)
			results[dh] = d.data

			delete(dstByMetaHash, h)
		} else {
			// The meld is additive - any new argument or response field is included,
			// but we need to discard excess cookies.
			if cookie := s.GetMeta().GetHttp().GetCookie(); cookie != nil {
				cookieName := cookie.Key
				isNewCookie := !cookies.Contains(cookieName)
				if isNewCookie {
					if maxNumCookies, exists := opts.MaxNumCookies.Get(); exists && len(cookies) >= maxNumCookies {
						// Too many cookies.
						cookiesDiscarded.Insert(cookieName)
						continue
					}
				}
			}

			results[k] = s
		}
	}

	// Add any dst values without matching meta from src.
	for _, d := range dstByMetaHash {
		results[d.hash] = d.data
	}

	// Clear the original dst and replace with new results.
	for k := range dst {
		delete(dst, k)
	}
	for k, v := range results {
		dst[k] = v
	}

	return len(cookiesDiscarded), nil
}

func isOptional(d *pb.Data) bool {
	_, isOptional := d.Value.(*pb.Data_Optional)
	return isOptional
}

func mergeExampleValues(dst, src *pb.Data) {
	examples := make(map[string]*pb.ExampleValue, 2)

	// Get all (unique) example keys.
	keySet := make(map[string]struct{}, len(src.ExampleValues)+len(dst.ExampleValues))
	exampleMaps := []map[string]*pb.ExampleValue{dst.ExampleValues, src.ExampleValues}
	for _, exampleMap := range exampleMaps {
		for k := range exampleMap {
			keySet[k] = struct{}{}
		}
	}
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}

	// Keep the two smallest example keys, discard the rest.
	sort.Strings(keys)
	for _, k := range keys {
		if v, ok := src.ExampleValues[k]; ok {
			examples[k] = v
		} else if v, ok := dst.ExampleValues[k]; ok {
			examples[k] = v
		}

		if len(examples) >= 2 {
			break
		}
	}

	dst.ExampleValues = examples
}

// Makes given Data optional if it isn't already.  If it is pb.None,
// make it Optional<None>, which reflects that this field may be
// absent (optional) and that it may also accept null values (None).
func makeOptional(d *pb.Data) {
	isNone := d.GetOptional().GetNone() != nil
	if !isOptional(d) || isNone {
		// Maintain the invariant that the outermost Data object (the outer
		// Optional, in this case) holds the nullable bit.
		//
		// d retains its nullable bit, even as its value is changed
		// to be optional.
		d.Value = &pb.Data_Optional{
			Optional: &pb.Optional{
				Value: &pb.Optional_Data{
					Data: &pb.Data{Value: d.Value},
				},
			},
		}

		if isNone {
			d.Nullable = true
		}
	}
}

// Assumes that dst.Meta == src.Meta.
func MeldData(dst, src *pb.Data) (retErr error) {
	melder := &melder{mergeTracking: true}
	return melder.meldData(dst, src)
}

// Assumes that dst.Meta == src.Meta.
// Melds src into dst.  Leaves tracking data in dst untouched.
func MeldDataIgnoreTracking(dst, src *pb.Data) (retErr error) {
	melder := &melder{mergeTracking: false}
	return melder.meldData(dst, src)
}

type melder struct {
	// If true, sums tracking data on meld.  Otherwise leaves
	// tracking data unmodified in dst.
	mergeTracking bool
}

// If the given src and dst have the following invariant on all OneOfs contained
// within, then this is preserved.
//
//   - At most one variant in the OneOf is a struct.
//   - At most one variant in the OneOf is a list.
//   - All other variants in the OneOf is a primitive.
//
// If the given src and dst have the following invariant on all OneOfs and
// Optionals, then this is preserved.
//
//   - If any children are nullable, then the nullable bit is set on
//     the parent, not the children.
//
// Assumes that dst.Meta == src.Meta.
//
// XXX: In some cases, this modifies src as well as dst :/
func (m *melder) meldData(dst, src *pb.Data) (retErr error) {
	// Set to true if dst and src are recorded as a conflict.
	hasConflict := false
	defer func() {
		// Merge example values if there wasn't a conflict. Examples are merged in
		// the conflict handler.
		if !hasConflict && retErr == nil {
			mergeExampleValues(dst, src)
		}
	}()

	// If src or dst is nullable, then melded(dst, src) is too.
	dst.Nullable = dst.Nullable || src.Nullable

	// Check if src is already a oneof. This can happen if src is the collapsed
	// element from a list originally containing elements with conflicting types.
	if srcOf, ok := src.Value.(*pb.Data_Oneof); ok {
		if _, ok := dst.Value.(*pb.Data_Oneof); ok {
			// dst already encodes a conflict. Merge the conflicts.
			return m.meldOneOf(dst, srcOf.Oneof)
		}

		// dst is not a oneof. Swap src and dst and re-use the logic below.
		//
		// XXX Modifies src. Would fixing this have undesired downstream effects?
		dst.Value, src.Value = src.Value, dst.Value
	}

	// Special handling if src is optional.
	if srcOpt, srcIsOpt := src.Value.(*pb.Data_Optional); srcIsOpt {
		switch opt := srcOpt.Optional.Value.(type) {
		case *pb.Optional_Data:
			// Meld dst with the non-optional version of src first, then mark the
			// result as optional.
			if err := m.meldData(dst, opt.Data); err != nil {
				return err
			}
			makeOptional(dst)
			return nil
		case *pb.Optional_None:
			// If src is a none, drop the none and mark the dst value as nullable.
			dst.Nullable = true
			return nil
		default:
			return fmt.Errorf("unknown optional value type: %s", reflect.TypeOf(srcOpt.Optional.Value).Name())
		}
	}

	// At this point, src should be neither a one-of nor an optional.

	switch v := dst.Value.(type) {
	case *pb.Data_Struct:
		// Special handling for struct to add unknown fields.
		if srcStruct, ok := src.Value.(*pb.Data_Struct); ok {
			return m.meldStruct(v.Struct, srcStruct.Struct)
		} else {
			hasConflict = true
			return m.recordConflict(dst, src)
		}
	case *pb.Data_List:
		if srcList, ok := src.Value.(*pb.Data_List); ok {
			return m.meldList(v.List, srcList.List)
		} else {
			hasConflict = true
			return m.recordConflict(dst, src)
		}
	case *pb.Data_Optional:
		switch opt := v.Optional.Value.(type) {
		case *pb.Optional_Data:
			// Meld src with the non-optional version of dst.
			err := m.meldData(opt.Data, src)
			if err != nil {
				return err
			}

			// Lift the nullable bit to dst from the updated child,
			// and clear the bit from the child.
			dst.Nullable = dst.Nullable || opt.Data.Nullable
			opt.Data.Nullable = false
			return err
		case *pb.Optional_None:
			// If dst is a none, replace dst with a nullable version of src.
			dst.Nullable = true
			dst.Value = src.Value
			return nil
		default:
			return fmt.Errorf("unknown optional value type: %s", reflect.TypeOf(v.Optional.Value).Name())
		}
	case *pb.Data_Oneof:
		hasConflict = true
		return m.meldOneOfVariant(dst, optionals.None[string](), src)
	default:
		hasConflict = true
		return m.recordConflict(dst, src)
	}
}

// Meld a component of a OneOf that has been identified
// as a type-match (struct with struct or list with list.)
// This requires re-inserting it because the hash has been changed
func (m *melder) meldAndRehashOption(oneof *pb.OneOf, oldHash string, option *pb.Data, srcNoMeta *pb.Data) error {
	err := m.meldData(option, srcNoMeta)
	if err != nil {
		return err
	}
	newHash := ir_hash.HashDataToString(option)
	if newHash != oldHash {
		delete(oneof.Options, oldHash)
		oneof.Options[newHash] = option
	}
	return nil
}

// Two prims have compatible types if their base types join (in their
// Value field) and they have the same data format kind, if any.
func haveCompatibleTypes(dst, src *pb.Primitive) bool {
	if dst == nil || src == nil {
		return false
	}

	// First, check that the base types can join.
	baseJoin := joinBaseTypes(dst, src)
	if baseJoin == nil {
		return false
	}

	// Types are compatible if the base types can join and the format kinds
	// are equal.
	if dst.FormatKind == src.FormatKind {
		return true
	}

	// Types are compatible if the base types can join and least one type does
	// not have any data formats identified.
	if dst.FormatKind == "" || src.FormatKind == "" {
		return true
	}

	return false
}

// Returns a new Primitive with the Value set as the type-theoretic join of
// dst.Value and src.Value.  For example, join(int32, uint32) = int64.
// Returns nil if no such join exists, e.g. join(int64, uint64) = nil.
func joinBaseTypes(dst, src *pb.Primitive) *pb.Primitive {
	dstType := reflect.TypeOf(dst.Value)
	srcType := reflect.TypeOf(src.Value)

	if dstType == srcType {
		return &pb.Primitive{Value: dst.Value}
	}

	// NOTE(cns): When the CLI builds witnesses from wire traffic, it parses integers
	// as int64 whenever possible and only falls back to uint64 for values >= 2^63.
	// However, we could see other behavior from uploaded specs.
	switch dst.Value.(type) {
	case *pb.Primitive_Int32Value:
		switch src.Value.(type) {
		case *pb.Primitive_Int64Value, *pb.Primitive_Uint32Value:
			return &pb.Primitive{Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}}}
		case *pb.Primitive_FloatValue:
			return &pb.Primitive{Value: &pb.Primitive_FloatValue{FloatValue: &pb.Float{}}}
		case *pb.Primitive_DoubleValue:
			return &pb.Primitive{Value: &pb.Primitive_DoubleValue{DoubleValue: &pb.Double{}}}
		}
	case *pb.Primitive_Int64Value:
		switch src.Value.(type) {
		case *pb.Primitive_Int32Value, *pb.Primitive_Uint32Value:
			return &pb.Primitive{Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}}}
		case *pb.Primitive_FloatValue, *pb.Primitive_DoubleValue:
			return &pb.Primitive{Value: &pb.Primitive_DoubleValue{DoubleValue: &pb.Double{}}}
		}
	case *pb.Primitive_Uint32Value:
		switch src.Value.(type) {
		case *pb.Primitive_Int32Value:
			return &pb.Primitive{Value: &pb.Primitive_Int64Value{Int64Value: &pb.Int64{}}}
		case *pb.Primitive_Uint64Value:
			return &pb.Primitive{Value: &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}}}
		case *pb.Primitive_FloatValue:
			return &pb.Primitive{Value: &pb.Primitive_FloatValue{FloatValue: &pb.Float{}}}
		case *pb.Primitive_DoubleValue:
			return &pb.Primitive{Value: &pb.Primitive_DoubleValue{DoubleValue: &pb.Double{}}}
		}
	case *pb.Primitive_Uint64Value:
		switch src.Value.(type) {
		case *pb.Primitive_Uint32Value:
			return &pb.Primitive{Value: &pb.Primitive_Uint64Value{Uint64Value: &pb.Uint64{}}}
		case *pb.Primitive_FloatValue, *pb.Primitive_DoubleValue:
			return &pb.Primitive{Value: &pb.Primitive_DoubleValue{DoubleValue: &pb.Double{}}}
		}
	case *pb.Primitive_FloatValue:
		switch src.Value.(type) {
		case *pb.Primitive_Int32Value, *pb.Primitive_Uint32Value:
			return &pb.Primitive{Value: &pb.Primitive_FloatValue{FloatValue: &pb.Float{}}}
		case *pb.Primitive_Int64Value, *pb.Primitive_Uint64Value, *pb.Primitive_DoubleValue:
			return &pb.Primitive{Value: &pb.Primitive_DoubleValue{DoubleValue: &pb.Double{}}}
		}
	case *pb.Primitive_DoubleValue:
		switch src.Value.(type) {
		case *pb.Primitive_Int32Value, *pb.Primitive_Uint32Value, *pb.Primitive_Int64Value, *pb.Primitive_Uint64Value, *pb.Primitive_FloatValue:
			return &pb.Primitive{Value: &pb.Primitive_DoubleValue{DoubleValue: &pb.Double{}}}
		}
	}

	return nil
}

// Merges src into dst.  Introduces a OneOf when dst and src are different
// types, e.g. string/int, list/object, list/int, or if they are the same
// type but have different data format kinds.  (Different data formats of
// the same kind are merged.)
//
// Assumes dst and src are different base types or are both primitives.
func (m *melder) recordConflict(dst, src *pb.Data) error {
	// If src and dst are Primitives, we meld them if they have the same type
	// (in their Value field) and the same data format kind, if any.
	// Otherwise, src and dst are in conflict, and we introduce a OneOf.
	dstPrim := dst.GetPrimitive()
	srcPrim := src.GetPrimitive()
	arePrims := dstPrim != nil && srcPrim != nil

	if arePrims && haveCompatibleTypes(dstPrim, srcPrim) {
		// No conflict.  Merge primitive metadata.
		err := m.meldPrimitive(dstPrim, srcPrim)
		if err != nil {
			return err
		}
		mergeExampleValues(dst, src)
	} else {
		// New conflict detected. Create oneof to record the conflict.
		// For HTTP specs, oneof options all have the same metadata, recorded in
		// the Data.Meta field of the containing Data.
		dstNoMeta := proto.Clone(dst).(*pb.Data)
		dstNoMeta.Meta = nil
		srcNoMeta := proto.Clone(src).(*pb.Data)
		srcNoMeta.Meta = nil
		options := make(map[string]*pb.Data, 2)
		for _, d := range []*pb.Data{dstNoMeta, srcNoMeta} {
			h := ir_hash.HashDataToString(d)
			options[h] = d
		}

		// Update dst to contain a conflict between dstNoMeta and srcNoMeta.
		dst.Value = &pb.Data_Oneof{
			Oneof: &pb.OneOf{Options: options, PotentialConflict: true},
		}
		// Example values from dst are recorded inside the oneof as dstNoMeta.
		dst.ExampleValues = nil
	}

	return nil
}

func (m *melder) meldStruct(dst, src *pb.Struct) error {
	if isMap(dst) {
		if isMap(src) {
			return m.meldMap(dst, src)
		}

		// dst is a map, but src is not. Swap the two to reuse the logic for
		// melding a map into a struct.
		src.Fields, src.MapType, dst.Fields, dst.MapType = dst.Fields, dst.MapType, src.Fields, src.MapType
	}
	if isMap(src) {
		// Melding a map into a struct. Convert dst into a map and meld the two
		// maps.
		m.structToMap(dst)
		return m.meldMap(dst, src)
	}

	// If a field appears in both structs, it is assumed to be required.
	// If it appears in one, but not the other, then it should become
	// optional (if not optional already.)

	if dst.Fields == nil {
		dst.Fields = src.Fields
		return nil
	}
	for k, dstData := range dst.Fields {
		if _, ok := src.Fields[k]; !ok {
			// Fields in dst but not in src.
			makeOptional(dstData)
		}
	}
	for k, srcData := range src.Fields {
		if dstData, ok := dst.Fields[k]; ok {
			// Found in both, MeldData handles if either is already
			// optional.
			if err := m.meldData(dstData, srcData); err != nil {
				return errors.Wrapf(err, "failed to meld struct key %s", k)
			}
		} else {
			// Fields found in src but not in dst.
			makeOptional(srcData)
			dst.Fields[k] = srcData
		}
	}

	// Apply a heuristic for deciding when to convert structs to maps.
	if StructShouldBeMap(dst) {
		m.structToMap(dst)
	}

	return nil
}

// Determines whether the given pb.Struct represents a map.
func isMap(struc *pb.Struct) bool {
	return struc.MapType != nil
}

// Melds two maps together. The given pb.Structs are assumed to represent maps.
func (m *melder) meldMap(dst, src *pb.Struct) error {
	// Try to make the key and value in dst non-nil.
	if dst.MapType.Key == nil {
		src.MapType.Key, dst.MapType.Key = dst.MapType.Key, src.MapType.Key
	}
	if dst.MapType.Value == nil {
		src.MapType.Value, dst.MapType.Value = dst.MapType.Value, src.MapType.Value
	}

	// Meld keys.
	if src.MapType.Key != nil {
		if err := m.meldData(dst.MapType.Key, src.MapType.Key); err != nil {
			return err
		}
	}

	// Meld values.
	if src.MapType.Value != nil {
		if err := m.meldData(dst.MapType.Value, src.MapType.Value); err != nil {
			return err
		}
	}

	return nil
}

// Converts in place a pb.Struct (assumed to represent a struct) into a map.
func (m *melder) structToMap(struc *pb.Struct) {
	// The map's value Data is obtained by melding all field types together into
	// a single Data, while stripping away any optionality.
	var mapKey *pb.Data
	var mapValue *pb.Data
	for fieldName, curValue := range struc.Fields {
		if mapKey == nil {
			// TODO: Infer a data format from the field's name and meld map keys.
			// For now, just hard-code map keys as unformatted strings.
			_ = fieldName

			// ugh
			mapKey = &pb.Data{
				Value: &pb.Data_Primitive{
					Primitive: &pb.Primitive{
						Value: &pb.Primitive_StringValue{
							StringValue: &pb.String{},
						},
					},
				},
			}
		}

		// Strip any optionality from the current field's value and meld into the
		// map's value.
		curValue = stripOptional(curValue)
		if mapValue == nil {
			mapValue = curValue
			//} else if curValue != nil {
		} else if curValue != nil {
			m.meldData(mapValue, curValue)
		}
	}

	struc.Fields = nil
	struc.MapType = &pb.MapData{
		Key:   mapKey,
		Value: mapValue,
	}
}

// Strips away one layer of optionality from the given Data. If the given Data
// is non-optional, it is returned.
func stripOptional(data *pb.Data) *pb.Data {
	optional := data.GetOptional()
	if optional == nil {
		return data
	}
	return optional.GetData()
}

func (m *melder) meldList(dst, src *pb.List) error {
	srcOffset := 0
	if len(dst.Elems) == 0 {
		if len(src.Elems) == 0 {
			return nil
		}
		dst.Elems = []*pb.Data{src.Elems[0]}
		srcOffset = 1
	} else if len(dst.Elems) > 1 {
		for i := 1; i < len(dst.Elems); i++ {
			m.meldData(dst.Elems[0], dst.Elems[i])
		}
		dst.Elems = dst.Elems[0:1]
	}

	for i, e := range src.Elems[srcOffset:] {
		if err := m.meldData(dst.Elems[0], e); err != nil {
			return errors.Wrapf(err, "failed to meld list index %d", i)
		}
	}
	return nil
}

// Assumes haveCompatibleTypes(dst, src), returns an error otherwise.
// Meld data formats, tracking data, etc. from src to dst.
// XXX(cns): In some cases, this modifies src as well as dst :/
func (m *melder) meldPrimitive(dst, src *pb.Primitive) error {
	// Special case: If and only if one data has a type hint, assign it to the other
	// data so that the difference does not trigger a conflict and the type hint is preserved.
	// XXX(cns): This modifies src!  Not ideal, but I don't know if it's safe
	// to remove this behavior.
	{
		if src.TypeHint != dst.TypeHint {
			if dst.TypeHint == "" {
				dst.TypeHint = src.TypeHint
			} else if src.TypeHint == "" {
				src.TypeHint = dst.TypeHint
			}
		}
	}

	// Join base types.
	baseJoin := joinBaseTypes(dst, src)
	if baseJoin == nil {
		return errors.Errorf("failed to join base types")
	}
	dst.Value = baseJoin.Value

	// If either side has no data formats (i.e. is just a base type), then
	// the resulting meld will similarly have no data formats.  This implements
	// a type-theoretic join, as the base type without formats subsumes one
	// restricted to specific formats.
	if dst.FormatKind == "" || src.FormatKind == "" {
		dst.FormatKind = ""
		dst.Formats = nil
	} else if dst.FormatKind == src.FormatKind {
		// Merge data formats
		mergedDataFormats := make(map[string]bool, len(src.Formats)+len(dst.Formats))
		for k := range src.Formats {
			mergedDataFormats[k] = true
		}
		for k := range dst.Formats {
			mergedDataFormats[k] = true
		}
		if len(mergedDataFormats) > 0 {
			dst.Formats = mergedDataFormats
		}
	} else {
		return errors.Errorf("failed to meld primitives because format kinds are not equal")
	}

	return nil
}

// Assumes dst represents a one-of, and melds in the variants of src. On return,
// if the result of the meld is no longer a one-of (e.g., if melding caused the
// variants' types to be unified), then dst will no longer represent a one-of.
//
// An error occurs if the given dst does not represent a one-of.
func (m *melder) meldOneOf(dst *pb.Data, src *pb.OneOf) error {
	for srcHash, srcVariant := range src.Options {
		if err := m.meldOneOfVariant(dst, optionals.Some(srcHash), srcVariant); err != nil {
			return err
		}
	}

	return nil
}

// Assumes dst represents a one-of, and melds in the given variant. On return,
// if the result of the meld is no longer a one-of (e.g., if melding caused the
// variants' types to be unified), then dst will no longer represent a one-of.
//
// An error occurs if the given dst does not represent a one-of.
func (m *melder) meldOneOfVariant(dst *pb.Data, srcHashOpt optionals.Optional[string], srcVariant *pb.Data) error {
	// Check that dst represents a one-of.
	dstOneOf := dst.GetOneof()
	if dstOneOf == nil {
		return fmt.Errorf("meldOneOfVariant called with a %s, not a one-of", reflect.TypeOf(dst.Value))
	}

	// Make sure the meta and nullable fields of srcVariant are cleared. For HTTP
	// specs, OneOf variants all have the same metadata, recorded in the Data.Meta
	// field of the containing Data.  The nullable bit was copied to the dst Data
	// object in meldData before calling meldOneOfVariant.
	if srcVariant.Meta != nil || srcVariant.Nullable {
		srcVariant = proto.Clone(srcVariant).(*pb.Data)
		srcVariant.Meta = nil
		srcVariant.Nullable = false

		// We'll recompute the hash.
		srcHashOpt = optionals.None[string]()
	}

	// Hash if needed.
	srcHash := srcHashOpt.GetOrComputeNoError(func() string {
		return ir_hash.HashDataToString(srcVariant)
	})

	// There might be an existing option with the same hash because we ignore
	// example values in the hash. If this is the case, just merge examples.
	if existing, ok := dstOneOf.Options[srcHash]; ok {
		mergeExampleValues(existing, srcVariant)
		return nil
	}

	// See if we can meld the srcVariant into one of the existing variants. For
	// example, melding struct into struct or list into list. When we do this, we
	// need to change the hash.
	switch srcVariant.Value.(type) {
	case *pb.Data_Struct:
		// If the destination has a struct variant, merge with that.
		for oldDstHash, dstVariant := range dstOneOf.Options {
			if _, dstIsStruct := dstVariant.Value.(*pb.Data_Struct); dstIsStruct {
				return m.meldAndRehashOption(dstOneOf, oldDstHash, dstVariant, srcVariant)
			}
		}

	case *pb.Data_List:
		// If the destination has a list variant, merge with that.
		for oldDstHash, dstVariant := range dstOneOf.Options {
			if _, dstIsList := dstVariant.Value.(*pb.Data_List); dstIsList {
				return m.meldAndRehashOption(dstOneOf, oldDstHash, dstVariant, srcVariant)
			}
		}

	case *pb.Data_Primitive:
		// The destination might have more than one primitive variant whose type is
		// compatible with the source variant. If the source variant has no data
		// formats (i.e., is just a base type), then all these variants will
		// collapse down to a single variant when melded. Therefore, using
		// srcVariant as an accumulator, we remove all compatible variants from the
		// destination and meld them together with the source variant. Later on, we
		// will add the resulting source variant to the destination.
		needToRehashSrc := false
		for dstHash, dstVariant := range dstOneOf.Options {
			if _, dstIsPrim := dstVariant.Value.(*pb.Data_Primitive); dstIsPrim && haveCompatibleTypes(dstVariant.GetPrimitive(), srcVariant.GetPrimitive()) {
				delete(dstOneOf.Options, dstHash)

				// To avoid modifying *srcVariant, we meld srcVariant into dstVariant,
				// and then set srcVariant = dstVariant.
				if err := m.meldData(dstVariant, srcVariant); err != nil {
					return err
				}
				srcVariant = dstVariant
				needToRehashSrc = true
			}
		}

		if needToRehashSrc {
			srcHash = ir_hash.HashDataToString(srcVariant)
		}

	default:
		return fmt.Errorf("unknown one-of variant type: %s", reflect.TypeOf(srcVariant.Value).Name())
	}

	if len(dstOneOf.Options) == 0 {
		// No longer have a one-of. Replace dst's value with the singleton variant.
		dst.Value = srcVariant.Value
		return nil
	}

	// Add a new variant.
	dstOneOf.Options[srcHash] = srcVariant
	return nil
}
