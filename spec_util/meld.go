package spec_util

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/pbhash"
)

// Melds src into dst, resolving conflicts using oneof. Assumes that dst and src
// are for the same endpoint.
func MeldMethod(dst, src *pb.Method) error {
	if dst.Args == nil {
		dst.Args = src.Args
	} else if err := meldTopLevelDataMap(dst.Args, src.Args); err != nil {
		return errors.Wrap(err, "failed to meld arg map")
	}

	if dst.Responses == nil {
		dst.Responses = src.Responses
	} else if err := meldTopLevelDataMap(dst.Responses, src.Responses); err != nil {
		return errors.Wrap(err, "failed to meld response map")
	}

	return nil
}

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
	if dst.ExampleValues == nil {
		dst.ExampleValues = make(map[string]*pb.ExampleValue, len(src.ExampleValues))
	}
	for k, v := range src.ExampleValues {
		dst.ExampleValues[k] = v
	}
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
		_, srcIsStruct := srcNoMeta.Value.(*pb.Data_Struct)
		_, srcIsList := srcNoMeta.Value.(*pb.Data_List)
		for _, option := range v.Oneof.Options {
			switch option.Value.(type) {
			case *pb.Data_Struct:
				if srcIsStruct {
					return MeldData(option, srcNoMeta)
				}
			case *pb.Data_List:
				if srcIsList {
					return MeldData(option, srcNoMeta)
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
	if dst.Fields == nil {
		dst.Fields = src.Fields
		return nil
	}
	for k, srcData := range src.Fields {
		if dstData, ok := dst.Fields[k]; ok {
			if err := MeldData(dstData, srcData); err != nil {
				return errors.Wrapf(err, "failed to meld struct key %s", k)
			}
		} else {
			// The meld is additive - any new field is included.
			dst.Fields[k] = srcData
		}
	}
	return nil
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
