package http_rest

import "fmt"

// A path element for identifying the location of a field. See
// SpecVisitorContext.GetFieldLocation.
type FieldPathElement interface {
	String() string

	IsFieldName() bool
	IsArrayElement() bool
}

type fieldPathElementKind int

const (
	fieldNameKind fieldPathElementKind = iota
	arrayElementKind
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
