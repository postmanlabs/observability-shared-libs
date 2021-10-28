package tls

import (
	"github.com/akitasoftware/akita-libs/akinet"
	"github.com/akitasoftware/akita-libs/memview"
	"github.com/google/gopacket/reassembly"
)

// Returns a parser factory for the client half of a TLS connection.
func NewTLSClientParserFactory() akinet.TCPParserFactory {
	return &tlsClientParserFactory{}
}

type tlsClientParserFactory struct{}

func (*tlsClientParserFactory) Name() string {
	return "TLS 1.2/1.3 Client Parser Factory"
}

func (factory *tlsClientParserFactory) Accepts(input memview.MemView, isEnd bool) (decision akinet.AcceptDecision, discardFront int64) {
	decision, discardFront = factory.accepts(input)

	if decision == akinet.NeedMoreData && isEnd {
		decision = akinet.Reject
		discardFront = input.Len()
	}

	return decision, discardFront
}

func (*tlsClientParserFactory) accepts(input memview.MemView) (decision akinet.AcceptDecision, discardFront int64) {
	if input.Len() < minTLSClientHelloLength_bytes {
		return akinet.NeedMoreData, 0
	}

	// Accept if we match a "Client Hello" handshake message. Reject if we fail to
	// match.

	expectedBytes := map[int]byte{
		// Record header (5 bytes)
		0: 0x16,          // handshake record
		1: 0x03, 2: 0x01, // protocol version 3.1 (TLS 1.0)
		// 2 bytes of handshake payload size

		// Handshake header (4 bytes)
		5: 0x01, // Client Hello
		// 3 bytes of Client Hello payload size

		// Client Version (2 bytes)
		9: 0x03, 10: 0x03, // protocol version 3.3 (TLS 1.2)
	}

	for idx, expectedByte := range expectedBytes {
		if input.GetByte(int64(idx)) != expectedByte {
			return akinet.Reject, input.Len()
		}
	}

	return akinet.Accept, 0
}

func (factory *tlsClientParserFactory) CreateParser(id akinet.TCPBidiID, seq, ack reassembly.Sequence) akinet.TCPParser {
	return newTLSClientHelloParser(id)
}
