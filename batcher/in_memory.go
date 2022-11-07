package batcher

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

// A wrapper around a Buffer[Item] that manages thread-safety and flushing.
type InMemory[Item any] struct {
	buf Buffer[Item] // protected by mu

	signalClose                chan struct{} // closed when the batcher is closed
	signalPeriodicFlushStopped chan struct{} // closed when we stopped periodic flushing
	mu                         sync.Mutex
}

func NewInMemory[Item any](
	buf Buffer[Item],
	flushDuration time.Duration,
) *InMemory[Item] {
	m := &InMemory[Item]{
		buf:                        buf,
		signalClose:                make(chan struct{}),
		signalPeriodicFlushStopped: make(chan struct{}),
	}

	go func() {
		defer close(m.signalPeriodicFlushStopped)
		ticker := time.NewTicker(flushDuration)
		defer ticker.Stop()
		for {
			select {
			case <-m.signalClose:
				return
			case <-ticker.C:
				m.mu.Lock()
				m.buf.Flush()
				m.mu.Unlock()
			}
		}
	}()

	return m
}

func (m *InMemory[Item]) Add(items ...Item) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, item := range items {
		full, err := m.buf.Add(item)
		if err != nil {
			return errors.Wrap(err, "unable to add item to batch")
		}

		if full {
			m.buf.Flush()
		}
	}

	return nil
}

func (m *InMemory[_]) Close() {
	// Stop the periodic flusher.
	close(m.signalClose)
	<-m.signalPeriodicFlushStopped

	m.mu.Lock()
	defer m.mu.Unlock()
	m.buf.Flush()
}
