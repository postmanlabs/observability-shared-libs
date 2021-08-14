package http_rest

import "fmt"

// A path element for identifying the location of a field. See
// SpecVisitorContext.GetFieldLocation.
type FieldPathElement interface {
	String() string

	IsFieldName() bool
	IsArrayElement() bool
	IsOneOfVariant() bool
	IsMapKeyType() bool
	IsMapValueType() bool
}

type fieldPathElementKind int

const (
	fieldNameKind fieldPathElementKind = iota
	arrayElementKind
	oneOfVariantKind
	mapKeyTypeKind
	mapValueTypeKind
)

type abstractFieldPathElement struct {
	kind fieldPathElementKind
}

func (elt *abstractFieldPathElement) IsFieldName() bool {
	return elt.kind == fieldNameKind
}

func (elt *abstractFieldPathElement) IsArrayElement() bool {
	return elt.kind == arrayElementKind
}

func (elt *abstractFieldPathElement) IsOneOfVariant() bool {
	return elt.kind == oneOfVariantKind
}

func (elt *abstractFieldPathElement) IsMapKeyType() bool {
	return elt.kind == mapKeyTypeKind
}

func (elt *abstractFieldPathElement) IsMapValueType() bool {
	return elt.kind == mapValueTypeKind
}

// Identifies a field of an object.
type FieldName struct {
	abstractFieldPathElement

	Name string
}

var _ FieldPathElement = (*FieldName)(nil)

func NewFieldName(name string) *FieldName {
	return &FieldName{
		abstractFieldPathElement: abstractFieldPathElement{
			kind: fieldNameKind,
		},
		Name: name,
	}
}

func (f *FieldName) String() string {
	return f.Name
}

// Identifies an element of an array.
type ArrayElement struct {
	abstractFieldPathElement

	Index int
}

var _ FieldPathElement = (*ArrayElement)(nil)

func NewArrayElement(index int) *ArrayElement {
	return &ArrayElement{
		abstractFieldPathElement: abstractFieldPathElement{
			kind: arrayElementKind,
		},
		Index: index,
	}
}

func (ae *ArrayElement) String() string {
	return fmt.Sprintf("[%d]", ae.Index)
}

// Identifies a branch of a variant ("one of").
type OneOfVariant struct {
	abstractFieldPathElement

	// Identifies the variant being represented.
	Index int

	// The number of possible variants.
	NumVariants int
}

var _ FieldPathElement = (*OneOfVariant)(nil)

func NewOneOfVariant(index int, numVariants int) *OneOfVariant {
	return &OneOfVariant{
		abstractFieldPathElement: abstractFieldPathElement{
			kind: oneOfVariantKind,
		},
		Index:       index,
		NumVariants: numVariants,
	}
}

func (oov *OneOfVariant) String() string {
	return fmt.Sprintf("(format %d of %d)", oov.Index, oov.NumVariants)
}

// Identifies a map's key type.
type MapKeyType struct {
	abstractFieldPathElement
}

var _ FieldPathElement = (*MapKeyType)(nil)

func NewMapKeyType() *MapKeyType {
	return &MapKeyType{
		abstractFieldPathElement: abstractFieldPathElement{
			kind: mapKeyTypeKind,
		},
	}
}

func (mk *MapKeyType) String() string {
	return "keys"
}

// Identifies a map's value type.
type MapValueType struct {
	abstractFieldPathElement
}

var _ FieldPathElement = (*MapValueType)(nil)

func NewMapValueType() *MapValueType {
	return &MapValueType{
		abstractFieldPathElement: abstractFieldPathElement{
			kind: mapValueTypeKind,
		},
	}
}

func (mv *MapValueType) String() string {
	return "values"
}
