package spec_util

import (
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/pbhash"
)

type dataAndHash struct {
	hash string
	data *pb.Data
}

// Meld top-level args or responses map, where the keys are the hashes of the
// data. This means that we need to compare DataMeta to determine if we should
// meld two Data.
func meldTopLevelDataMap(dst, src map[string]*pb.Data) error {
	dstByMetaHash := map[string]dataAndHash{}
	for k, d := range dst {
		h, err := pbhash.HashProto(d.Meta)
		if err != nil {
			return errors.Wrapf(err, "failed to hash %v", d.Meta)
		}
		dstByMetaHash[h] = dataAndHash{hash: k, data: d}
	}

	results := make(map[string]*pb.Data, len(dstByMetaHash))
	for k, s := range src {
		h, err := pbhash.HashProto(s.Meta)
		if err != nil {
			return errors.Wrapf(err, "failed to hash %v", s.Meta)
		}

		if d, ok := dstByMetaHash[h]; ok {
			// d and s have the same DataMeta, meaning that they are refering to the
			// same HTTP field. Meld them.
			if err := MeldData(d.data, s); err != nil {
				return err
			}

			// Rehash because the proto has changed.
			dh, err := pbhash.HashProto(d.data)
			if err != nil {
				return errors.Wrapf(err, "failed to hash %v", d)
			}
			results[dh] = d.data

			delete(dstByMetaHash, h)
		} else {
			// The meld is additive - any new argument or response field is included.
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

	return nil
}

func isOptional(d *pb.Data) bool {
	_, isOptional := d.Value.(*pb.Data_Optional)
	return isOptional
}

func isNone(d *pb.Data) bool {
	if opt, ok := d.Value.(*pb.Data_Optional); ok {
		_, isNone := opt.Optional.Value.(*pb.Optional_None)
		return isNone
	}
	return false
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

func makeOptional(d *pb.Data) {
	if !isOptional(d) {
		d.Value = &pb.Data_Optional{
			Optional: &pb.Optional{
				Value: &pb.Optional_Data{
					Data: &pb.Data{Value: d.Value},
				},
			},
		}
	}
}

// Assumes that dst.Meta == src.Meta.
func MeldData(dst, src *pb.Data) (retErr error) {
	// Set to true if dst and src are recorded as a conflict.
	hasConflict := false
	defer func() {
		// Merge example values if there wasn't a conflict. Examples are merged in
		// the conflict handler.
		if !hasConflict && retErr == nil {
			mergeExampleValues(dst, src)
		}
	}()

	// Check if src is already a oneof. This can happen if src is the collapsed
	// element from a list originally containing elements with conflicting types.
	if srcOf, ok := src.Value.(*pb.Data_Oneof); ok {
		if v, ok := dst.Value.(*pb.Data_Oneof); ok {
			// If dst already encodes a conflict, merge the conflicts.
			for k, d := range srcOf.Oneof.Options {
				v.Oneof.Options[k] = d
			}
			return nil
		}

		// dst is just a regular value (which may happen to be a oneof). Swap src
		// and dst and re-use the logic below.
		dst.Value, src.Value = src.Value, dst.Value
	}

	// Special handling if src is optional.
	if srcOpt, srcIsOpt := src.Value.(*pb.Data_Optional); srcIsOpt {
		switch opt := srcOpt.Optional.Value.(type) {
		case *pb.Optional_Data:
			// Meld dst with the non-optional version of src first, then mark the
			// result as optional.
			if err := MeldData(dst, opt.Data); err != nil {
				return err
			}
			makeOptional(dst)
			return nil
		case *pb.Optional_None:
			// If src is a none, drop the none and mark the dst value as optional.
			makeOptional(dst)
			return nil
		}
	}

	switch v := dst.Value.(type) {
	case *pb.Data_Struct:
		// Special handling for struct to add unknown fields.
		if srcStruct, ok := src.Value.(*pb.Data_Struct); ok {
			return meldStruct(v.Struct, srcStruct.Struct)
		} else {
			hasConflict = true
			return recordConflict(dst, src)
		}
	case *pb.Data_List:
		if srcList, ok := src.Value.(*pb.Data_List); ok {
			return meldList(v.List, srcList.List)
		} else {
			hasConflict = true
			return recordConflict(dst, src)
		}
	case *pb.Data_Optional:
		switch opt := v.Optional.Value.(type) {
		case *pb.Optional_Data:
			// Meld src with the non-optional version of dst.
			return MeldData(opt.Data, src)
		case *pb.Optional_None:
			// If dst is a none, replace dst with an optional version of src.
			if isOptional(src) {
				dst.Value = src.Value
			} else {
				dst.Value = &pb.Data_Optional{
					Optional: &pb.Optional{
						Value: &pb.Optional_Data{
							Data: &pb.Data{Value: src.Value},
						},
					},
				}
			}
			return nil
		default:
			return recordConflict(dst, src)
		}
	case *pb.Data_Oneof:
		hasConflict = true
		// Add src as a new option after clearing its meta field since for
		// HTTP specs, oneof options all have the same metadata, recorded in the
		// Data.Meta field of the containing Data.
		srcNoMeta := proto.Clone(src).(*pb.Data)
		srcNoMeta.Meta = nil

		// See if we can meld the src into one of the options. For example,
		// melding struct into struct or list into list.
		// When we do this, we need to change the hash
		_, srcIsStruct := srcNoMeta.Value.(*pb.Data_Struct)
		_, srcIsList := srcNoMeta.Value.(*pb.Data_List)
		for oldHash, option := range v.Oneof.Options {
			switch option.Value.(type) {
			case *pb.Data_Struct:
				if srcIsStruct {
					return meldAndRehashOption(v.Oneof, oldHash, option, srcNoMeta)
				}
			case *pb.Data_List:
				if srcIsList {
					return meldAndRehashOption(v.Oneof, oldHash, option, srcNoMeta)
				}
			}
		}

		// Create a new conflict option.
		h, err := pbhash.HashProto(srcNoMeta)
		if err != nil {
			return errors.Wrapf(err, "failed to hash data: %v", srcNoMeta)
		}

		if existing, ok := v.Oneof.Options[h]; ok {
			// There might be an existing option with the same hash because we
			// ignore example values in the hash. If this is the case, merge
			// examples.
			mergeExampleValues(existing, src)
		} else {
			v.Oneof.Options[h] = srcNoMeta
		}
		return nil
	default:
		hasConflict = true
		return recordConflict(dst, src)
	}
}

// Meld a component of a OneOf that has been identified
// as a type-match (struct with struct or list with list.)
// This requires re-inserting it because the hash has been changed
func meldAndRehashOption(oneof *pb.OneOf, oldHash string, option *pb.Data, srcNoMeta *pb.Data) error {
	err := MeldData(option, srcNoMeta)
	if err != nil {
		return err
	}
	newHash, err := pbhash.HashProto(option)
	if err != nil {
		return err
	}
	if newHash != oldHash {
		delete(oneof.Options, oldHash)
		oneof.Options[newHash] = option
	}
	return nil
}

func dataEqual(dst, src *pb.Data) bool {
	srcExampleValues := src.ExampleValues
	dstExampleValues := dst.ExampleValues
	src.ExampleValues = nil
	dst.ExampleValues = nil

	defer func() {
		// Reinstate original example values
		src.ExampleValues = srcExampleValues
		dst.ExampleValues = dstExampleValues
	}()

	return proto.Equal(dst, src)
}

func recordConflict(dst, src *pb.Data) error {
	// Special case: If and only if one data has a type hint, assign it to the other
	// data so that the difference does not trigger a conflict and the type hint is preserved.
	{
		srcTypeHint := getTypeHint(src)
		dstTypeHint := getTypeHint(dst)
		if srcTypeHint != dstTypeHint {
			if dstTypeHint == "" {
				assignTypeHint(dst, srcTypeHint)
			} else if srcTypeHint == "" {
				assignTypeHint(src, dstTypeHint)
			}
		}
	}

	// Special case: If there are otherwise no conflicts, then merge data
	// formats.  However, if there are conflicts, then leave data formats
	// unmerged in their respective objects.
	srcDataFormats := getDataFormats(src)
	dstDataFormats := getDataFormats(dst)
	assignDataFormats(src, nil)
	assignDataFormats(dst, nil)

	// Special case: If there are otherwise no conflicts, then merge example
	// values. However, if there are conflicts, then leave example values
	// unmerged in their respective objects.
	srcExampleValues := src.ExampleValues
	dstExampleValues := dst.ExampleValues
	// Temporarily assign nil to example values for comparison.
	src.ExampleValues = nil
	dst.ExampleValues = nil

	// Note: witnesses should not contain actual values and we've taken out data
	// formats and example values above, so simple equality comparison works to
	// determine whether 2 Data proto have conflict.
	if !proto.Equal(dst, src) {
		// Reinstate original data formats
		assignDataFormats(src, srcDataFormats)
		assignDataFormats(dst, dstDataFormats)

		// Reinstate original example values
		src.ExampleValues = srcExampleValues
		dst.ExampleValues = dstExampleValues

		// New conflict detected. Create oneof to record the conflict.
		// For HTTP specs, oneof options all have the same metadata, recorded in
		// the Data.Meta field of the containing Data.
		dstNoMeta := proto.Clone(dst).(*pb.Data)
		dstNoMeta.Meta = nil
		srcNoMeta := proto.Clone(src).(*pb.Data)
		srcNoMeta.Meta = nil
		options := make(map[string]*pb.Data, 2)
		for _, d := range []*pb.Data{dstNoMeta, srcNoMeta} {
			h, err := pbhash.HashProto(d)
			if err != nil {
				return errors.Wrapf(err, "failed to hash data: %v", d)
			}
			options[h] = d
		}

		// Update dst to contain a conflict between dstNoMeta and srcNoMeta.
		dst.Value = &pb.Data_Oneof{
			&pb.OneOf{Options: options, PotentialConflict: true},
		}
		// Example values from dst are recorded inside the oneof as dstNoMeta.
		dst.ExampleValues = nil
	} else {
		// Merge data formats
		mergedDataFormats := make(map[string]bool, len(srcDataFormats)+len(dstDataFormats))
		for k := range srcDataFormats {
			mergedDataFormats[k] = true
		}
		for k := range dstDataFormats {
			mergedDataFormats[k] = true
		}
		if len(mergedDataFormats) > 0 {
			assignDataFormats(dst, mergedDataFormats)
		}

		// Reinstate source data formats, in order to leave src unmodified
		assignDataFormats(src, srcDataFormats)

		// Reinstate source example values, in order to leave src unmodified
		src.ExampleValues = srcExampleValues

		// Merge example values.
		dst.ExampleValues = dstExampleValues
		mergeExampleValues(dst, src)
	}
	return nil
}

func getTypeHint(d *pb.Data) string {
	switch x := d.Value.(type) {
	case *pb.Data_Primitive:
		return x.Primitive.TypeHint
	}
	return ""
}

func assignTypeHint(d *pb.Data, assignment string) {
	switch x := d.Value.(type) {
	case *pb.Data_Primitive:
		x.Primitive.TypeHint = assignment
	}
}

func getDataFormats(d *pb.Data) map[string]bool {
	switch x := d.Value.(type) {
	case *pb.Data_Primitive:
		return x.Primitive.Formats
	}
	return make(map[string]bool, 0)
}

func assignDataFormats(d *pb.Data, formats map[string]bool) {
	switch x := d.Value.(type) {
	case *pb.Data_Primitive:
		x.Primitive.Formats = formats
	}
}

func meldStruct(dst, src *pb.Struct) error {
	if isMap(dst) {
		if isMap(src) {
			return meldMap(dst, src)
		}

		// dst is a map, but src is not. Swap the two to reuse the logic for
		// melding a map into a struct.
		src.Fields, src.MapType, dst.Fields, dst.MapType = dst.Fields, dst.MapType, src.Fields, src.MapType
	}
	if isMap(src) {
		// Melding a map into a struct. Convert dst into a map and meld the two
		// maps.
		structToMap(dst)
		return meldMap(dst, src)
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
			if err := MeldData(dstData, srcData); err != nil {
				return errors.Wrapf(err, "failed to meld struct key %s", k)
			}
		} else {
			// Fields found in src but not in dst.
			makeOptional(srcData)
			dst.Fields[k] = srcData
		}
	}

	// Apply a heuristic for deciding when to convert structs to maps.
	if structShouldBeMap(dst) {
		structToMap(dst)
	}

	return nil
}

// Determines whether the given pb.Struct represents a map.
func isMap(struc *pb.Struct) bool {
	return struc.MapType != nil
}

// Tuning parameters for deciding when a struct should be turned into a map.
const maxOptionalFieldsPerStruct = 15
const maxFieldsPerStruct = 100

// Heuristically determines whether the given pb.Struct (assumed to not
// represent a map) should be a map.
func structShouldBeMap(struc *pb.Struct) bool {
	// A struct should be a map if its total number of fields exceeds
	// maxFieldsPerStruct.
	if len(struc.Fields) > maxFieldsPerStruct {
		return true
	}

	// A struct should be a map if its number of optional fields exceeds
	// maxOptionalFieldsPerStruct.
	numOptionalFields := 0
	for _, field := range struc.Fields {
		if field.GetOptional() != nil {
			numOptionalFields++
			if numOptionalFields > maxOptionalFieldsPerStruct {
				return true
			}
		}
	}

	return false
}

// Melds two maps together. The given pb.Structs are assumed to represent maps.
func meldMap(dst, src *pb.Struct) error {
	// Try to make the key and value in dst non-nil.
	if dst.MapType.Key == nil {
		src.MapType.Key, dst.MapType.Key = dst.MapType.Key, src.MapType.Key
	}
	if dst.MapType.Value == nil {
		src.MapType.Value, dst.MapType.Value = dst.MapType.Value, src.MapType.Value
	}

	// Meld keys.
	if src.MapType.Key != nil {
		if err := MeldData(dst.MapType.Key, src.MapType.Key); err != nil {
			return err
		}
	}

	// Meld values.
	if src.MapType.Value != nil {
		if err := MeldData(dst.MapType.Value, src.MapType.Value); err != nil {
			return err
		}
	}

	return nil
}

// Converts in place a pb.Struct (assumed to represent a struct) into a map.
func structToMap(struc *pb.Struct) {
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
		} else {
			MeldData(mapValue, curValue)
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

func meldList(dst, src *pb.List) error {
	srcOffset := 0
	if len(dst.Elems) == 0 {
		if len(src.Elems) == 0 {
			return nil
		}
		dst.Elems = []*pb.Data{src.Elems[0]}
		srcOffset = 1
	} else if len(dst.Elems) > 1 {
		for i := 1; i < len(dst.Elems); i++ {
			MeldData(dst.Elems[0], dst.Elems[i])
		}
		dst.Elems = dst.Elems[0:1]
	}

	for i, e := range src.Elems[srcOffset:] {
		if err := MeldData(dst.Elems[0], e); err != nil {
			return errors.Wrapf(err, "failed to meld list index %d", i)
		}
	}
	return nil
}
