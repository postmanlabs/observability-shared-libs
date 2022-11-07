package api_schema

import (
	"net"
	"time"

	"github.com/akitasoftware/akita-libs/akid"
)

type WitnessReport struct {
	// CLI v0.20.0 and later will only ever provide "INBOUND" reports. Anything
	// marked "OUTBOUND" is ignored by the Akita back end.
	Direction NetworkDirection `json:"direction"`

	OriginAddr      net.IP `json:"origin_addr"`
	OriginPort      uint16 `json:"origin_port"`
	DestinationAddr net.IP `json:"destination_addr"`
	DestinationPort uint16 `json:"destination_port"`

	ClientWitnessTime time.Time `json:"client_witness_time"`

	// A serialized Witness protobuf in base64 URL encoded format.
	WitnessProto string `json:"witness_proto"`

	ID akid.WitnessID `json:"id"`

	// Hash of the witness proto. Only used internally in the client.
	Hash string `json:"-"`
}

// Returns an approximation of the size of this report.
func (report *WitnessReport) SizeInBytes() int {
	result := 0
	result += len("INBOUND")           // Direction
	result += len("255.255.255.255")   // OriginAddr
	result += len("65535")             // OriginPort
	result += len("255.255.255.255")   // DestinationAddr
	result += len("65535")             // DestinationPort
	result += len(time.RFC3339Nano)    // ClientWitnessTime
	result += len(report.WitnessProto) // WitnessProto
	result += 26                       // ID
	return result
}
