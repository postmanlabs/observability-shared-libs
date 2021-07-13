package http

import (
	"bufio"
	"bytes"
	"io"
	"net/http"

	"github.com/google/gopacket/reassembly"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/akitasoftware/akita-libs/akinet"
	"github.com/akitasoftware/akita-libs/memview"
)

var (
	// Default maximum HTTP length supported.
	// Can be altered by the CLI as a configuration setting, but doing so after parsing
	// has started will be a race condition.
	MaximumHTTPLength int64 = 1024 * 1024
)

// Implements TCPParser
type httpParser struct {
	w *io.PipeWriter

	allInput memview.MemView

	// Signal that read side of the pipe has closed.
	readClosed chan error

	resultChan chan akinet.ParsedNetworkContent
	isRequest  bool

	// Maximum length of HTTP protocol unit supported; larger requests
	// or responses may be truncated.
	maxHttpLength int64
}

func (p *httpParser) Name() string {
	if p.isRequest {
		return "HTTP/1.x Request Parser"
	}
	return "HTTP/1.x Response Parser"
}

func (p *httpParser) Parse(input memview.MemView, isEnd bool) (result akinet.ParsedNetworkContent, unused memview.MemView, err error) {
	var consumedBytes int64
	defer func() {
		if err == nil {
			return
		}

		// Adjust the number of bytes that were read by the reader but were unused.
		switch e := err.(type) {
		case httpPipeReaderDone:
			result = <-p.resultChan
			unused = input.SubView(consumedBytes-int64(e), input.Len())
			err = nil
		case httpPipeReaderError:
			unused = p.allInput
			err = e.err
		default:
			err = errors.Wrap(err, "encountered unknown HTTP pipe reader error")
		}
	}()

	p.allInput.Append(input)

	// The PipeWriter blocks until the reader is done consuming all the bytes.
	consumedBytes, err = io.Copy(p.w, input.CreateReader())
	if err != nil {
		return
	}

	// The reader might close (aka parse complete) after the write returns, so we
	// need to check. We force an empty write such that:
	// - If the parse is indeed complete, the reader no longer consumes anything,
	// 	 so this call will block until the reader closes.
	// - If the parse is not done yet, the empty write doesn't change things.
	_, err = p.w.Write([]byte{})
	if err != nil {
		return
	}

	// If the reader has not closed yet, tell it we have no more input. This case
	// happens if there's no content-length and we're reading until connection
	// close.
	if isEnd {
		p.w.Close()
		err = <-p.readClosed
	}

	// If the HTTP request or response is longer than our maximum length, close the pipe
	// anyway. This will leave the input stream in a state where it probably can't find
	// the next header until the accumulated data in the reassembly buffer is all skipped.
	if p.allInput.Len() > p.maxHttpLength {
		p.w.Close()
		err = <-p.readClosed
	}

	return
}

func newHTTPParser(isRequest bool, bidiID akinet.TCPBidiID, seq, ack reassembly.Sequence) *httpParser {
	// Unfortunately, go's http request parser blocks. So we need to run it in a
	// separate goroutine. This needs to be addressed as part of
	// https://app.clubhouse.io/akita-software/story/600
	resultChan := make(chan akinet.ParsedNetworkContent)
	readClosed := make(chan error, 1)
	r, w := io.Pipe()
	go func() {
		var req *http.Request
		var resp *http.Response
		var body []byte
		var err error
		br := bufio.NewReader(r)
		if isRequest {
			req, body, err = readSingleHTTPRequest(br)
		} else {
			resp, body, err = readSingleHTTPResponse(br)
		}
		if err != nil {
			err = httpPipeReaderError{
				err:         err,
				unusedBytes: int64(br.Buffered()),
			}
			r.CloseWithError(err)
			readClosed <- err
			return
		}

		// Close the reader to signal to the pipe writer that result is ready.
		err = httpPipeReaderDone(br.Buffered())
		r.CloseWithError(err)
		readClosed <- err

		var c akinet.ParsedNetworkContent
		if isRequest {
			// Because HTTP requires the request to finish before sending a response,
			// TCP ack number on the first segment of the HTTP request is equal to the
			// TCP seq number on the first segment of the corresponding HTTP response.
			// Hence we use it to differntiate differnt pairs of HTTP request and
			// response on the same TCP stream.
			c = akinet.FromStdRequest(uuid.UUID(bidiID), int(ack), req, body)
		} else {
			// Because HTTP requires the request to finish before sending a response,
			// TCP ack number on the first segment of the HTTP request is equal to the
			// TCP seq number on the first segment of the corresponding HTTP response.
			// Hence we use it to differntiate differnt pairs of HTTP request and
			// response on the same TCP stream.
			c = akinet.FromStdResponse(uuid.UUID(bidiID), int(seq), resp, body)
		}
		resultChan <- c
	}()

	return &httpParser{
		w:             w,
		resultChan:    resultChan,
		readClosed:    readClosed,
		isRequest:     isRequest,
		maxHttpLength: MaximumHTTPLength,
	}
}

// Reads a single HTTP request, only consuming the exact number of bytes that
// form the request and its body, but there may be unused bytes left in the
// bufio.Reader's buffer.
func readSingleHTTPRequest(r *bufio.Reader) (*http.Request, []byte, error) {
	req, err := http.ReadRequest(r)
	if err != nil {
		return nil, nil, err
	}

	if req.Body == nil {
		return req, nil, nil
	}

	// Read the body to move the reader's position to the end of the body.
	var body bytes.Buffer
	_, bodyErr := io.Copy(&body, req.Body)
	req.Body.Close()
	return req, body.Bytes(), bodyErr
}

// Reads a single HTTP response, only consuming the exact number of bytes that
// form the responseand its body, but there may be unused bytes left in the
// bufio.Reader's buffer.
func readSingleHTTPResponse(r *bufio.Reader) (*http.Response, []byte, error) {
	resp, err := http.ReadResponse(r, nil)
	if err != nil {
		return nil, nil, err
	}

	if resp.Body == nil {
		return resp, nil, nil
	}

	// Read the body to move the reader's position to the end of the body.
	var body bytes.Buffer
	_, bodyErr := io.Copy(&body, resp.Body)
	resp.Body.Close()

	if errors.Is(bodyErr, io.ErrUnexpectedEOF) {
		// Let the next level try to handle a body that was truncated.
		bodyErr = nil
	}

	return resp, body.Bytes(), bodyErr
}

// Indicates the pipe reader has successfully completed parsing. The integer
// specifies the number of bytes read from the pipe writer but were unused.
type httpPipeReaderDone int64

func (httpPipeReaderDone) Error() string {
	return "HTTP pipe reader success"
}

type httpPipeReaderError struct {
	err         error // the actual err
	unusedBytes int64 // number of bytes read from the pipe writer but were unused
}

func (e httpPipeReaderError) Error() string {
	return e.err.Error()
}
