package tls

import (
	"errors"

	"github.com/akitasoftware/akita-libs/akid"
	"github.com/akitasoftware/akita-libs/akinet"
	"github.com/akitasoftware/akita-libs/memview"
	"github.com/google/uuid"
)

func newTLSClientHelloParser(bidiID akinet.TCPBidiID) *tlsClientHelloParser {
	return &tlsClientHelloParser{
		connectionID: akid.NewConnectionID(uuid.UUID(bidiID)),
	}
}

type tlsClientHelloParser struct {
	connectionID akid.ConnectionID
	allInput     memview.MemView
}

var _ akinet.TCPParser = (*tlsClientHelloParser)(nil)

func (*tlsClientHelloParser) Name() string {
	return "TLS 1.2/1.3 Client-Hello Parser"
}

func (parser *tlsClientHelloParser) Parse(input memview.MemView, isEnd bool) (result akinet.ParsedNetworkContent, unused memview.MemView, err error) {
	result, numBytesConsumed, err := parser.parse(input, isEnd)
	// It's an error if we're at the end and we don't yet have a result.
	if isEnd && result == nil && err == nil {
		// We never got the full TLS record. This is an error.
		err = errors.New("incomplete TLS record for Client Hello")
	}

	// If we have an error, then we cannot consume any input according to the
	// contract for Parse.
	if err != nil {
		numBytesConsumed = 0
	}

	unused = parser.allInput.SubView(numBytesConsumed, parser.allInput.Len())
	return result, unused, err
}

func (parser *tlsClientHelloParser) parse(input memview.MemView, isEnd bool) (result akinet.ParsedNetworkContent, numBytesConsumed int64, err error) {
	// Add the incoming bytes to our buffer.
	parser.allInput.Append(input)

	// Wait until we have at least the TLS record header.
	if parser.allInput.Len() < tlsRecordHeaderLength_bytes {
		return nil, 0, nil
	}

	// The last two bytes of the record header give the total length of the
	// handshake message that appears after the record header.
	handshakeMsgLen_bytes := parser.allInput.GetUint16(tlsRecordHeaderLength_bytes - 2)
	handshakeMsgEndPos := int64(tlsRecordHeaderLength_bytes + handshakeMsgLen_bytes)

	// Wait until we have the full handshake record.
	if parser.allInput.Len() < handshakeMsgEndPos {
		return nil, 0, nil
	}

	// Get a Memview of the handshake record.
	buf := parser.allInput.SubView(tlsRecordHeaderLength_bytes, handshakeMsgEndPos)

	// Seek past some headers.
	buf, err = seek(buf, handshakeHeaderLength_bytes+clientVersionLength_bytes+clientRandomLength_bytes)
	if err != nil {
		return nil, 0, err
	}

	// Now at the session ID, which is a variable-length vector. The first byte
	// indicates the vector's length in bytes.
	sessionIdLen_bytes := buf.GetByte(0)
	buf, err = seek(buf, int64(sessionIdLen_bytes)+1)
	if err != nil {
		return nil, 0, err
	}

	// Now at the cipher suites. The first two bytes gives the length of this
	// header in bytes.
	cipherSuitesLen_bytes := buf.GetUint16(0)
	buf, err = seek(buf, int64(cipherSuitesLen_bytes)+2)
	if err != nil {
		return nil, 0, err
	}

	// Now at the compression methods. The first byte gives the length of this
	// header in bytes.
	compressionMethodsLen_bytes := buf.GetByte(0)
	buf, err = seek(buf, int64(compressionMethodsLen_bytes)+1)
	if err != nil {
		return nil, 0, err
	}

	// Now at the extensions. The first two bytes gives the length of the
	// extensions in bytes.
	extensionsLength_bytes := buf.GetUint16(0)
	buf, err = seek(buf, 2)
	if err != nil {
		return nil, 0, err
	}

	// Isolate the section that contains the TLS extensions.
	if buf.Len() < int64(extensionsLength_bytes) {
		return nil, 0, errors.New("malformed TLS message")
	}
	buf = buf.SubView(0, int64(extensionsLength_bytes))

	dnsHostname := (*string)(nil)
	protocols := []string{}

	for buf.Len() > 0 {
		// The first two bytes of the extension give the extension type.
		extensionType := tlsExtensionID(buf.GetUint16(0))
		buf, err = seek(buf, 2)
		if err != nil {
			return nil, 0, err
		}

		// The following two bytes give the extension's content length in bytes.
		extensionContentLength_bytes := buf.GetUint16(0)
		buf, err = seek(buf, 2)
		if err != nil {
			return nil, 0, err
		}

		extensionContent := buf.SubView(0, int64(extensionContentLength_bytes))
		buf, err = seek(buf, int64(extensionContentLength_bytes))
		if err != nil {
			return nil, 0, err
		}

		switch extensionType {
		case serverNameTLSExtensionID:
			serverName, err := parser.parseServerNameExtension(extensionContent)
			if err == nil {
				dnsHostname = &serverName
			}

		case alpnTLSExtensionID:
			protocols = parser.parseALPNExtension(extensionContent)
		}
	}

	hello := akinet.TLSClientHello{
		ConnectionID:       parser.connectionID,
		Hostname:           dnsHostname,
		SupportedProtocols: protocols,
	}

	return hello, handshakeMsgEndPos, nil
}

