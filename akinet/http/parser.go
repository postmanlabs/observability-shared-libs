package http

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/google/gopacket/reassembly"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/akitasoftware/akita-libs/akinet"
	"github.com/akitasoftware/akita-libs/buffer_pool"
	"github.com/akitasoftware/akita-libs/memview"
)

var (
	// Default maximum HTTP length supported.
	// Can be altered by the CLI as a configuration setting, but doing so after parsing
	// has started will be a race condition.
	MaximumHTTPLength int64 = 1024 * 1024
)

const (
	// Maximum length of a chunk-size line or a trailer line in a chunked body.
	maxChunkLineLength = 4096
)

var (
	lf = []byte{'\n'}

	bufioReaderPool = sync.Pool{
		New: func() any { return bufio.NewReader(nil) },
	}
)

type parseState int

const (
	// Accumulating bytes until the end of the header block.
	stateHeaders parseState = iota

	// Body framed by Content-Length; bodyRemaining bytes left.
	stateFixedBody

	// Body framed by the end of the connection.
	stateUntilClose

	// Chunked body: reading a chunk-size line.
	stateChunkSize

	// Chunked body: reading chunk data; bodyRemaining bytes left in the chunk.
	stateChunkData

	// Chunked body: reading the CRLF that follows chunk data.
	stateChunkDataCRLF

	// Chunked body: reading trailer lines until an empty line.
	stateChunkTrailer
)

// Parses a single HTTP request or response.
//
// The parser is incremental and runs synchronously in the caller's goroutine.
// It buffers input until the header block is complete, uses Go's HTTP library
// to parse the header block, and then frames the body itself (Content-Length,
// chunked, or until the end of the connection) so that it knows exactly where
// the message ends.
type httpParser struct {
	// Indicates whether this parser is for a request or a response.
	isRequest bool

	// Maximum length of HTTP request or response supported; larger requests or
	// responses may be truncated.
	maxHttpLength int64

	pool   buffer_pool.BufferPool
	bidiID akinet.TCPBidiID
	seq    reassembly.Sequence
	ack    reassembly.Sequence

	// The total number of bytes consumed from the stream being parsed.
	totalBytesConsumed int64

	state parseState

	// Input accumulated while in stateHeaders.
	headerBuf memview.MemView

	// Offset in headerBuf from which to resume scanning for the end of the
	// header block.
	headerScanPos int64

	// Partial line accumulated while reading chunk-size or trailer lines.
	lineBuf []byte

	// Remaining bytes in a Content-Length body or in the current chunk.
	bodyRemaining int64

	// Whether the '\r' of the CRLF after chunk data has been seen.
	sawChunkCR bool

	// Set once the header block has been parsed.
	req  *http.Request
	resp *http.Response

	// Holds the body. Allocated once the header block has been parsed.
	//
	// XXX This is used in a very non-local fashion. Consumers of the body are
	// responsible for resetting the buffer, but there is no way to guarantee
	// that this will happen.
	body buffer_pool.Buffer

	// Set once the buffer pool runs out of space; the rest of the body is
	// framed but not stored.
	bodyTruncated bool
}

var _ akinet.TCPParser = (*httpParser)(nil)

func (p *httpParser) Name() string {
	if p.isRequest {
		return "HTTP/1.x Request Parser"
	}
	return "HTTP/1.x Response Parser"
}

// Any bytes in `input` that were consumed and not returned via `unused`
// can be discarded.
func (p *httpParser) Parse(input memview.MemView, isEnd bool) (result akinet.ParsedNetworkContent, unused memview.MemView, totalBytesConsumed int64, err error) {
	consumedBefore := p.totalBytesConsumed
	p.totalBytesConsumed += input.Len()

	done, end, err := p.consume(input)
	if err != nil {
		p.releaseBody()
		return nil, memview.MemView{}, p.totalBytesConsumed, err
	}

	if done {
		unused = input.SubView(end, input.Len())
		p.totalBytesConsumed = consumedBefore + end
		return p.result(), unused, p.totalBytesConsumed, nil
	}

	// If no more input is forthcoming, finish with what we have. This case
	// happens if there's no content-length and we're reading until connection
	// close, or if the body was truncated.
	//
	// Also, if the HTTP request or response is longer than our maximum length,
	// finish anyway. This will leave the input stream in a state where it
	// probably can't find the next header until the accumulated data in the
	// reassembly buffer is all skipped.
	if isEnd || p.totalBytesConsumed > p.maxHttpLength {
		if p.state == stateHeaders {
			p.headerBuf = memview.MemView{}
			return nil, memview.MemView{}, p.totalBytesConsumed, io.ErrUnexpectedEOF
		}
		return p.result(), memview.MemView{}, p.totalBytesConsumed, nil
	}

	return nil, memview.MemView{}, p.totalBytesConsumed, nil
}

