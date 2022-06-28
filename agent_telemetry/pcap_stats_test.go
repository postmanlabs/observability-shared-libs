package agent_telemetry

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPacketCaptureStatsJSON(t *testing.T) {
	empty := NewPacketCountSummary()

	one := NewPacketCountSummary()
	one.Update(PacketCounters{
		Interface:     "lo0",
		SrcPort:       80,
		DstPort:       80,
		TCPPackets:    1,
		HTTPRequests:  1,
		HTTPResponses: 1,
		Unparsed:      1,
	})

	two := NewPacketCountSummary()
	one.Update(PacketCounters{
		Interface:     "lo0",
		SrcPort:       80,
		DstPort:       80,
		TCPPackets:    1,
		HTTPRequests:  1,
		HTTPResponses: 1,
		Unparsed:      1,
	})
	two.Update(PacketCounters{
		Interface:     "if1",
		SrcPort:       443,
		DstPort:       443,
		TCPPackets:    2,
		HTTPRequests:  2,
		HTTPResponses: 2,
		Unparsed:      2,
	})

	testCases := []struct {
		name  string
		stats *PacketCountSummary
	}{
		{name: "empty", stats: empty},
		{name: "one interface, one port", stats: one},
		{name: "two interfaces, two ports", stats: two},
	}

	for _, tc := range testCases {
		bs, err := json.Marshal(tc.stats)
		assert.NoError(t, err, "[%s] failed to marshal", tc.name)

		var unmarshalled PacketCountSummary
		err = json.Unmarshal(bs, &unmarshalled)
		assert.NoError(t, err, "[%s] failed to unmarshal", tc.name)

		assert.Equal(t, tc.stats, &unmarshalled, "[%s] not equal", tc.name)
	}
}
