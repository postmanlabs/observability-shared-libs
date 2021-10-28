package tls

import (
	"encoding/binary"
	"errors"

	"github.com/akitasoftware/akita-libs/memview"
)

// Returns the uint16 at buf[offset..offset+1]. Returns 0 if offset+1 >=
// len(buf).
func uint16At(buf memview.MemView, offset int64) uint16 {
	if buf.Len() <= offset+1 {
		return 0
	}
	return binary.BigEndian.Uint16([]byte(buf.SubView(offset, offset+2).String()))
}

// Returns the uint24 at buf[offset..offset+2] as a uint32. Return 0 if offset+2
// >= len(buf).
func uint24At(buf memview.MemView, offset int64) uint32 {
	if buf.Len() <= offset+2 {
		return 0
	}

	slice := []byte{0}
	slice = append(slice, []byte(buf.SubView(offset, offset+3).String())...)
	return binary.BigEndian.Uint32(slice)
}

// Returns a view of buf[offset..len(buf)]. Returns an error if offset >
// len(buf).
func seek(buf memview.MemView, offset int64) (memview.MemView, error) {
	if offset > buf.Len() {
		return buf, errors.New("malformed message")
	}

	return buf.SubView(offset, buf.Len()), nil
}