// Feeds input to the state machine. Returns done=true and the offset in input
// just past the end of the message once the message is complete.
func (p *httpParser) consume(input memview.MemView) (done bool, end int64, err error) {
	pos := int64(0)

	if p.state == stateHeaders {
		prevLen := p.headerBuf.Len()

		// Scan the accumulated header bytes followed by the input without copying
		// the input; it is only copied if it has to outlive this call.
		headerBuf := p.headerBuf.DeepCopy()
		headerBuf.Append(input)

		headerEnd := p.findHeaderEnd(headerBuf)
		if headerEnd < 0 {
			// Copy the input as the caller assumes its consumed
			p.headerBuf.Append(dup(input))
			return false, 0, nil
		}

		// The end of the header block is always in the latest input, since we
		// would have found it earlier otherwise.
		pos = headerEnd - prevLen

		header := headerBuf.SubView(0, headerEnd)
		p.headerBuf = memview.MemView{}
		if err := p.parseHeader(header); err != nil {
			return false, 0, err
		}
	}

	return p.consumeBody(input, pos)
}

// Scans buf for the empty line that terminates the header block. Returns the
// offset just past the empty line, or -1 if it has not been seen yet.
func (p *httpParser) findHeaderEnd(buf memview.MemView) int64 {
	for {
		nl := buf.Index(p.headerScanPos, lf)
		if nl < 0 {
			p.headerScanPos = buf.Len()
			return -1
		}

		// Look at what follows the line terminator: "\n" or "\r\n" is an empty
		// line.
		if nl+1 >= buf.Len() {
			p.headerScanPos = nl
			return -1
		}
		switch buf.GetByte(nl + 1) {
		case '\n':
			return nl + 2
		case '\r':
			if nl+2 >= buf.Len() {
				p.headerScanPos = nl
				return -1
			}
			if buf.GetByte(nl+2) == '\n' {
				return nl + 3
			}
		}
		p.headerScanPos = nl + 1
	}
}

// Parses a complete header block and determines how the body is framed.
func (p *httpParser) parseHeader(header memview.MemView) error {
	br := bufioReaderPool.Get().(*bufio.Reader)
	br.Reset(header.CreateReader())
	defer func() {
		br.Reset(nil)
		bufioReaderPool.Put(br)
	}()

	var body io.ReadCloser
	var contentLength int64
	var transferEncoding []string
	if p.isRequest {

		req, err := http.ReadRequest(br)

		if err != nil {
			return err
		}
		p.req = req
		body, contentLength, transferEncoding = req.Body, req.ContentLength, req.TransferEncoding
		req.Body = http.NoBody
	} else {
		// XXX BUG Because a nil http.Request is provided to ReadResponse, the http
		// library assumes a GET request. If this is actually a response to a HEAD
		// request and the Content-Length header is present, the parser will treat
		// the bytes after the end of the response as a response body.
		resp, err := http.ReadResponse(br, nil)
		if err != nil {
			return err
		}
		p.resp = resp
		body, contentLength, transferEncoding = resp.Body, resp.ContentLength, resp.TransferEncoding
		resp.Body = http.NoBody
	}

	p.body = p.pool.NewBuffer()

	// Let Go's HTTP library decide whether there is a body (e.g. 204 and 304
	// responses have none), and how it is framed.
	switch {
	case body == nil || body == http.NoBody:
		p.state = stateFixedBody
		p.bodyRemaining = 0
	case len(transferEncoding) > 0:
		// Go only accepts "chunked"; others are rejected by ReadRequest and
		// ReadResponse.
		p.state = stateChunkSize
	case contentLength >= 0:
		p.state = stateFixedBody
		p.bodyRemaining = contentLength
	default:
		p.state = stateUntilClose
	}
	return nil
}

