package client_telemetry

// We produce a set of packet counters indexed by interface, host and
// port number (*either* source or destination.)
type PacketCounts struct {
	// Flow
	Interface string `json:"interface"`
	SrcHost   string `json:"src_host"`
	DstHost   string `json:"dst_host"`
	SrcPort   int    `json:"src_port"`
	DstPort   int    `json:"dst_port"`

	// Number of events
	TCPPackets              int `json:"tcp_packets"`
	HTTPRequests            int `json:"http_requests"`
	HTTPResponses           int `json:"http_responses"`
	HTTPRequestsRateLimited int `json:"http_requests_rate_limited"`
	OversizedWitnesses      int `json:"oversized_witnesses"` // These witnesses were dropped.
	TLSHello                int `json:"tls_hello"`
	HTTP2Prefaces           int `json:"http2_prefaces"`
	QUICHandshakes          int `json:"quic_handshakes"`
	Unparsed                int `json:"unparsed"`
}

func (c *PacketCounts) Add(d PacketCounts) {
	c.TCPPackets += d.TCPPackets
	c.HTTPRequests += d.HTTPRequests
	c.HTTPResponses += d.HTTPResponses
	c.HTTPRequestsRateLimited += d.HTTPRequestsRateLimited
	c.OversizedWitnesses += d.OversizedWitnesses
	c.TLSHello += d.TLSHello
	c.HTTP2Prefaces += d.HTTP2Prefaces
	c.QUICHandshakes += d.QUICHandshakes
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
const Version = "v0.3"

type PacketCountSummary struct {
	Version        string                   `json:"version"`
	Total          PacketCounts             `json:"total"`
	TopByPort      map[int]*PacketCounts    `json:"top_by_port"`
	TopByInterface map[string]*PacketCounts `json:"top_by_interface"`
	TopByHost      map[string]*PacketCounts `json:"top_by_host"`

	// Maximum number of elements allowed in the TopByX maps.
	ByPortOverflowLimit      int `json:"by_port_overflow_limit"`
	ByInterfaceOverflowLimit int `json:"by_interface_overflow_limit"`
	ByHostOverflowLimit      int `json:"by_host_overflow_limit"`

	// Counts for elements in excess of the overflow limits.  Nil means
	// there was no overflow.
	ByPortOverflow      *PacketCounts `json:"by_port_overflow,omitempty"`
	ByInterfaceOverflow *PacketCounts `json:"by_interface_overflow,omitempty"`
	ByHostOverflow      *PacketCounts `json:"by_host_overflow,omitempty"`
}

func NewPacketCountSummary() *PacketCountSummary {
	return &PacketCountSummary{
		Version:        Version,
		TopByPort:      make(map[int]*PacketCounts),
		TopByInterface: make(map[string]*PacketCounts),
	}
}
