package runtime

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

// teeReadCloser intercepts an HTTP request body up to a limit (1MB).
type teeReadCloser struct {
	original io.ReadCloser
	buffer   *RefCountedBuffer
	limit    int64
	read     int64
}

func (t *teeReadCloser) Read(p []byte) (n int, err error) {
	n, err = t.original.Read(p)
	if n > 0 {
		t.read += int64(n)
		if t.buffer.Buffer.Len() < int(t.limit) {
			// Only write up to the limit
			spaceLeft := int(t.limit) - t.buffer.Buffer.Len()
			if n > spaceLeft {
				t.buffer.Buffer.Write(p[:spaceLeft])
			} else {
				t.buffer.Buffer.Write(p[:n])
			}
		}
	}
	return n, err
}

func (t *teeReadCloser) Close() error {
	t.buffer.Decref()
	return t.original.Close()
}

// interceptorResponseWriter wraps http.ResponseWriter to capture telemetry.
type interceptorResponseWriter struct {
	http.ResponseWriter
	buffer      *RefCountedBuffer
	limit       int64
	wroteHeader bool
	statusCode  int
	isHijacked  bool
	isZeroCopy  bool
	hijackFn    func() // Callback to publish event immediately on hijack
}

func (w *interceptorResponseWriter) WriteHeader(statusCode int) {
	if w.isHijacked {
		return
	}
	if statusCode >= 100 && statusCode < 200 {
		// 1xx Informational Lockout Bug: Bypass state lock
		w.ResponseWriter.WriteHeader(statusCode)
		return
	}
	if w.wroteHeader {
		return
	}
	w.statusCode = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *interceptorResponseWriter) Write(b []byte) (int, error) {
	if w.isHijacked {
		return 0, http.ErrHijacked
	}
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	if w.buffer.Buffer.Len() < int(w.limit) {
		spaceLeft := int(w.limit) - w.buffer.Buffer.Len()
		if len(b) > spaceLeft {
			w.buffer.Buffer.Write(b[:spaceLeft])
		} else {
			w.buffer.Buffer.Write(b)
		}
	}

	return w.ResponseWriter.Write(b)
}

// Flush implements http.Flusher
func (w *interceptorResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements http.Hijacker
func (w *interceptorResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	
	conn, rw, err := h.Hijack()
	if err == nil {
		w.isHijacked = true
		w.statusCode = http.StatusSwitchingProtocols // 101
		if w.hijackFn != nil {
			w.hijackFn()
		}
	}
	return conn, rw, err
}

// ReadFrom implements io.ReaderFrom to support OS sendfile (Zero-Copy Paradox)
func (w *interceptorResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	w.isZeroCopy = true
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}
	// Fallback to standard io.Copy without bypassing our Write method
	return io.Copy(w, r)
}
