package http

import (
	"bufio"
	"bytes"
	"compress/flate"
	"errors"
	"fmt"
	"io"
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
	// The message is only complete once the input ends, e.g. a body framed by
	// connection close.
	requiresEnd bool
}

// Runs c through a fresh parser for each way of segmenting its input. If
// signalEnd is set, isEnd is passed with the last segment; otherwise the parser
// must find the end of the message on its own.
func runParseTestCase(isRequest bool, c parseTestCase, pool buffer_pool.BufferPool, signalEnd bool) error {
	return runParseTestCaseWithGarble(isRequest, c, pool, signalEnd, false)
}

// Like runParseTestCase, but if garble is set, each input is given to the
// parser as a private copy whose consumed bytes (those not returned in unused)
// are overwritten once Parse returns. The parser may discard consumed bytes, so
// it must not hold on to them.
func runParseTestCaseWithGarble(isRequest bool, c parseTestCase, pool buffer_pool.BufferPool, signalEnd bool, garble bool) error {
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

			var buf []byte
			if garble {
				buf = []byte(input.String())
				input = memview.New(buf)
			}

			pnc, unused, totalBytesConsumed, err = p.Parse(input, signalEnd && i == len(inputs)-1)

			if garble {
				consumed := buf
				if err == nil {
					consumed = buf[:int64(len(buf))-unused.Len()]
				}
				for k := range consumed {
					consumed[k] = 0xff
				}
			}

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

func requestParseTestCases() []parseTestCase {
	return []parseTestCase{
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
}

func TestHTTPRequestParser(t *testing.T) {
	pool, err := buffer_pool.MakeBufferPool(1024*1024, 4*1024)
	if err != nil {
		t.Error(err)
	}

	for _, c := range requestParseTestCases() {
		t.Logf(c.name)
		if err := runParseTestCase(true, c, pool, true); err != nil {
			t.Error(err)
		}
	}
}

func responseParseTestCases() []parseTestCase {
	return []parseTestCase{
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
			name:        "frame by connection close",
			input:       "HTTP/1.0 200 OK\r\n\r\nhello this is prince",
			requiresEnd: true,
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
			// No Content-Length, so the body is framed by connection close.
			requiresEnd: true,
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
		{
			name:  "chunked body with extension and trailer",
			input: "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n5;foo=bar\r\nhello\r\n6\r\n world\r\n0\r\nX-Trailer: t\r\n\r\n",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				Body:       memview.New([]byte("hello world")),
			},
		},
		{
			name:      "malformed chunk size",
			input:     "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\nzz\r\nhello\r\n0\r\n\r\n",
			expectErr: true,
		},
		{
			name: "pipelined responses",
			verbatimInput: []memview.MemView{
				memview.New([]byte("HTTP/1.1 200 OK\r\nContent-Length: 3\r\n\r\nfooHTTP/1.1 204 No Content\r\n\r\n")),
			},
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				Header:     map[string][]string{"Content-Length": {"3"}},
				Body:       memview.New([]byte("foo")),
			},
			bytesRemaining: int64(len("HTTP/1.1 204 No Content\r\n\r\n")),
		},
		{
			name: "304 has no body despite content-length",
			verbatimInput: []memview.MemView{
				memview.New([]byte("HTTP/1.1 304 Not Modified\r\nContent-Length: 5\r\n\r\nHTTP/1.1")),
			},
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 304,
				Header:     map[string][]string{"Content-Length": {"5"}},
			},
			bytesRemaining: int64(len("HTTP/1.1")),
		},
	}
}

func TestHTTPResponseParser(t *testing.T) {
	pool, err := buffer_pool.MakeBufferPool(1024*1024, 4*1024)
	if err != nil {
		t.Error(err)
	}

	for _, c := range responseParseTestCases() {
		if err := runParseTestCase(false, c, pool, true); err != nil {
			t.Error(err)
		}
	}
}

// Every message except one framed by connection close must be recognized as
// complete as soon as its last byte arrives. Otherwise the parser would only
// finish when the connection ends, and the isEnd fallback in Parse would hide
// that in TestHTTPRequestParser and TestHTTPResponseParser.
func TestParseCompletesWithoutEnd(t *testing.T) {
	pool, err := buffer_pool.MakeBufferPool(1024*1024, 4*1024)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		isRequest bool
		cases     []parseTestCase
	}{
		{true, requestParseTestCases()},
		{false, responseParseTestCases()},
	} {
		for _, c := range tc.cases {
			if c.requiresEnd {
				continue
			}
			if err := runParseTestCase(tc.isRequest, c, pool, false); err != nil {
				t.Error(err)
			}
		}
	}
}

