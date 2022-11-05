package api_schema

import "github.com/akitasoftware/akita-libs/akid"

type UploadReportsRequest struct {
	ClientID       akid.ClientID          `json:"client_id"`
	Witnesses      []*WitnessReport       `json:"witnesses"`
	TCPConnections []*TCPConnectionReport `json:"tcp_connections"`
	TLSHandshakes  []*TLSHandshakeReport  `json:"tls_handshakes"`

	// An approximation of the size of the reports in this request.
	reportsSize_bytes int `json:"-"`
}

func (req *UploadReportsRequest) AddWitnessReport(report *WitnessReport) {
	req.Witnesses = append(req.Witnesses, report)
	req.reportsSize_bytes += report.SizeInBytes()
}

func (req *UploadReportsRequest) AddTCPConnectionReport(report *TCPConnectionReport) {
	req.TCPConnections = append(req.TCPConnections, report)
	req.reportsSize_bytes += report.SizeInBytes()
}

func (req *UploadReportsRequest) AddTLSHandshakeReport(report *TLSHandshakeReport) {
	req.TLSHandshakes = append(req.TLSHandshakes, report)
	req.reportsSize_bytes += report.SizeInBytes()
}

func (req *UploadReportsRequest) IsEmpty() bool {
	return len(req.Witnesses)+len(req.TCPConnections)+len(req.TLSHandshakes) == 0
}

// Removes all reports from this request.
func (req *UploadReportsRequest) Clear() {
	// Clear without reallocating memory.
	req.Witnesses = req.Witnesses[:0]
	req.TCPConnections = req.TCPConnections[:0]
	req.TLSHandshakes = req.TLSHandshakes[:0]
	req.reportsSize_bytes = 0
}

// Returns an approximation of the size of this request.
func (req *UploadReportsRequest) SizeInBytes() int {
	return req.reportsSize_bytes + 26 // AKIDs are 26 bytes long
}
