package tls

import (
	"github.com/akitasoftware/akita-libs/akinet"
	"github.com/akitasoftware/akita-libs/memview"
	"github.com/google/gopacket/reassembly"
)

// Returns a parser factory for the server half of a TLS connection.
func NewTLSServerParserFactory() akinet.TCPParserFactory {
	return &tlsServerParserFactory{}
}

type tlsServerParserFactory struct{}

func (*tlsServerParserFactory) Name() string {
	return "TLS 1.2/1.3 Server Parser Factory"
}

func (factory *tlsServerParserFactory) Accepts(input memview.MemView, isEnd bool) (decision akinet.AcceptDecision, discardFront int64) {
	decision, discardFront = factory.accepts(input)

	if decision == akinet.NeedMoreData && isEnd {
		decision = akinet.Reject
		discardFront = input.Len()
	}

	return decision, discardFront
}

var serverHelloHandshakeBytes = []byte{
	// Record header (5 bytes)
	0x16,       // handshake record
	0x03, 0x03, // protocol version 3.3 (TLS 1.2)
	0x00, 0x00, // handshake payload size (ignored)

	// Handshake header (4 bytes)
	0x02,             // Server Hello
	0x00, 0x00, 0x00, // Server Hello payload size (ignored)

	// Server Version (2 bytes)
	0x03, 0x03, // protocol version 3.3 (TLS 1.2)
}

var serverHelloHandshakeMask = []byte{
	// Record header (5 bytes)
	0xff,       // handshake record
	0xff, 0xff, // protocol version
	0x00, 0x00, // handshake payload size (ignored)

	// Handshake header (4 bytes)
	0xff,             // Server Hello
	0x00, 0x00, 0x00, // Server Hello payload size (ignored)

	// Server Version (2 bytes)
	0xff, 0xff, // protocol version
}

func (*tlsServerParserFactory) accepts(input memview.MemView) (decision akinet.AcceptDecision, discardFront int64) {
	if input.Len() < minTLSServerHelloLength_bytes {
		return akinet.NeedMoreData, 0
	}

	// Accept if we match a "Server Hello" handshake message. Reject if we fail to
	// match.
	for idx, expectedByte := range serverHelloHandshakeBytes {
		if input.GetByte(int64(idx))&serverHelloHandshakeMask[idx] != expectedByte {
			return akinet.Reject, input.Len()
		}
	}

	return akinet.Accept, 0
}

func (factory *tlsServerParserFactory) CreateParser(id akinet.TCPBidiID, seq, ack reassembly.Sequence) akinet.TCPParser {
	return newTLSServerHelloParser(id)
}
