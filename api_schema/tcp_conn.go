package api_schema

import (
	"net"
	"time"

	"github.com/akitasoftware/akita-libs/akid"
)

// Details about a TCP connection that was observed.
type TCPConnectionReport struct {
	ID akid.ConnectionID `json:"id"`

	SrcAddr  net.IP `json:"src_addr"`
	SrcPort  uint16 `json:"src_port"`
	DestAddr net.IP `json:"dest_addr"`
	DestPort uint16 `json:"dest_port"`

	FirstObserved time.Time `json:"first_observed"`
	LastObserved  time.Time `json:"last_observed"`

	Direction NetworkDirection `json:"direction"`

	// If true, source is known to have initiated the connection. Otherwise,
	// "source" and "destination" is arbitrary.
	InitiatorKnown bool `json:"initiator_known"`

	// Whether and how the connection was closed.
	EndState TCPConnectionEndState `json:"end_state"`
}

// Indicates whether a TCP connection was closed, and if so, how.
type TCPConnectionEndState string

const (
	// Neither the FIN nor RST flag was seen.
	ConnectionOpen TCPConnectionEndState = "OPEN"

	// The FIN flag was seen, but not the RST flag.
	ConnectionClosed TCPConnectionEndState = "CLOSED"

	// The RST flag was seen.
	ConnectionReset TCPConnectionEndState = "RESET"
)

type UploadTCPConnectionReportsRequest struct {
	ClientID akid.ClientID          `json:"client_id"`
	Reports  []*TCPConnectionReport `json:"reports"`
}
