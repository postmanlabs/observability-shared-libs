package api_schema

import (
	"net"
	"time"

	"github.com/akitasoftware/akita-libs/akid"
	"github.com/akitasoftware/akita-libs/akinet"
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

	// If true, source is known to have initiated the connection. Otherwise,
	// "source" and "destination" is arbitrary.
	InitiatorKnown bool `json:"initiator_known"`

	// Whether and how the connection was closed.
	EndState akinet.TCPConnectionEndState `json:"end_state"`
}

func (report TCPConnectionReport) GetID() akid.ID {
	return report.ID
}

// Returns an approximation of the size of this report.
func (report *TCPConnectionReport) SizeInBytes() int {
	result := 26                     // ID
	result += len("255.255.255.255") // SrcAddr
	result += len("65535")           // SrcPort
	result += len("255.255.255.255") // DstAddr
	result += len("65535")           // DstAddr
	result += len(time.RFC3339Nano)  // FirstObserved
	result += len(time.RFC3339Nano)  // LastObserved
	result += len("false")           // InitiatorKnown
	result += len(report.EndState)   // EndState
	return result
}
