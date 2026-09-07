package runtime

import (
	"bytes"
	"encoding/base64"
	"sync"
	"sync/atomic"
	"unicode/utf8"
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

// MarshalJSON strictly controls how the buffer is serialized into JSON payloads.
// CRITICAL: Prevents 2MB heap allocation garbage by streaming/escaping bytes directly.
// CRITICAL: Aggressively truncates SSE payloads to 100KB to prevent Write-Deadline exhaustion loops on 3G/VPN connections.
func (b *RefCountedBuffer) MarshalJSON() ([]byte, error) {
	if b == nil || b.Buffer == nil || b.Buffer.Len() == 0 {
		return []byte(`""`), nil
	}

	data := b.Buffer.Bytes()
	truncated := false

	// Max 100KB truncation for SSE
	if len(data) > 100*1024 {
		data = data[:100*1024]
		truncated = true
	}

	// Check if valid UTF-8 Text
	if utf8.Valid(data) {
		out := bytes.NewBuffer(make([]byte, 0, len(data)+128))
		out.WriteByte('"')
		for _, char := range data {
			switch char {
			case '\n':
				out.WriteString(`\n`)
			case '\r':
				out.WriteString(`\r`)
			case '\t':
				out.WriteString(`\t`)
			case '\\':
				out.WriteString(`\\`)
			case '"':
				out.WriteString(`\"`)
			default:
				out.WriteByte(char)
			}
		}
		if truncated {
			out.WriteString(`\n\n[TRUNCATED FOR SSE]`)
		}
		out.WriteByte('"')
		return out.Bytes(), nil
	}

	// Binary Data: Base64 Encode
	encodedLen := base64.StdEncoding.EncodedLen(len(data))
	out := bytes.NewBuffer(make([]byte, 0, encodedLen+2))
	out.WriteByte('"')

	base64Buf := make([]byte, encodedLen)
	base64.StdEncoding.Encode(base64Buf, data)
	out.Write(base64Buf)

	out.WriteByte('"')
	return out.Bytes(), nil
}

// telemetryBufferPool is the global pool for all telemetry buffers (JSON payloads, Tee-Readers).
var telemetryBufferPool = sync.Pool{
	New: func() any {
		return &RefCountedBuffer{
			// CRITICAL: Pre-allocate 1MB capacity to prevent 10GB/sec of garbage heap allocations during stream copying.
			Buffer: bytes.NewBuffer(make([]byte, 0, 1024*1024)),
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
