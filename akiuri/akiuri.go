package akiuri

// There are three kinds of URIs supported here:
//  * akita://serviceName
//      For these, the ObjectType will be nil and the ObjectName will be empty.
//  * akita://serviceName:objectType
//      For these, the ObjectName will be empty.
//  * akita://serviceName:objectType:objectName

import (
	"fmt"
	"strings"
)

const (
	Scheme = "akita://"
)

type ObjectType int

const (
	SPEC  ObjectType = iota
	TRACE            // aka learn session
)

func (o ObjectType) Ptr() *ObjectType {
	return &o
}

// Inspection methods *********************************************************

func (o1 *ObjectType) Is(o2 ObjectType) bool {
	return o1 != nil && *o1 == o2
}

func (o *ObjectType) IsSpec() bool {
	return o.Is(SPEC)
}

func (o *ObjectType) IsTrace() bool {
	return o.Is(TRACE)
}

// ****************************************************************************

func stringToObjectType(s string) (*ObjectType, error) {
	switch s {
	case "spec":
		return SPEC.Ptr(), nil
	case "trace":
		return TRACE.Ptr(), nil
	}
	return nil, fmt.Errorf("%q is an unknown object type", s)
}

func (o ObjectType) String() string {
	switch o {
	case SPEC:
		return "spec"
	case TRACE:
		return "trace"
	default:
		return "unknown"
	}
}

type URI struct {
	ServiceName string
	ObjectName  string
	ObjectType  *ObjectType
}

func (u URI) String() string {
	var sb strings.Builder
	sb.WriteString(Scheme)
	sb.WriteString(u.ServiceName)

	if u.ObjectType != nil {
		sb.WriteString(":")
		sb.WriteString(u.ObjectType.String())

		if u.ObjectName != "" {
			sb.WriteString(":")
			sb.WriteString(u.ObjectName)
		}
	}

	return sb.String()
}

func (u URI) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *URI) UnmarshalText(data []byte) error {
	text := string(data)
	if !strings.HasPrefix(text, Scheme) {
		return fmt.Errorf("%q does not start with %q", text, Scheme)
	}

	parts := strings.Split(text[len(Scheme):], ":")
	if len(parts) > 3 {
		return fmt.Errorf("%q has more than 3 parts", text)
	}

	var objT *ObjectType = nil
	if len(parts) > 1 {
		objType, err := stringToObjectType(parts[1])
		if err != nil {
			return err
		}
		objT = objType
	}

	u.ServiceName = parts[0]
	u.ObjectType = objT
	if len(parts) >= 3 {
		u.ObjectName = parts[2]
	} else {
		u.ObjectName = ""
	}

	return nil
}

func Parse(s string) (URI, error) {
	var u URI
	err := u.UnmarshalText([]byte(s))
	return u, err
}
