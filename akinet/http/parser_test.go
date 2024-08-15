package http

import (
	"bytes"
	"compress/flate"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"github.com/akitasoftware/akita-libs/akinet"
	"github.com/akitasoftware/akita-libs/buffer_pool"
	"github.com/akitasoftware/akita-libs/memview"
)

var (
	testBidiID   = akinet.TCPBidiID(uuid.MustParse("3744e3d7-2c08-4cd2-9ee9-2306dfba6727"))
	chunkedBody  bytes.Buffer
	deflatedBody bytes.Buffer

	multipartFormData = strings.Join([]string{
		"--b9580db\r\n",
		"Content-Disposition: form-data; name=\"field1\"\r\n",
		"\r\n",
		"value1\r\n",
		"--b9580db\r\n",
		"Content-Disposition: form-data; name=\"field2\"\r\n",
		"Content-Type: application/json\r\n",
		"\r\n",
		`{"foo": "bar", "baz": 123}` + "\r\n",
		"--b9580db--",
	}, "")
)

func init() {
	cw := httputil.NewChunkedWriter(&chunkedBody)
	cw.Write([]byte("hello "))
	cw.Write([]byte("thi"))
	cw.Write([]byte("s is chunk"))
	cw.Write([]byte("ed body"))
	cw.Close()
	// Must manually write the last CRLF after tailers.
	chunkedBody.Write([]byte("\r\n"))

	dw, err := flate.NewWriter(&deflatedBody, flate.BestCompression)
	if err != nil {
		panic(err)
	}
	dw.Write([]byte("hello this is deflated body"))
	dw.Close()
}

type parseTestCase struct {
	name string
	// input will get segmented in O(n^2) different ways to test robustness. Use
	// verbatimInput instead of large inputs.
	input string
	// verbatimInput will not get segmented.
	verbatimInput  []memview.MemView
	expected       akinet.ParsedNetworkContent
	expectErr      bool
	bytesRemaining int64 // num bytes from inputs expected to be left unconsumed
}

func runParseTestCase(isRequest bool, c parseTestCase, pool buffer_pool.BufferPool) error {
	var segments <-chan []memview.MemView
	if c.verbatimInput != nil {
		s := make(chan []memview.MemView)
		segments = s
		go func() {
			s <- c.verbatimInput
			close(s)
		}()
	} else {
		segments = segment3(c.input)
	}

	for inputs := range segments {
		var pnc akinet.ParsedNetworkContent
		var unused memview.MemView
		var totalBytesConsumed int64
		var err error

		inputSize := int64(0)
		p := newHTTPParser(isRequest, testBidiID, 522, 1203, pool)
		for i, input := range inputs {
			inputSize += input.Len()

			pnc, unused, totalBytesConsumed, err = p.Parse(input, i == len(inputs)-1)
			if err != nil {
				break
			} else if pnc != nil {
				break
			}
		}

		if pnc != nil {
			if c.expectErr {
				return fmt.Errorf("[%s] expected error, got none input=%s", c.name, dump(inputs))
			} else {
				if diff := cmp.Diff(c.expected, pnc, cmpopts.EquateEmpty(), cmpopts.IgnoreUnexported(akinet.HTTPRequest{}, akinet.HTTPResponse{})); diff != "" {
					return fmt.Errorf("[%s] found diff: %s input=%s", c.name, diff, dump(inputs))
				}
				if unused.Len() != c.bytesRemaining {
					return fmt.Errorf("[%s] expected %d bytes remaining, got %d input=%s", c.name, c.bytesRemaining, unused.Len(), dump(inputs))
				}
			}

			pnc.ReleaseBuffers()
		} else if err != nil {
			if !c.expectErr {
				return fmt.Errorf("[%s] expected no error, got: %v input=%s", c.name, err, dump(inputs))
			}
		} else {
			return fmt.Errorf("[%s] parsing incomplete input=%s", c.name, dump(inputs))
		}

		expectedTotalBytesConsumed := inputSize
		if err == nil {
			expectedTotalBytesConsumed -= c.bytesRemaining
		}
		if totalBytesConsumed != expectedTotalBytesConsumed {
			return fmt.Errorf("[%s] expected %d bytes consumed, got %d input=%s", c.name, expectedTotalBytesConsumed, totalBytesConsumed, dump(inputs))
		}
	}
	return nil
}

