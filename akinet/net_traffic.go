package akinet

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/akitasoftware/akita-libs/memview"
)

// Represents a generic network traffic that has been parsed from the wire.
type ParsedNetworkTraffic struct {
	SrcIP   net.IP
	SrcPort int
	DstIP   net.IP
	DstPort int
	Content ParsedNetworkContent

	// The time at which the first packet was observed
	ObservationTime time.Time

	// The time at which the final packet arrived, for
	// multi-packet content.  Equal to ObservationTime
	// for single packets.
	FinalPacketTime time.Time
}

// Interface implemented by all types of data that can be parsed from the
// network.
type ParsedNetworkContent interface {
	implParsedNetworkContent()
}

type RawBytes memview.MemView

func (RawBytes) implParsedNetworkContent() {}

func (rb RawBytes) String() string {
	return memview.MemView(rb).String()
}

type HTTPRequest struct {
	// StreamID and Seq uniquely identify a pair of request and response.
	StreamID uuid.UUID
	Seq      int

	Method           string
	ProtoMajor       int // e.g. 1 in HTTP/1.0
	ProtoMinor       int // e.g. 0 in HTTP/1.0
	URL              *url.URL
	Host             string
	Header           http.Header
	Body             []byte // nil means no body
	BodyDecompressed bool   // true if the body is already decompressed
	Cookies          []*http.Cookie
}

func (HTTPRequest) implParsedNetworkContent() {}

// Returns a string key that associates this request with its corresponding
// response.
func (r HTTPRequest) GetStreamKey() string {
	return r.StreamID.String() + ":" + strconv.Itoa(r.Seq)
}

type HTTPResponse struct {
	// StreamID and Seq uniquely identify a pair of request and response.
	StreamID uuid.UUID
	Seq      int

	StatusCode       int
	ProtoMajor       int // e.g. 1 in HTTP/1.0
	ProtoMinor       int // e.g. 0 in HTTP/1.0
	Header           http.Header
	Body             []byte // nil means no body
	BodyDecompressed bool   // true if the body is already decompressed
	Cookies          []*http.Cookie
}

func (HTTPResponse) implParsedNetworkContent() {}

// Returns a string key that associates this response with its corresponding
// request.
func (r HTTPResponse) GetStreamKey() string {
	return r.StreamID.String() + ":" + strconv.Itoa(r.Seq)
}

// For testing only.
type AkitaPrince string

func (AkitaPrince) implParsedNetworkContent() {}

// For testing only.
type AkitaPineapple string

func (AkitaPineapple) implParsedNetworkContent() {}
