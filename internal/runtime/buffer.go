package runtime

import (
	"bytes"
	"sync"
	"sync/atomic"
)

// RefCountedBuffer provides a zero-allocation, thread-safe memory buffer
// that can be shared across multiple Fan-Out goroutines (EDL-056).
type RefCountedBuffer struct {
	Buffer *bytes.Buffer
	pool   *sync.Pool
	refs   atomic.Int32
}

// AddRef increments the reference count by n.
func (b *RefCountedBuffer) AddRef(n int32) {
	b.refs.Add(n)
}

// Decref decrements the reference count. When it reaches 0, the buffer is
// reset and returned to its originating sync.Pool to prevent memory leaks.
func (b *RefCountedBuffer) Decref() {
	if b.refs.Add(-1) == 0 {
		if b.Buffer != nil {
			b.Buffer.Reset()
		}
		if b.pool != nil {
			b.pool.Put(b)
		}
	}
}

// telemetryBufferPool is the global pool for all telemetry buffers (JSON payloads, Tee-Readers).
var telemetryBufferPool = sync.Pool{
	New: func() any {
		return &RefCountedBuffer{
			Buffer: bytes.NewBuffer(make([]byte, 0, 4096)), // Initial capacity of 4KB
		}
	},
}

// AcquireTelemetryBuffer returns a RefCountedBuffer from the sync.Pool
// with an initial reference count of 1. It MUST be released via Decref().
func AcquireTelemetryBuffer() *RefCountedBuffer {
	buf := telemetryBufferPool.Get().(*RefCountedBuffer)
	buf.pool = &telemetryBufferPool
	buf.refs.Store(1)
	return buf
}