func TestHTTPRequestParser(t *testing.T) {
	pool, err := buffer_pool.MakeBufferPool(1024*1024, 4*1024)
	if err != nil {
		t.Error(err)
	}

	testCases := []parseTestCase{
		{
			name:  "request line only",
			input: "GET / HTTP/1.0\r\n\r\n",
			expected: akinet.HTTPRequest{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        1203,
				Method:     "GET",
				ProtoMajor: 1,
				ProtoMinor: 0,
				URL:        &url.URL{Path: "/"},
			},
		},
		{
			name:      "bad header",
			input:     "GET / HTTP/1.1\r\nHost: \r\nexample.com\r\n\r\n",
			expectErr: true,
		},
		{
			name:  "simple request without body",
			input: "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expected: akinet.HTTPRequest{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        1203,
				Method:     "GET",
				ProtoMajor: 1,
				ProtoMinor: 1,
				URL:        &url.URL{Path: "/"},
				Host:       "example.com",
			},
		},
		{
			name:  "simple request with body",
			input: "POST /foo HTTP/1.1\r\nHost: example.com\r\nContent-Length: 9\r\n\r\nfoobarbaz",
			expected: akinet.HTTPRequest{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        1203,
				Method:     "POST",
				ProtoMajor: 1,
				ProtoMinor: 1,
				URL:        &url.URL{Path: "/foo"},
				Host:       "example.com",
				Header:     map[string][]string{"Content-Length": {"9"}},
				Body:       memview.New([]byte("foobarbaz")),
			},
		},
		{
			name: "ignore trailing bytes",
			verbatimInput: []memview.MemView{
				memview.New([]byte("POST /foo HTTP/1.1\r\n")),
				memview.New([]byte("Host: example.com\r\nContent-Length: 9\r\n\r\n")),
				memview.New([]byte("foobarbaz thisshouldnotshowup")),
			},
			expected: akinet.HTTPRequest{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        1203,
				Method:     "POST",
				ProtoMajor: 1,
				ProtoMinor: 1,
				URL:        &url.URL{Path: "/foo"},
				Host:       "example.com",
				Header:     map[string][]string{"Content-Length": {"9"}},
				Body:       memview.New([]byte("foobarbaz")),
			},
			bytesRemaining: int64(len(" thisshouldnotshowup")),
		},
		{
			name: "ignore CRLF in body",
			input: strings.Join([]string{
				"POST /foo HTTP/1.1\r\n",
				"Host: example.com\r\n",
				"Content-Length: 11\r\n\r\nfoobar\r\nbaz",
			}, ""),
			expected: akinet.HTTPRequest{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        1203,
				Method:     "POST",
				ProtoMajor: 1,
				ProtoMinor: 1,
				URL:        &url.URL{Path: "/foo"},
				Host:       "example.com",
				Header:     map[string][]string{"Content-Length": {"11"}},
				Body:       memview.New([]byte("foobar\r\nbaz")),
			},
		},
		{
			name: "chunked body",
			input: strings.Join([]string{
				"GET / HTTP/1.1\r\n",
				"Host: example.com\r\n",
				"Transfer-Encoding: chunked\r\n",
				"\r\n",
				chunkedBody.String(),
			}, ""),
			expected: akinet.HTTPRequest{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        1203,
				Method:     "GET",
				ProtoMajor: 1,
				ProtoMinor: 1,
				URL:        &url.URL{Path: "/"},
				Host:       "example.com",
				Body:       memview.New([]byte("hello this is chunked body")),
			},
		},
		{
			name:  "content-length 0",
			input: "POST / HTTP/1.1\r\nContent-Length: 0\r\n\r\n",
			expected: akinet.HTTPRequest{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        1203,
				Method:     "POST",
				ProtoMajor: 1,
				ProtoMinor: 1,
				URL:        &url.URL{Path: "/"},
				Header:     map[string][]string{"Content-Length": {"0"}},
			},
		},
		{
			name: "multipart/form-data",
			input: strings.Join([]string{
				"POST / HTTP/1.1\r\n",
				fmt.Sprintf("Content-Length: %d\r\n", len(multipartFormData)),
				"Content-Type: multipart/form-data;boundary=b9580db\r\n",
				"\r\n",
				multipartFormData,
			}, ""),
			expected: akinet.HTTPRequest{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        1203,
				Method:     "POST",
				ProtoMajor: 1,
				ProtoMinor: 1,
				URL:        &url.URL{Path: "/"},
				Header: map[string][]string{
					"Content-Type":   {"multipart/form-data;boundary=b9580db"},
					"Content-Length": {strconv.Itoa(len(multipartFormData))},
				},
				Body: memview.New([]byte(multipartFormData)),
			},
		},
		{
			name:  "simple request with cookies",
			input: "GET / HTTP/1.1\r\nHost: example.com\r\nCookie: c1=1;c2=2\r\n\r\n",
			expected: akinet.HTTPRequest{
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

	for _, c := range testCases {
		t.Logf(c.name)
		if err := runParseTestCase(true, c, pool); err != nil {
			t.Error(err)
		}
	}
}

func TestHTTPResponseParser(t *testing.T) {
	pool, err := buffer_pool.MakeBufferPool(1024*1024, 4*1024)
	if err != nil {
		t.Error(err)
	}

	testCases := []parseTestCase{
		{
			name:  "status line only",
			input: "HTTP/1.0 204 No Content\r\n\r\n",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 0,
				StatusCode: 204,
			},
		},
		{
			name:      "bad header",
			input:     "HTTP/1.1 204 No Content\r\nX-Akita-Dog: \r\nprince\r\n\r\n",
			expectErr: true,
		},
		{
			name:  "simple response without body",
			input: "HTTP/1.1 204 No Content\r\nX-Akita-Dog: prince\r\n\r\n",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				StatusCode: 204,
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header:     map[string][]string{"X-Akita-Dog": {"prince"}},
			},
		},
		{
			name:  "simple response with body",
			input: "HTTP/1.1 200 OK\r\nX-Akita-Dog: prince\r\nContent-Length: 9\r\n\r\nfoobarbaz",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				StatusCode: 200,
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header: map[string][]string{
					"X-Akita-Dog":    {"prince"},
					"Content-Length": {"9"},
				},
				Body: memview.New([]byte("foobarbaz")),
			},
		},
		{
			name: "example response from Postman health check missing reason phrase",
			input: "HTTP/1.1 200\r\nServer: nginx/1.25.3\r\nDate: Fri, 02 Aug 2024 03:04:14 GMT\r\n" +
				"Content-Type: application/json; charset=utf-8\r\nContent-Length: 28\r\n" +
				"Connection: close\r\nuWebSockets: 20\r\n\r\n{\"response\":\"Who is there?\"}",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				StatusCode: 200,
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header: map[string][]string{
					"Server": {"nginx/1.25.3"},
					// "Connection":     {"close"},    // ignored by parser
					"Date":           {"Fri, 02 Aug 2024 03:04:14 GMT"},
					"Uwebsockets":    {"20"}, // normalized
					"Content-Length": {"28"},
					"Content-Type":   {"application/json; charset=utf-8"},
				},
				Body: memview.New([]byte("{\"response\":\"Who is there?\"}")),
			},
		},
		{
			name: "ignore trailing bytes",
			verbatimInput: []memview.MemView{
				memview.New([]byte("HTTP/1.1 200 OK\r\n")),
				memview.New([]byte("X-Akita-Dog: prince\r\nContent-Length: 9\r\n\r\n")),
				memview.New([]byte("foobarbaz thisshouldnotshowup")),
			},
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				StatusCode: 200,
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header: map[string][]string{
					"X-Akita-Dog":    {"prince"},
					"Content-Length": {"9"},
				},
				Body: memview.New([]byte("foobarbaz")),
			},
			bytesRemaining: int64(len(" thisshouldnotshowup")),
		},
		{
			name:  "ignore CRLF in body",
			input: "HTTP/1.1 200 OK\r\nContent-Length: 11\r\n\r\nfoobar\r\nbaz",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				Header:     map[string][]string{"Content-Length": {"11"}},
				Body:       memview.New([]byte("foobar\r\nbaz")),
			},
		},
		{
			name: "chunked body",
			input: strings.Join([]string{
				"HTTP/1.1 200 OK\r\n",
				"Transfer-Encoding: chunked\r\n",
				"\r\n",
				chunkedBody.String(),
			}, ""),
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				Body:       memview.New([]byte("hello this is chunked body")),
			},
		},
		{
			name:  "content-length 0",
			input: "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				Header:     map[string][]string{"Content-Length": {"0"}},
			},
		},
		{
			// No Content-Length, we need to read the body until input has ended.
			name:  "frame by connection close",
			input: "HTTP/1.0 200 OK\r\n\r\nhello this is prince",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 0,
				StatusCode: 200,
				Body:       memview.New([]byte("hello this is prince")),
			},
		},
		{
			name: "content-encoding unchanged",
			input: strings.Join([]string{
				"HTTP/1.1 200 OK\r\n",
				"Content-Encoding: deflate\r\n",
				"\r\n",
				deflatedBody.String(),
			}, ""),
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				// We expect the body to not get inflated because this library does not
				// handle content-encoding.
				Header: map[string][]string{"Content-Encoding": {"deflate"}},
				Body:   memview.New(deflatedBody.Bytes()),
			},
		},
		{
			// Unfortunately, go's HTTP reader returns an error when handling
			// non-chunked transfer-encoding. This is a limitation that is not easy to
			// fix short of writing our own HTTP parser.
			name: "non-chunked transfer-encoding not handled",
			input: strings.Join([]string{
				"HTTP/1.1 200 OK\r\n",
				"Transfer-Encoding: deflate\r\n",
				"\r\n",
				deflatedBody.String(),
			}, ""),
			expectErr: true,
		},
		{
			name: "multipart/form-data",
			input: strings.Join([]string{
				"HTTP/1.1 200 OK\r\n",
				fmt.Sprintf("Content-Length: %d\r\n", len(multipartFormData)),
				"Content-Type: multipart/form-data;boundary=b9580db\r\n",
				"\r\n",
				multipartFormData,
			}, ""),
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				Header: map[string][]string{
					"Content-Type":   {"multipart/form-data;boundary=b9580db"},
					"Content-Length": {strconv.Itoa(len(multipartFormData))},
				},
				Body: memview.New([]byte(multipartFormData)),
			},
		},
		{
			name:  "simple response with set cookie",
			input: "HTTP/1.1 204 No Content\r\nX-Akita-Dog: prince\r\nSet-Cookie: c1=1\r\n\r\n",
			expected: akinet.HTTPResponse{
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

	for _, c := range testCases {
		if err := runParseTestCase(false, c, pool); err != nil {
			t.Error(err)
		}
	}
}

func TestOversizedResponse(t *testing.T) {
	pool, err := buffer_pool.MakeBufferPool(10*1024*1024, 4*1024)
	if err != nil {
		t.Error(err)
	}

	packets := []string{
		"HTTP/1.1 200 OK\r\nHost: example.com\r\n",
		"Content-Length: 2000000\r\n\r\n",
	}
	bigPayload := make([]byte, 2000000)

	inputs := []memview.MemView{
		memview.New([]byte(packets[0])),
		memview.New([]byte(packets[1])),
		memview.New(bigPayload[:500000]),
		memview.New(bigPayload[500000:1000000]),
		memview.New(bigPayload[1000000:1200000]),
		memview.New(bigPayload[1200000:1400000]),
		memview.New(bigPayload[1400000:1600000]),
		memview.New(bigPayload[1600000:1800000]),
		memview.New(bigPayload[1800000:2000000]),
	}

	p := newHTTPParser(false, testBidiID, 522, 1203, pool)
	var pnc akinet.ParsedNetworkContent
	var unused memview.MemView
	var totalBytesConsumed int64

	expectedBytesConsumed := int64(0)

	for _, i := range inputs {
		expectedBytesConsumed += i.Len()

		pnc, unused, totalBytesConsumed, err = p.Parse(i, false)
		if err != nil {
			t.Fatalf("Got error: %v", err)
		}
		if pnc != nil {
			break
		}
	}

	defer pnc.ReleaseBuffers()

	if pnc == nil {
		t.Fatalf("didn't parse a packet")
	}
	if unused.Len() != 0 {
		t.Errorf("unused length is %v", unused.Len())
	}

	response := pnc.(akinet.HTTPResponse)
	// First payload larger than http limit
	if response.Body.Len() > 1200000 || response.Body.Len() < 1000000 {
		t.Errorf("got packet with body length %v", response.Body.Len())
	}

	if expectedBytesConsumed != totalBytesConsumed {
		t.Errorf("expected %d bytes consumed, but actually consumed %d bytes", expectedBytesConsumed, totalBytesConsumed)
	}
}
