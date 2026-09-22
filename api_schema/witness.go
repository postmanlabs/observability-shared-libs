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

	// Pipeline timing checkpoints. Nil unless the agent sampled this witness.
	EventTimestamps *EventTimestamps `json:"event_timestamps,omitempty"`

	// Hash of the witness proto. Only used internally in the client.
	Hash string `json:"-"`
}

// EventTimestamps records when a single witness reached each stage of the
// pipeline, in microseconds since the Unix epoch, so that end-to-end latency
// can be broken down by stage. Each stage fills in its own fields and never
// rewrites another's.
//
// Agent-stamped and backend-stamped fields are on DIFFERENT CLOCKS. Only
// subtract two fields stamped by the same side; the interval between the two
// sides is transit plus clock skew, not elapsed work.
//
// # A zero field means the stage was not reached -- omitempty keeps it off the wire
//
// Most fields are stamped per witness. Uploaded and IngestRecv are stamped once
// per upload batch and shared by every witness in it, so they are identical
// across a batch and the spread within one is not observable. An interval that
// ends at either of them therefore includes waiting for the rest of the batch
// to accumulate, which is real latency but is queueing, not work done on that
// witness.
type EventTimestamps struct {
	// Stamped by the agent. The first three are packet capture timestamps on
	// the kernel's clock; the rest are the agent's wall clock.
	ReqRecv   int64 `json:"req_recv,omitempty"`   // last packet of the request
	RespStart int64 `json:"resp_start,omitempty"` // first packet of the response
	RespRecv  int64 `json:"resp_recv,omitempty"`  // last packet of the response
	Paired    int64 `json:"paired,omitempty"`     // request and response merged
	Redacted  int64 `json:"redacted,omitempty"`   // redaction returned
	Batched   int64 `json:"batched,omitempty"`    // batcher handed it to the upload buffer
	Buffered  int64 `json:"buffered,omitempty"`   // added to the upload report
	Uploaded  int64 `json:"uploaded,omitempty"`   // batch dispatched: just before the POST, per batch

	// Stamped by the back end
	IngestRecv int64 `json:"ingest_recv,omitempty"` // batch received by async_witnesses, per batch
	KafkaPub   int64 `json:"kafka_pub,omitempty"`   // published to Kafka
	AsmRecv    int64 `json:"asm_recv,omitempty"`    // witness_assembler consumed it
	ChInsert   int64 `json:"ch_insert,omitempty"`   // inserted into ClickHouse
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

	// EventTimestamps, counted at its maximum: the object wrapper plus every
	// field present. Each field costs its quoted name, a colon, a comma, and up
	// to 16 digits of microsecond timestamp. Fields left zero are omitted from
	// the JSON, so this over-counts rather than under-counts, which is the safe
	// direction for the callers that gate on it.
	if report.EventTimestamps != nil {
		result += len(`"event_timestamps":{},`)
		result += eventTimestampsMaxSize_bytes
	}

	return result
}

// The largest EventTimestamps object that can appear on the wire: every field
// set, each costing its quoted name plus a colon, a comma, and 16 digits.
var eventTimestampsMaxSize_bytes = func() int {
	result := 0
	for _, name := range []string{
		"req_recv", "resp_start", "resp_recv", "paired",
		"redacted", "batched", "buffered", "uploaded",
		"ingest_recv", "kafka_pub", "asm_recv", "ch_insert",
	} {
		result += len(name) + 20
	}
	return result
}()
