package http_rest

import "fmt"

// A path element for identifying the location of a field. See
// SpecVisitorContext.GetFieldLocation.
type FieldLocationElement interface {
	String() string

	IsFieldName() bool
	IsArrayElement() bool
}

type fieldLocationElementKind int

const (
	fieldNameKind fieldLocationElementKind = iota
	arrayElementKind
)

type abstractFieldLocationElement struct {
	kind fieldLocationElementKind
}

func (elt *abstractFieldLocationElement) IsFieldName() bool {
	return elt.kind == fieldNameKind
}

func (elt *abstractFieldLocationElement) IsArrayElement() bool {
	return elt.kind == arrayElementKind
}

// Identifies a field of an object.
type FieldName struct {
	abstractFieldLocationElement

	Name string
}

var _ FieldLocationElement = (*FieldName)(nil)

func NewFieldName(name string) *FieldName {
	return &FieldName{
		abstractFieldLocationElement: abstractFieldLocationElement{
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
	abstractFieldLocationElement

	Index int
}

var _ FieldLocationElement = (*ArrayElement)(nil)

func NewArrayElement(index int) *ArrayElement {
	return &ArrayElement{
		abstractFieldLocationElement: abstractFieldLocationElement{
			kind: arrayElementKind,
		},
		Index: index,
	}
}

func (ae *ArrayElement) String() string {
	return fmt.Sprintf("[%d]", ae.Index)
}
