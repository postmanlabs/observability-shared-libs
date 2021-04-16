package akiuri

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUnparse(t *testing.T) {
	tests := []struct {
		string
		URI
	}{
		{"akita://my_service:spec:foobar",
			URI{
				ServiceName: "my_service",
				ObjectName:  "foobar",
				ObjectType:  SPEC.Ptr(),
			}},
		{"akita://my_service:trace:foobar",
			URI{
				ServiceName: "my_service",
				ObjectName:  "foobar",
				ObjectType:  TRACE.Ptr(),
			}},
		{"akita://my_service:spec",
			URI{
				ServiceName: "my_service",
				ObjectName:  "",
				ObjectType:  SPEC.Ptr(),
			}},
		{"akita://my_service:trace",
			URI{
				ServiceName: "my_service",
				ObjectName:  "",
				ObjectType:  TRACE.Ptr(),
			}},
		{"akita://my_service",
			URI{
				ServiceName: "my_service",
				ObjectName:  "",
				ObjectType:  nil,
			}},
	}

	for _, test := range tests {
		t.Run("Parse "+test.string, func(t *testing.T) {
			u, err := Parse(test.string)
			assert.NoError(t, err)
			assert.Equal(t, test.URI, u)
		})

		t.Run("Unparse "+test.string, func(t *testing.T) {
			assert.Equal(t, test.string, test.URI.String())
		})
	}
}
