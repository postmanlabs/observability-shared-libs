package tls

import (
	"crypto/x509"

	"github.com/akitasoftware/akita-libs/akid"
	"github.com/akitasoftware/akita-libs/akinet"
	"github.com/akitasoftware/akita-libs/memview"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func newTLSServerHelloParser(bidiID akinet.TCPBidiID) *tlsServerHelloParser {
	return &tlsServerHelloParser{
		connectionID: akid.NewConnectionID(uuid.UUID(bidiID)),
	}
}

type tlsServerHelloParser struct {
	connectionID akid.ConnectionID
	allInput     memview.MemView
}

var _ akinet.TCPParser = (*tlsServerHelloParser)(nil)

func (*tlsServerHelloParser) Name() string {
	return "TLS 1.2/1.3 Server-Hello Parser"
}

func (parser *tlsServerHelloParser) Parse(input memview.MemView, isEnd bool) (result akinet.ParsedNetworkContent, unused memview.MemView, err error) {
	result, numBytesConsumed, err := parser.parse(input, isEnd)
	// It's an error if we're at the end and we don't yet have a result.
	if isEnd && result == nil && err == nil {
		// We never got the full TLS record. This is an error.
		err = errors.New("incomplete TLS record for Server Hello")
	}

	// If we have an error, then cannot consume any input according to the
	// contract for Parse.
	if err != nil {
		numBytesConsumed = 0
	}

	unused = parser.allInput.SubView(numBytesConsumed, parser.allInput.Len())
	return result, unused, err
}

func (parser *tlsServerHelloParser) parse(input memview.MemView, isEnd bool) (result akinet.ParsedNetworkContent, numBytesConsumed int64, err error) {
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
	buf, err = seek(buf, handshakeHeaderLength_bytes+serverVersionLength_bytes+serverRandomLength_bytes)
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

	// Seek past more headers.
	buf, err = seek(buf, serverCiphersuiteLength_bytes+serverCompressionMethodLength_bytes)
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

	selectedVersion := akinet.TLS_v1_2
	selectedProtocol := (*string)(nil)
	dnsNames := ([]string)(nil)

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
		case supportedVersionsTLSExtensionID:
			version, err := parser.parseSupportedVersionsExtension(extensionContent)
			if err == nil {
				selectedVersion = version
			}

		case alpnTLSExtensionID:
			protocol, err := parser.parseALPNExtension(extensionContent)
			if err == nil {
				selectedProtocol = &protocol
			}
		}
	}

	if selectedVersion == akinet.TLS_v1_2 {
		// We have TLS 1.2. There should be a second TLS record with a handshake
		// message containing the server's certificate. Get the certificate's CN and
		// SANs.

		// Get a view of the bytes after the first handshake message.
		buf := parser.allInput.SubView(handshakeMsgEndPos, parser.allInput.Len())

		// Wait until we have at least the header for the second TLS record.
		if buf.Len() < tlsRecordHeaderLength_bytes {
			return nil, 0, nil
		}

		// Expect the first three bytes to be as follows:
		//   0x16 - handshake record
		//   0x0303 - protocol version 3.3 (TLS 1.2)
		for idx, expectedByte := range []byte{0x16, 0x03, 0x03} {
			if buf.GetByte(int64(idx)) != expectedByte {
				return nil, 0, errors.New("expected a TLS message containing the server's certificate, but found a malformed TLS record")
			}
		}

		// The last two bytes of the record header give the total length of the
		// handshake message that appears after the record header.
		handshakeMsgLen_bytes := buf.GetUint16(tlsRecordHeaderLength_bytes - 2)
		handshakeMsgEndPos = int64(tlsRecordHeaderLength_bytes + handshakeMsgLen_bytes)

		// Wait until we have the full handshake record.
		if buf.Len() < handshakeMsgEndPos {
			return nil, 0, nil
		}

		// Get a Memview of the handshake record.
		buf = buf.SubView(tlsRecordHeaderLength_bytes, handshakeMsgEndPos)

		// The first byte of the handshake message gives its type. Expect a
		// certificate handshake message (type 0x0b).
		messageType := buf.GetByte(0)
		buf, err = seek(buf, 1)
		if err != nil {
			return nil, 0, errors.New("expected a TLS message containing the server's certificate, but found a malformed handshake message")
		}
		if messageType != 0x0b {
			return nil, 0, errors.Errorf("expected a TLS certificate handshake message (type 13) containing the server's certificate, but found a type %d handshake message", messageType)
		}

		// The next three bytes gives the length of the certificate message.
		certMsgLen_bytes := int64(buf.GetUint24(0))
		buf, err = seek(buf, 3)
		if err != nil {
			return nil, 0, errors.New("expected a TLS message containing the server's certificate, but found a malformed certificate handshake message")
		}

		// Isolate the section that contains the certificate message.
		if buf.Len() < certMsgLen_bytes {
			return nil, 0, errors.New("expected a TLS message containing the server's certificate, but found a malformed certificate handshake message")
		}
		buf = buf.SubView(0, certMsgLen_bytes)

		// The next three bytes gives the length of the certificate data that
		// follows.
		certDataLen_bytes := int64(buf.GetUint24(0))
		buf, err = seek(buf, 3)
		if err != nil {
			return nil, 0, errors.New("expected a TLS message containing the server's certificate, but found a malformed certificate handshake message")
		}

		// Isolate the section that contains the certificate data.
		if buf.Len() < certDataLen_bytes {
			return nil, 0, errors.New("expected a TLS message containing the server's certificate, but found a malformed certificate handshake message")
		}
		buf = buf.SubView(0, certDataLen_bytes)

		// The first certificate is the one that was issued to the server, so we
		// only need to look at that.

		// The next three bytes gives the length of the first certificate.
		certLen_bytes := int64(buf.GetUint24(0))
		buf, err = seek(buf, 3)
		if err != nil {
			return nil, 0, errors.New("expected a TLS message containing the server's certificate, but found a malformed certificate handshake message")
		}

		// Isolate the section that contains the first certificate.
		if buf.Len() < certLen_bytes {
			return nil, 0, errors.New("expected a TLS message containing the server's certificate, but found a malformed certificate handshake message")
		}
		buf = buf.SubView(0, certLen_bytes)

		var cert *x509.Certificate
		cert, err = x509.ParseCertificate([]byte(buf.String()))
		if err != nil {
			return nil, 0, errors.Wrap(err, "error parsing server certificate")
		}

		dnsNames = cert.DNSNames
	}

	hello := akinet.TLSServerHello{
		ConnectionID:     parser.connectionID,
		Version:          selectedVersion,
		SelectedProtocol: selectedProtocol,
		DNSNames:         dnsNames,
	}

	return hello, handshakeMsgEndPos, nil
}

