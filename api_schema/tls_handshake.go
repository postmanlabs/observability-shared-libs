package api_schema

import (
	"github.com/akitasoftware/akita-libs/akid"
	"github.com/akitasoftware/akita-libs/akinet"
)

// Details about a TLS handshake that was observed.
type TLSHandshakeReport struct {
	ID akid.ConnectionID `json:"id"`

	// The inferred TLS version. Only populated if the Server Hello was seen.
	Version *akinet.TLSVersion

	// The DNS hostname extracted from the client's SNI extension, if any.
	SNIHostname *string

	// The list of protocols supported by the client, as seen in the ALPN
	// extension.
	SupportedProtocols []string

	// The selected application-layer protocol, as seen in the server's ALPN
	// extension, if any.
	SelectedProtocol *string

	// The SANs seen in the server's certificate. The server's certificate is
	// encrypted in TLS 1.3, so this is only populated for TLS 1.2 connections.
	SubjectAlternativeNames []string
}
