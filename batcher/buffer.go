package batcher

// Encapsulates a buffer of items and its consumer. The buffer has a soft size
// limit. Implementations will accept items that cause the buffer to exceed this
// limit, but clients are expected to call Flush on the buffer when the Add
// operation reports that the soft limit has been reached.
//
// Implementations do not need to be thread-safe.
type Buffer[Item any] interface {
	// Adds the given item to the buffer, even if it would cause the buffer to
	// exceed its size limit. Returns true if the buffer is at or exceeds its size
	// limit after this operation, and false otherwise.
	Add(item Item) (bool, error)

	// Flushes the buffer to the items' consumer. On return, the buffer will be
	// empty, even when an error occurs.
	Flush() error
}
