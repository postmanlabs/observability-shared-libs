package http

import (
	"github.com/golang/glog"
	"github.com/google/gopacket/reassembly"

	"github.com/akitasoftware/akita-libs/akinet"
	"github.com/akitasoftware/akita-libs/buffer_pool"
	"github.com/akitasoftware/akita-libs/memview"
)

// Returns a factory for creating HTTP requests whose bodies will be allocated
// from the given buffer pool.
func NewHTTPRequestParserFactory(pool buffer_pool.BufferPool) akinet.TCPParserFactory {
	return httpRequestParserFactory{
		bufferPool: pool,
	}
}

// Returns a factory for creating HTTP responses whose bodies will be allocated
// from the given buffer pool.
func NewHTTPResponseParserFactory(pool buffer_pool.BufferPool) akinet.TCPParserFactory {
	return httpResponseParserFactory{
		bufferPool: pool,
	}
}

type httpRequestParserFactory struct {
	bufferPool buffer_pool.BufferPool
}

func (httpRequestParserFactory) Name() string {
	return "HTTP/1.x Request Parser Factory"
}

func (httpRequestParserFactory) Accepts(input memview.MemView, isEnd bool) (decision akinet.AcceptDecision, discardFront int64) {
	defer func() {
		if decision == akinet.NeedMoreData && isEnd {
			decision = akinet.Reject
			discardFront = input.Len()
		}
	}()

	if input.Len() < minSupportedHTTPMethodLength {
		return akinet.NeedMoreData, 0
	}

	for _, m := range supportedHTTPMethods {
		if start := input.Index(0, []byte(m)); start >= 0 {
			d := hasValidHTTPRequestLine(input.SubView(start+int64(len(m)), input.Len()))
			switch d {
			case akinet.Accept:
				return akinet.Accept, start
			case akinet.NeedMoreData:
				return akinet.NeedMoreData, start
			}
		}
	}
	// Handle the case where the suffix of input is a prefix of the method in a
	// HTTP request line (e.g.  input=`<garbage>GE` where the next input is
	// `T / HTTP/1.1`.
	if input.Len() < maxSupportedHTTPMethodLength {
		return akinet.NeedMoreData, 0
	}
	return akinet.Reject, input.Len()
}

func (f httpRequestParserFactory) CreateParser(id akinet.TCPBidiID, seq, ack reassembly.Sequence) akinet.TCPParser {
	return newHTTPParser(true, id, seq, ack, f.bufferPool)
}

type httpResponseParserFactory struct {
	bufferPool buffer_pool.BufferPool
}

func (httpResponseParserFactory) Name() string {
	return "HTTP/1.x Response Parser Factory"
}

func (httpResponseParserFactory) Accepts(input memview.MemView, isEnd bool) (decision akinet.AcceptDecision, discardFront int64) {
	defer func() {
		if decision == akinet.NeedMoreData && isEnd {
			decision = akinet.Reject
			discardFront = input.Len()
		}
	}()

	if input.Len() < minHTTPResponseStatusLineLength {
		return akinet.NeedMoreData, 0
	}

	for _, v := range []string{"HTTP/1.1", "HTTP/1.0"} {
		if start := input.Index(0, []byte(v)); start >= 0 {
			switch hasValidHTTPResponseStatusLine(input.SubView(start+int64(len(v)), input.Len())) {
			case akinet.Accept:
				return akinet.Accept, start
			case akinet.NeedMoreData:
				return akinet.NeedMoreData, start
			}
		}
	}
	return akinet.Reject, input.Len()
}

func (f httpResponseParserFactory) CreateParser(id akinet.TCPBidiID, seq, ack reassembly.Sequence) akinet.TCPParser {
	return newHTTPParser(false, id, seq, ack, f.bufferPool)
}

// Checks whether there is a valid HTTP request line as defiend in RFC 2616
// Section 5. The input should start right after the HTTP method.
func hasValidHTTPRequestLine(input memview.MemView) akinet.AcceptDecision {
	if input.Len() == 0 {
		return akinet.NeedMoreData
	}

	// A space separates the HTTP method from Request-URI.
	if input.GetByte(0) != ' ' {
		glog.V(6).Info("rejecting HTTP request: lack of space between HTTP method and request-URI")
		return akinet.Reject
	}

	nextSP := input.Index(1, []byte(" "))
	if nextSP < 0 {
		// Could be dealing with a very long request URI.
		if input.Len()-1 > maxHTTPRequestURILength {
			glog.Warning("rejecting potential HTTP request with request URI longer than ", maxHTTPRequestURILength)
			return akinet.Reject
		}
		return akinet.NeedMoreData
	} else if nextSP == 1 {
		glog.V(6).Info("rejecting HTTP request: two spaces after HTTP method")
		return akinet.Reject
	}

	// Need at least 10 bytes to get the HTTP version on tail of the request line,
	// for example `HTTP/1.x\r\n`
	tail := input.SubView(nextSP+1, input.Len())
	if tail.Len() < 10 {
		return akinet.NeedMoreData
	}
	if tail.Index(0, []byte("HTTP/1.1\r\n")) == 0 || tail.Index(0, []byte("HTTP/1.0\r\n")) == 0 {
		return akinet.Accept
	}
	glog.V(6).Info("rejecting HTTP request: request line does not end with HTTP version")
	return akinet.Reject
}

// Checks whether there is a valid HTTP response status line as defiend in
// RFC 2616 Section 6.1. The input should start right after the HTTP version.
func hasValidHTTPResponseStatusLine(input memview.MemView) akinet.AcceptDecision {
	if input.Len() < 5 {
		// Need a space, 3 bytes for status code, and a space or CR.
		return akinet.NeedMoreData
	}

	// A space separates the HTTP version from status code.
	// The format is SP Status-Code SP Reason-Phrase CR LF
	// Some servers do not include the second 'SP" though and skip straight to 'CR'.
	if input.GetByte(0) != ' ' {
		return akinet.Reject
	}
	if input.GetByte(4) != ' ' && input.GetByte(4) != '\r' {
		return akinet.Reject
	}

	// Bytes 1-3 should be in [0-9] for HTTP status code. We don't check that the
	// first digit is in [1-5] to allow custom status codes.
	if !isASCIIDigit(input.GetByte(1)) || !isASCIIDigit(input.GetByte(2)) || !isASCIIDigit(input.GetByte(3)) {
		return akinet.Reject
	}

	if input.Index(0, []byte("\r\n")) < 0 {
		// Could be dealing with a very long reason phrase.
		if input.Len()-4 > maxHTTPReasonPhraseLength {
			glog.Warning("rejecting potential HTTP response with reason phrase longer than ", maxHTTPReasonPhraseLength)
			return akinet.Reject
		}
		return akinet.NeedMoreData
	}

	return akinet.Accept
}

func isASCIIDigit(b byte) bool {
	return '0' <= rune(b) && rune(b) <= '9'
}
