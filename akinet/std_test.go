package akinet

import (
	"github.com/akitasoftware/akita-libs/buffer_pool"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestRequestConversion(t *testing.T) {
	testBidiID := TCPBidiID(uuid.MustParse("3744e3d7-2c08-4cd2-9ee9-2306dfba6727"))
	cases := []struct {
		name  string
		input HTTPRequest
	}{
		{
			name: "simple request with cookies",
			input: HTTPRequest{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        1203,
				Method:     "GET",
				ProtoMajor: 1,
				ProtoMinor: 1,
				URL:        &url.URL{Path: "/"},
				Host:       "example.com",
				Cookies: []*http.Cookie{
					{Name: "c1", Value: "1"},
					{Name: "c2", Value: "2"},
				},
				Header: map[string][]string{
					"Cookie": {"c1=1;c2=2"},
				},
			},
		},
	}

	for _, tc := range cases {
		pool, err := buffer_pool.MakeBufferPool(1024, 1024)
		assert.NoError(t, err)
		buffer := pool.NewBuffer()
		_, err = buffer.Write([]byte(tc.input.Body.String()))
		assert.NoError(t, err)
		tc.input.buffer = buffer

		assert.Equal(t, tc.input, FromStdRequest(tc.input.StreamID, tc.input.Seq, tc.input.ToStdRequest(), buffer), tc.name)
	}
}

func TestResponseConversion(t *testing.T) {
	testBidiID := TCPBidiID(uuid.MustParse("3744e3d7-2c08-4cd2-9ee9-2306dfba6727"))
	cases := []struct {
		name  string
		input HTTPResponse
	}{
		{
			name: "simple response with set cookie",
			input: HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				StatusCode: 204,
				ProtoMajor: 1,
				ProtoMinor: 1,
				Cookies: []*http.Cookie{
					{Name: "c1", Value: "1", Raw: "c1=1"},
				},
				Header: map[string][]string{
					"Set-Cookie":  {"c1=1"},
					"X-Akita-Dog": {"prince"},
				},
			},
		},
	}

	for _, tc := range cases {
		pool, err := buffer_pool.MakeBufferPool(1024, 1024)
		assert.NoError(t, err)
		buffer := pool.NewBuffer()
		_, err = buffer.Write([]byte(tc.input.Body.String()))
		assert.NoError(t, err)
		tc.input.buffer = buffer

		assert.Equal(t, tc.input, FromStdResponse(tc.input.StreamID, tc.input.Seq, tc.input.ToStdResponse(), buffer), tc.name)
	}
}