// The bytes that Parse consumes may be discarded, and so overwritten, as soon
// as it returns, so the parser must copy whatever it keeps of them, e.g. a
// partial header block or chunk-size line.
func TestParsedInputCanBeGarbled(t *testing.T) {
	pool, err := buffer_pool.MakeBufferPool(1024*1024, 4*1024)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		isRequest bool
		cases     []parseTestCase
	}{
		{true, requestParseTestCases()},
		{false, responseParseTestCases()},
	} {
		for _, c := range tc.cases {
			if err := runParseTestCaseWithGarble(tc.isRequest, c, pool, true, true); err != nil {
				t.Error(err)
			}
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

var getNoBodyRequest = []byte(strings.Join([]string{
	"GET /test/path HTTP/1.1\r\n",
	"Host: 172.17.0.2:8080\r\n",
	"Accept: */*\r\n",
	"Accept-Encoding: *\r\n",
	"User-Agent: Mozilla/5.0 (unknown-aarch64-linux-gnu) Siege/4.2.2\r\n",
	"Connection: close\r\n",
	"\r\n",
}, ""))

var Scratch = make([]byte, 64*1024*1024)

func BenchmarkHTTPRequestParserGETNoBody(b *testing.B) {
	pool, err := buffer_pool.MakeBufferPool(1024*1024, 4*1024)
	if err != nil {
		b.Fatal(err)
	}
	/*
		inputs := [][]byte{}
		for i := 0; i < b.N; i++ {
			input := append([]byte{}, getNoBodyRequest...)
			inputs = append(inputs, input)
		}

		flush := func() {
			// flush cache
			for i := 0; i < len(Scratch); i++ {
				Scratch[i]++
			}
			_ = Scratch
		}

		//flush()
		_ = flush
	*/
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := newHTTPParser(true, testBidiID, 522, 1203, pool)
		pnc, _, _, err := p.Parse(memview.New(getNoBodyRequest), false)
		if err != nil {
			b.Fatal(err)
		}
		if pnc == nil {
			b.Fatal("parsing incomplete")
		}
		pnc.ReleaseBuffers()

		b.StopTimer()
		//flush()
		b.StartTimer()
	}
}

// Benchmarks only Go's http.ReadRequest on the same input, as a floor for
// BenchmarkHTTPRequestParserGETNoBody. The bufio.Reader is reused across
// iterations, as httpParser does.
func BenchmarkStdReadRequestGETNoBody(b *testing.B) {
	input := getNoBodyRequest
	r := bytes.NewReader(input)
	br := bufio.NewReader(r)

	b.SetBytes(int64(len(input)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Reset(input)
		br.Reset(r)
		if _, err := http.ReadRequest(br); err != nil {
			b.Fatal(err)
		}
	}
}

// Reports whether n bytes can still be taken from pool.
func poolHasRoom(pool buffer_pool.BufferPool, n int) bool {
	probe := pool.NewBuffer()
	defer probe.Release()
	_, err := probe.Write(make([]byte, n))
	return err == nil
}

// When the buffer pool runs out, the body is truncated but still framed, so
// the message ends in the right place and bytes after it are left unused.
func TestBodyTruncatedWhenPoolExhausted(t *testing.T) {
	const poolSize = 16
	pool, err := buffer_pool.MakeBufferPool(poolSize, poolSize)
	if err != nil {
		t.Fatal(err)
	}

	body := strings.Repeat("x", 4*poolSize)
	msg := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
	trailer := "HTTP/1.1 204 No Content\r\n\r\n"

	p := newHTTPParser(false, testBidiID, 522, 1203, pool)
	pnc, unused, consumed, err := p.Parse(memview.New([]byte(msg+trailer)), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pnc == nil {
		t.Fatal("parsing incomplete")
	}
	defer pnc.ReleaseBuffers()

	resp := pnc.(akinet.HTTPResponse)
	if got := resp.Body.Len(); got == 0 || got >= int64(len(body)) {
		t.Errorf("body length = %d, want truncated to between 1 and %d", got, len(body)-1)
	}
	if got := unused.String(); got != trailer {
		t.Errorf("unused = %q, want %q", got, trailer)
	}
	if consumed != int64(len(msg)) {
		t.Errorf("consumed %d bytes, want %d", consumed, len(msg))
	}
}

// A parse error gives the partially filled body buffer back to the pool.
func TestParseErrorReleasesBody(t *testing.T) {
	const chunkSize = 16
	const poolSize = 2 * chunkSize
	pool, err := buffer_pool.MakeBufferPool(poolSize, chunkSize)
	if err != nil {
		t.Fatal(err)
	}

	p := newHTTPParser(false, testBidiID, 522, 1203, pool)
	start := fmt.Sprintf("HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n%x\r\n%s\r\n", chunkSize, strings.Repeat("x", chunkSize))
	if pnc, _, _, err := p.Parse(memview.New([]byte(start)), false); pnc != nil || err != nil {
		t.Fatalf("expected incomplete parse, got pnc=%v err=%v", pnc, err)
	}
	if poolHasRoom(pool, poolSize) {
		t.Fatal("body should be holding a chunk of the pool")
	}

	if _, _, _, err := p.Parse(memview.New([]byte("zz\r\n")), false); err == nil {
		t.Fatal("expected error for malformed chunk size")
	}
	if !poolHasRoom(pool, poolSize) {
		t.Error("parse error did not return the body's buffers to the pool")
	}
}

// Input that ends before the header block is complete is an error, and all of
// it counts as consumed.
func TestHeadersIncompleteAtEnd(t *testing.T) {
	pool, err := buffer_pool.MakeBufferPool(1024*1024, 4*1024)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		isRequest bool
		inputs    []string
	}{
		{true, []string{"GET / HTTP/1.1\r\n", "Host: exa"}},
		{true, []string{"GET / HTTP/1.1\r\nHost: example.com\r\n"}},
		{false, []string{"HTTP/1.1 200 OK\r\n", "Content-Length: 3\r\n\r"}},
	} {
		p := newHTTPParser(c.isRequest, testBidiID, 522, 1203, pool)
		var pnc akinet.ParsedNetworkContent
		var consumed int64
		inputSize := int64(0)
		for i, in := range c.inputs {
			inputSize += int64(len(in))
			pnc, _, consumed, err = p.Parse(memview.New([]byte(in)), i == len(c.inputs)-1)
		}
		if pnc != nil {
			t.Errorf("%q: expected no result, got %v", c.inputs, pnc)
		}
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("%q: err = %v, want %v", c.inputs, err, io.ErrUnexpectedEOF)
		}
		if consumed != inputSize {
			t.Errorf("%q: consumed %d bytes, want %d", c.inputs, consumed, inputSize)
		}
	}
}

// Input that ends partway through the body yields the message with as much of
// the body as arrived.
func TestBodyIncompleteAtEnd(t *testing.T) {
	pool, err := buffer_pool.MakeBufferPool(1024*1024, 4*1024)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []parseTestCase{
		{
			name:  "content-length body cut short",
			input: "HTTP/1.1 200 OK\r\nContent-Length: 10\r\n\r\nabcd",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				Header:     map[string][]string{"Content-Length": {"10"}},
				Body:       memview.New([]byte("abcd")),
			},
		},
		{
			name:  "chunked body cut in chunk data",
			input: "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nhel",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				Body:       memview.New([]byte("hel")),
			},
		},
		{
			name:  "chunked body cut in chunk-size line",
			input: "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nhello\r\n1",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				Body:       memview.New([]byte("hello")),
			},
		},
	}

	for _, c := range testCases {
		if err := runParseTestCase(false, c, pool, true); err != nil {
			t.Error(err)
		}
	}
}