// Extracts the server-selected TLS version from a buffer containing a TLS
// Supported Versions extension.
func (*tlsServerHelloParser) parseSupportedVersionsExtension(buf memview.MemView) (selectedVersion akinet.TLSVersion, err error) {
	if buf.Len() < 2 {
		return "", errors.New("malformed Supported Versions extension")
	}

	selected := buf.GetUint16(0)
	if result, exists := tlsVersionMap[selected]; exists {
		return result, nil
	}

	return "", errors.Errorf("unknown TLS version selected: %d", selected)
}

// Extracts the server-selected application-layer protocol from a buffer
// containing a TLS ALPN extension.
func (*tlsServerHelloParser) parseALPNExtension(buf memview.MemView) (string, error) {
	// The first two bytes give the length of the rest of the ALPN extension.
	length := int64(buf.GetUint16(0))
	buf, err := seek(buf, 2)
	if err != nil {
		return "", err
	}

	// Isolate the section that contains the rest of the ALPN extension.
	if buf.Len() < length {
		return "", errors.New("malformed ALPN extension")
	}
	buf = buf.SubView(0, length)

	// The next byte gives the length of the string indicating the selected
	// protocol.
	length = int64(buf.GetByte(0))
	buf, err = seek(buf, 1)
	if err != nil {
		return "", err
	}

	if buf.Len() < length {
		return "", errors.New("malformed ALPN extension")
	}
	return buf.SubView(0, length).String(), nil
}