// Consumes body bytes from input starting at pos.
func (p *httpParser) consumeBody(input memview.MemView, pos int64) (done bool, end int64, err error) {
	n := input.Len()
	for {
		switch p.state {
		case stateFixedBody:
			take := min64(p.bodyRemaining, n-pos)
			p.writeBody(input.SubView(pos, pos+take))
			pos += take
			p.bodyRemaining -= take
			if p.bodyRemaining == 0 {
				return true, pos, nil
			}
			return false, 0, nil

		case stateUntilClose:
			p.writeBody(input.SubView(pos, n))
			return false, 0, nil

		case stateChunkSize, stateChunkTrailer:
			line, newPos, ok, err := p.readLine(input, pos)
			if err != nil {
				return false, 0, err
			}
			if !ok {
				return false, 0, nil
			}
			pos = newPos

			if p.state == stateChunkTrailer {
				// Trailers are ignored. An empty line ends the message.
				if line == "" {
					return true, pos, nil
				}
				continue
			}

			if i := strings.IndexByte(line, ';'); i >= 0 {
				// Ignore chunk extensions.
				line = line[:i]
			}
			size, err := strconv.ParseInt(strings.TrimSpace(line), 16, 64)
			if err != nil || size < 0 {
				return false, 0, errors.Errorf("malformed chunk size %q", line)
			}
			if size == 0 {
				p.state = stateChunkTrailer
			} else {
				p.state = stateChunkData
				p.bodyRemaining = size
			}

		case stateChunkData:
			if pos == n {
				return false, 0, nil
			}
			take := min64(p.bodyRemaining, n-pos)
			p.writeBody(input.SubView(pos, pos+take))
			pos += take
			p.bodyRemaining -= take
			if p.bodyRemaining == 0 {
				p.state = stateChunkDataCRLF
				p.sawChunkCR = false
			}

		case stateChunkDataCRLF:
			if pos == n {
				return false, 0, nil
			}
			b := input.GetByte(pos)
			pos++
			switch {
			case b == '\r' && !p.sawChunkCR:
				p.sawChunkCR = true
			case b == '\n':
				p.state = stateChunkSize
			default:
				return false, 0, errors.New("malformed chunked encoding: missing CRLF after chunk data")
			}

		default:
			return false, 0, errors.Errorf("unexpected HTTP parser state %d", p.state)
		}
	}
}

// Reads a line from input starting at pos, combining it with any partial line
// left over from previous input. Returns ok=false if the line is incomplete.
// The returned line excludes the line terminator.
func (p *httpParser) readLine(input memview.MemView, pos int64) (line string, newPos int64, ok bool, err error) {
	nl := input.Index(pos, lf)
	if nl < 0 {
		p.lineBuf = appendFromMemView(p.lineBuf, input.SubView(pos, input.Len()))
		if len(p.lineBuf) > maxChunkLineLength {
			return "", 0, false, errors.New("malformed chunked encoding: line too long")
		}
		return "", input.Len(), false, nil
	}

	p.lineBuf = appendFromMemView(p.lineBuf, input.SubView(pos, nl))
	line = string(bytes.TrimSuffix(p.lineBuf, []byte{'\r'}))
	p.lineBuf = p.lineBuf[0:0]
	return line, nl + 1, true, nil
}

func (p *httpParser) writeBody(data memview.MemView) {
	if p.bodyTruncated || data.Len() == 0 {
		return
	}
	if _, err := p.body.ReadFrom(data.CreateReader()); errors.Is(err, buffer_pool.ErrEmptyPool) {
		// Keep framing the rest of the body, but let the next level handle a body
		// that was truncated.
		p.bodyTruncated = true
	}
}

func (p *httpParser) releaseBody() {
	if p.body != nil {
		p.body.Release()
		p.body = nil
	}
}

func (p *httpParser) result() akinet.ParsedNetworkContent {
	// Because HTTP requires the request to finish before sending a response, TCP
	// ack number on the first segment of the HTTP request is equal to the TCP seq
	// number on the first segment of the corresponding HTTP response. Hence we
	// use it to differntiate differnt pairs of HTTP request and response on the
	// same TCP stream.
	if p.isRequest {
		return akinet.FromStdRequest(uuid.UUID(p.bidiID), int(p.ack), p.req, p.body)
	}
	return akinet.FromStdResponse(uuid.UUID(p.bidiID), int(p.seq), p.resp, p.body)
}

func newHTTPParser(isRequest bool, bidiID akinet.TCPBidiID, seq, ack reassembly.Sequence, pool buffer_pool.BufferPool) *httpParser {
	return &httpParser{
		isRequest:     isRequest,
		maxHttpLength: MaximumHTTPLength,
		pool:          pool,
		bidiID:        bidiID,
		seq:           seq,
		ack:           ack,
		state:         stateHeaders,
		lineBuf:       make([]byte, 0, 64),
	}
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func dup(src memview.MemView) memview.MemView {
	dst := make([]byte, src.Len())
	n, err := src.CreateReader().Read(dst)
	if err != nil || n != len(dst) {
		panic("unexpected MemView dup failed")
	}
	return memview.New(dst)
}

func appendFromMemView(dst []byte, src memview.MemView) []byte {
	pos := len(dst)
	cnt := int(src.Len())
	dst = extend(dst, len(dst)+cnt)
	n, err := src.CreateReader().Read(dst[pos:])
	if err != nil || n != cnt {
		panic("unexpected appendFromMemView failed")
	}
	return dst
}

func extend(what []byte, newSize int) []byte {
	if newSize <= cap(what) {
		return what[:newSize]
	} else {
		res := make([]byte, newSize, 2*newSize)
		copy(res, what)
		return res
	}
}
