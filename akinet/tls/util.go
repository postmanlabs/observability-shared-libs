package tls

import (
	"errors"

	"github.com/akitasoftware/akita-libs/memview"
)

// Returns a view of buf[offset..len(buf)]. Returns an error if offset >
// len(buf).
func seek(buf memview.MemView, offset int64) (memview.MemView, error) {
	if offset > buf.Len() {
		return buf, errors.New("malformed message")
	}

	return buf.SubView(offset, buf.Len()), nil
}
