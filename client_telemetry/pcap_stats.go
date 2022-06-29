package client_telemetry

import (
	"encoding/json"
	"sync"
)

// We produce a set of packet counters indexed by interface and
// port number (*either* source or destination.)
type PacketCounters struct {
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

func (c *PacketCounters) Add(d PacketCounters) {
	c.TCPPackets += d.TCPPackets
	c.HTTPRequests += d.HTTPRequests
	c.HTTPResponses += d.HTTPResponses
	c.Unparsed += d.Unparsed
}

// A consumer accepts incremental updates in the form
// of PacketCounters.
type PacketCountConsumer interface {
	// Add an additional measurement to the current count
	Update(delta PacketCounters)
}

// Discard the count
type PacketCountDiscard struct {
}

func (d *PacketCountDiscard) Update(_ PacketCounters) {
}

// A consumer that sums the count by (interface, port) pairs.
// In the future, this could put counters on a pipe and do the increments
// in a separate goroutine, but we would *still* need a mutex to read the
// totals out.
// TODO: limit maximum size
type PacketCountSummary struct {
	total       PacketCounters
	byPort      map[int]*PacketCounters
	byInterface map[string]*PacketCounters
	mutex       sync.RWMutex `json:"-"`
}

func NewPacketCountSummary() *PacketCountSummary {
	return &PacketCountSummary{
		byPort:      make(map[int]*PacketCounters),
		byInterface: make(map[string]*PacketCounters),
	}
}

// Reflects the version of the JSON encoding.  Increase the minor version
// number for backwards-compatible changes and the major number for non-
// backwards compatible changes.
const Version = "v0"

// Used to implement JSON marshalling/unmarshalling for PacketCountSummary.
// We define custom methods in order to avoid making the fields of
// PacketCountSummary publicly accessible.
type packetCountSummaryExporter struct {
	Total       PacketCounters             `json:"total"`
	ByPort      map[int]*PacketCounters    `json:"by_port"`
	ByInterface map[string]*PacketCounters `json:"by_interface"`

	// Reflects the version of this JSON encoding.
	Version string `json:"version"`
}

func (s *PacketCountSummary) UnmarshalJSON(b []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var exporter packetCountSummaryExporter
	if err := json.Unmarshal(b, &exporter); err != nil {
		return err
	}

	s.total = exporter.Total
	s.byPort = exporter.ByPort
	s.byInterface = exporter.ByInterface

	return nil
}

func (s *PacketCountSummary) MarshalJSON() ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	exporter := &packetCountSummaryExporter{
		Total:       s.total,
		ByPort:      s.byPort,
		ByInterface: s.byInterface,
		Version:     Version,
	}
	return json.Marshal(exporter)
}

func (s *PacketCountSummary) Update(c PacketCounters) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if prev, ok := s.byPort[c.SrcPort]; ok {
		prev.Add(c)
	} else {
		new := &PacketCounters{
			Interface: "*",
			SrcPort:   c.SrcPort,
			DstPort:   0,
		}
		new.Add(c)
		s.byPort[new.SrcPort] = new
	}

	if prev, ok := s.byPort[c.DstPort]; ok {
		prev.Add(c)
	} else {
		// Use SrcPort as the identifier in the
		// accumulated counter
		new := &PacketCounters{
			Interface: "*",
			SrcPort:   c.DstPort,
			DstPort:   0,
		}
		new.Add(c)
		s.byPort[new.SrcPort] = new
	}

	if prev, ok := s.byInterface[c.Interface]; ok {
		prev.Add(c)
	} else {
		new := &PacketCounters{
			Interface: c.Interface,
			SrcPort:   0,
			DstPort:   0,
		}
		new.Add(c)
		s.byInterface[new.Interface] = new
	}

	s.total.Add(c)
}

func (s *PacketCountSummary) Total() PacketCounters {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.total
}

// Packet counters summed over interface
func (s *PacketCountSummary) TotalOnInterface(name string) PacketCounters {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if count, ok := s.byInterface[name]; ok {
		return *count
	}

	return PacketCounters{Interface: name}
}

// Packet counters summed over port
func (s *PacketCountSummary) TotalOnPort(port int) PacketCounters {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if count, ok := s.byPort[port]; ok {
		return *count
	}
	return PacketCounters{Interface: "*", SrcPort: port}
}

// All available port numbers
func (s *PacketCountSummary) AllPorts() []PacketCounters {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	ret := make([]PacketCounters, 0, len(s.byPort))
	for _, v := range s.byPort {
		ret = append(ret, *v)
	}
	return ret
}
