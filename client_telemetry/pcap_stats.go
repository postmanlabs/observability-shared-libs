package client_telemetry

// We produce a set of packet counters indexed by interface and
// port number (*either* source or destination.)
type PacketCounts struct {
	// Flow
	Interface string `json:"interface"`
	SrcPort   int    `json:"src_port"`
	DstPort   int    `json:"dst_port"`

	// Number of events
	TCPPackets    int `json:"tcp_packets"`
	HTTPRequests  int `json:"http_requests"`
	HTTPResponses int `json:"http_responses"`
	Unparsed      int `json:"unparsed"`
}

func (c *PacketCounts) Add(d PacketCounts) {
	c.TCPPackets += d.TCPPackets
	c.HTTPRequests += d.HTTPRequests
	c.HTTPResponses += d.HTTPResponses
	c.Unparsed += d.Unparsed
}

func (c *PacketCounts) Copy() *PacketCounts {
	if c == nil {
		return nil
	}
	copy := *c
	return &copy
}

// Reflects the version of the JSON encoding.  Increase the minor version
// number for backwards-compatible changes and the major number for non-
// backwards compatible changes.
const Version = "v0"

type PacketCountSummary struct {
	Version        string                   `json:"version"`
	Total          PacketCounts             `json:"total"`
	TopByPort      map[int]*PacketCounts    `json:"top_by_port"`
	TopByInterface map[string]*PacketCounts `json:"top_by_interface"`
}

func NewPacketCountSummary() *PacketCountSummary {
	return &PacketCountSummary{
		Version:        Version,
		TopByPort:      make(map[int]*PacketCounts),
		TopByInterface: make(map[string]*PacketCounts),
	}
}