// Bytes following a message are left unused, wherever the message itself is
// split. The split points stay inside the message so that all of the trailing
// bytes arrive with its last byte.
func TestTrailingBytesAfterMessage(t *testing.T) {
	pool, err := buffer_pool.MakeBufferPool(1024*1024, 4*1024)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name      string
		isRequest bool
		msg       string
		trailer   string
		expected  akinet.ParsedNetworkContent
	}{
		{
			name:      "after chunked response",
			isRequest: false,
			msg:       "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nhello\r\n6\r\n world\r\n0\r\nX-Trailer: t\r\n\r\n",
			trailer:   "HTTP/1.1 204 No Content\r\n\r\n",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				Body:       memview.New([]byte("hello world")),
			},
		},
		{
			name:      "after chunked request",
			isRequest: true,
			msg:       "POST /a HTTP/1.1\r\nHost: example.com\r\nTransfer-Encoding: chunked\r\n\r\n3\r\nfoo\r\n0\r\n\r\n",
			trailer:   "GET /b HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expected: akinet.HTTPRequest{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        1203,
				Method:     "POST",
				ProtoMajor: 1,
				ProtoMinor: 1,
				URL:        &url.URL{Path: "/a"},
				Host:       "example.com",
				Body:       memview.New([]byte("foo")),
			},
		},
		{
			name:      "pipelined requests",
			isRequest: true,
			msg:       "GET /a HTTP/1.1\r\nHost: example.com\r\n\r\n",
			trailer:   "GET /b HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expected: akinet.HTTPRequest{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        1203,
				Method:     "GET",
				ProtoMajor: 1,
				ProtoMinor: 1,
				URL:        &url.URL{Path: "/a"},
				Host:       "example.com",
			},
		},
	} {
		full := c.msg + c.trailer
		for i := 0; i < len(c.msg); i++ {
			tc := parseTestCase{
				name: fmt.Sprintf("%s, split at %d", c.name, i),
				verbatimInput: []memview.MemView{
					memview.New([]byte(full[:i])),
					memview.New([]byte(full[i:])),
				},
				expected:       c.expected,
				bytesRemaining: int64(len(c.trailer)),
			}
			if err := runParseTestCase(c.isRequest, tc, pool, false); err != nil {
				t.Error(err)
			}
		}
	}
}

