package api_schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// EventTimestamps is additive: a nil pointer must serialize to nothing at all, so an
// agent that never samples sends the exact same bytes it did before, and an old
// backend can still decode a report from a new agent.
func TestWitnessReportEventTimestampsOmittedWhenNil(t *testing.T) {
	b, err := json.Marshal(&WitnessReport{})
	assert.NoError(t, err)
	assert.NotContains(t, string(b), "event_timestamps")
}

// A stage that was never reached must not appear on the wire, so that a witness
// carries only the checkpoints it actually reached.
func TestEventTimestampsOmitsUnsetFields(t *testing.T) {
	b, err := json.Marshal(&WitnessReport{
		EventTimestamps: &EventTimestamps{ReqRecv: 1790071339202964},
	})
	assert.NoError(t, err)
	assert.Contains(t, string(b), `"req_recv":1790071339202964`)
	assert.NotContains(t, string(b), "resp_start")
	assert.NotContains(t, string(b), "ch_insert")
}

// SizeInBytes gates the agent's upload batch, so it must never under-count what
// EventTimestamps adds -- including the fields stamped after the estimate is taken.
func TestWitnessReportSizeInBytesCoversEventTimestamps(t *testing.T) {
	bare := WitnessReport{WitnessProto: "abc"}

	full := bare
	full.EventTimestamps = &EventTimestamps{
		ReqRecv: 1790071339202964, RespStart: 1790071339202965, RespRecv: 1790071339202966,
		Paired: 1790071339203100, Redacted: 1790071339203200, Batched: 1790071339203300,
		Buffered: 1790071339203400, Uploaded: 1790071339204000,
		IngestRecv: 1790071339205000, KafkaPub: 1790071339206000,
		AsmRecv: 1790071339207000, ChInsert: 1790071339208000,
	}

	bareJSON, err := json.Marshal(&bare)
	assert.NoError(t, err)
	fullJSON, err := json.Marshal(&full)
	assert.NoError(t, err)

	actual := len(fullJSON) - len(bareJSON)
	estimated := full.SizeInBytes() - bare.SizeInBytes()
	assert.GreaterOrEqual(t, estimated, actual,
		"SizeInBytes under-counts: estimated %d, actual %d", estimated, actual)
}

// A partly-filled witness must also be covered, since the agent takes the size
// estimate before the last checkpoints are stamped.
func TestWitnessReportSizeInBytesCoversPartialEventTimestamps(t *testing.T) {
	bare := WitnessReport{WitnessProto: "abc"}
	partial := bare
	partial.EventTimestamps = &EventTimestamps{ReqRecv: 1790071339202964}

	bareJSON, _ := json.Marshal(&bare)
	partialJSON, _ := json.Marshal(&partial)

	assert.GreaterOrEqual(t,
		partial.SizeInBytes()-bare.SizeInBytes(),
		len(partialJSON)-len(bareJSON))
}

// Old agents and new back ends must interoperate: the JSON shape is unchanged
// by moving from a map to a struct.
func TestEventTimestampsDecodesLegacyMapPayload(t *testing.T) {
	var report WitnessReport
	err := json.Unmarshal([]byte(`{"event_timestamps":{"req_recv":42,"unknown_key":7}}`), &report)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), report.EventTimestamps.ReqRecv)
}