// Extracts the DNS hostname from a buffer containing a TLS SNI extension.
func (*tlsClientHelloParser) parseServerNameExtension(buf memview.MemView) (hostname string, err error) {
	// The SNI extension is a list of server names, each of a different type.
	// Currently, the only supported type is DNS (type 0x00) according to RFC
	// 6066.
	for buf.Len() > 0 {
		// First two bytes gives the length of the list entry.
		entryLen_bytes := buf.GetUint16(0)
		buf, err = seek(buf, 2)
		if err != nil {
			return "", err
		}

		// Next byte is the entry type.
		entryType := sniType(buf.GetByte(0))
		buf, err = seek(buf, 1)
		if err != nil {
			return "", err
		}

		switch entryType {
		case dnsHostnameSNIType:
			// The next two bytes gives the length of the hostname in bytes.
			hostnameLen_bytes := buf.GetUint16(0)
			buf, err = seek(buf, 2)
			if err != nil {
				return "", err
			}

			if buf.Len() < int64(hostnameLen_bytes) {
				return "", errors.New("malformed SNI extension entry")
			}
			return buf.SubView(0, int64(hostnameLen_bytes)).String(), nil
		}

		buf, err = seek(buf, int64(entryLen_bytes)-1)
		if err != nil {
			return "", err
		}
	}

	return "", errors.New("no DNS hostname found in SNI extension")
}

// Extracts the list of protocols from a buffer containing a TLS ALPN extension.
func (*tlsClientHelloParser) parseALPNExtension(buf memview.MemView) []string {
	result := []string{}
	var err error

	// The ALPN extension is a list of strings indicating the protocols supported
	// by the client. The first two bytes gives the length of the list in bytes.
	listLen_bytes := buf.GetUint16(0)
	buf, err = seek(buf, 2)
	if err != nil {
		return result
	}

	// Isolate the section that contains just the list.
	if buf.Len() < int64(listLen_bytes) {
		return result
	}
	buf = buf.SubView(0, int64(listLen_bytes))

	for buf.Len() > 0 {
		// The first byte of each list element gives the length of the string in
		// bytes.
		entryLen_bytes := buf.GetByte(0)
		buf, err = seek(buf, 1)
		if err != nil {
			return result
		}

		if buf.Len() < int64(entryLen_bytes) {
			return result
		}

		result = append(result, buf.SubView(0, int64(entryLen_bytes)).String())
		buf, err = seek(buf, int64(entryLen_bytes))
		if err != nil {
			return result
		}
	}

	return result
}