func TestMalformedChunkedBody(t *testing.T) {
	pool, err := buffer_pool.MakeBufferPool(1024*1024, 4*1024)
	if err != nil {
		t.Fatal(err)
	}

	const header = "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n"
	testCases := []parseTestCase{
		{
			name:      "missing CRLF after chunk data",
			input:     header + "5\r\nhelloX\r\n0\r\n\r\n",
			expectErr: true,
		},
		{
			name:      "negative chunk size",
			input:     header + "-1\r\nhello\r\n0\r\n\r\n",
			expectErr: true,
		},
		{
			name:      "chunk size overflows int64",
			input:     header + "ffffffffffffffffff\r\nhello\r\n0\r\n\r\n",
			expectErr: true,
		},
		{
			// Too long to segment every possible way. A valid chunk size padded
			// with leading zeros, so only its length makes it malformed. Its line
			// terminator arrives in a later input, so the length check on a
			// partial line is what fails.
			name: "chunk-size line too long",
			verbatimInput: []memview.MemView{
				memview.New([]byte(header)),
				memview.New([]byte(strings.Repeat("0", maxChunkLineLength) + "5")),
				memview.New([]byte("\r\nhello\r\n0\r\n\r\n")),
			},
			expectErr: true,
		},
	}

	for _, c := range testCases {
		if err := runParseTestCase(false, c, pool, true); err != nil {
			t.Error(err)
		}
	}
}

// Header blocks whose lines end in a bare LF instead of CRLF, which Go's HTTP
// library accepts and some clients send.
func TestHeaderBlockEndingInBareLF(t *testing.T) {
	pool, err := buffer_pool.MakeBufferPool(1024*1024, 4*1024)
	if err != nil {
		t.Fatal(err)
	}

	requestCases := []parseTestCase{
		{
			name:  "request with LF line endings",
			input: "GET / HTTP/1.1\nHost: example.com\n\n",
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
			name:  "request with CRLF lines and a bare LF empty line",
			input: "GET / HTTP/1.1\r\nHost: example.com\r\n\n",
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
	}
	responseCases := []parseTestCase{
		{
			name:  "response with LF line endings",
			input: "HTTP/1.1 200 OK\nContent-Length: 3\n\nfoo",
			expected: akinet.HTTPResponse{
				StreamID:   uuid.UUID(testBidiID),
				Seq:        522,
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				Header:     map[string][]string{"Content-Length": {"3"}},
				Body:       memview.New([]byte("foo")),
			},
		},
	}

	for _, c := range requestCases {
		if err := runParseTestCase(true, c, pool, false); err != nil {
			t.Error(err)
		}
	}
	for _, c := range responseCases {
		if err := runParseTestCase(false, c, pool, false); err != nil {
			t.Error(err)
		}
	}
}
