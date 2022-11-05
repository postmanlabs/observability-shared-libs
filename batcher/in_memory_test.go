package batcher

import (
	"testing"
	"time"
)

type testBuffer struct {
	items    []int
	bufSize  int
	consumer func([]int)
}

var _ Buffer[int] = (*testBuffer)(nil)

func (buf *testBuffer) Add(val int) (bool, error) {
	buf.items = append(buf.items, val)
	return len(buf.items) >= buf.bufSize, nil
}

func (buf *testBuffer) Flush() error {
	buf.consumer(buf.items)
	buf.items = buf.items[:0]
	return nil
}

func TestInMemoryBatcherFlushOnBatchSize(t *testing.T) {
	count := 0
	procCount := 0
	proc := func(batch []int) {
		procCount += 1
		count += len(batch)
	}

	// Set the flush duration to a long time we test the flush happening due to
	// full buffer.
	bufSize := 10
	b := NewInMemory[int](
		&testBuffer{
			bufSize:  bufSize,
			consumer: proc,
		},
		999*time.Minute,
	)
	defer b.Close()

	expectedItemCount := 2 * bufSize
	for i := 0; i < expectedItemCount; i++ {
		if err := b.Add(i); err != nil {
			t.Fatalf("unexpected error on add: %v", err)
		}
	}

	if procCount != 2 {
		t.Errorf("expected 2 call to processor, got %d", procCount)
	}
	if count != expectedItemCount {
		t.Errorf("expected %d items, got %d", expectedItemCount, count)
	}
}

func TestInMemoryBatcherFlushOnClose(t *testing.T) {
	count := 0
	procCount := 0
	proc := func(batch []int) {
		procCount += 1
		count += len(batch)
	}

	// Set the flush duration to a long time we test the flush happening due to
	// full buffer.
	bufSize := 10
	b := NewInMemory[int](
		&testBuffer{
			bufSize:  bufSize,
			consumer: proc,
		},
		999*time.Minute,
	)

	// Make the number of items not divisible by bufSize.
	expectedItemCount := 2*bufSize + 1
	for i := 0; i < expectedItemCount; i++ {
		if err := b.Add(i); err != nil {
			t.Fatalf("unexpected error on add: %v", err)
		}
	}

	// Close should trigger a flush.
	b.Close()

	if procCount != 3 {
		t.Errorf("expected 3 call to processor, got %d", procCount)
	}
	if count != expectedItemCount {
		t.Errorf("expected %d items, got %d", expectedItemCount, count)
	}
}

func TestInMemoryBatcherPeriodicFlush(t *testing.T) {
	count := 0
	proc := func(batch []int) {
		count += len(batch)
	}

	// Set the buffer size to something huge so we're forced to rely on periodic
	// flush.
	b := NewInMemory[int](
		&testBuffer{
			bufSize:  999999,
			consumer: proc,
		},
		5*time.Millisecond,
	)
	defer b.Close()
	expectedItemCount := 21
	for i := 0; i < expectedItemCount; i++ {
		if err := b.Add(i); err != nil {
			t.Fatalf("unexpected error on add: %v", err)
		}
	}

	startTime := time.Now()
	for count != expectedItemCount {
		if time.Now().Sub(startTime) > 3*time.Second {
			t.Errorf("timed out after 3s")
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
