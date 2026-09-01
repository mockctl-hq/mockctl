package runtime

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

)

func TestInterceptorResponseWriter(t *testing.T) {
	t.Parallel()

	t.Run("Implicit 200 OK on Write", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		buf := AcquireTelemetryBuffer()
		defer buf.Decref()

		interceptor := &interceptorResponseWriter{
			ResponseWriter: rr,
			buffer:         buf,
			limit:          1024,
		}

		interceptor.Write([]byte("hello"))

		if !interceptor.wroteHeader {
			t.Errorf("Expected wroteHeader to be true")
		}
		if interceptor.statusCode != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", interceptor.statusCode)
		}
		if rr.Code != http.StatusOK {
			t.Errorf("Expected recorder code 200, got %d", rr.Code)
		}
	})

	t.Run("1xx Informational Responses Bypass Lock", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		buf := AcquireTelemetryBuffer()
		defer buf.Decref()

		interceptor := &interceptorResponseWriter{
			ResponseWriter: rr,
			buffer:         buf,
			limit:          1024,
		}

		interceptor.WriteHeader(http.StatusContinue) // 100
		
		if interceptor.wroteHeader {
			t.Errorf("Expected wroteHeader to be false after 1xx response")
		}

		interceptor.WriteHeader(http.StatusCreated) // 201

		if !interceptor.wroteHeader {
			t.Errorf("Expected wroteHeader to be true after 2xx response")
		}
		if interceptor.statusCode != http.StatusCreated {
			t.Errorf("Expected status code 201, got %d", interceptor.statusCode)
		}
	})
	
	t.Run("TeeWriter respects limit", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		buf := AcquireTelemetryBuffer()
		defer buf.Decref()

		interceptor := &interceptorResponseWriter{
			ResponseWriter: rr,
			buffer:         buf,
			limit:          5, // Only buffer first 5 bytes
		}

		interceptor.Write([]byte("hello world"))

		if buf.Buffer.String() != "hello" {
			t.Errorf("Expected buffer to contain 'hello', got '%s'", buf.Buffer.String())
		}
		if rr.Body.String() != "hello world" {
			t.Errorf("Expected recorder to contain full body 'hello world', got '%s'", rr.Body.String())
		}
	})
}

func TestTeeReadCloser(t *testing.T) {
	t.Parallel()

	t.Run("Reads and respects limit", func(t *testing.T) {
		t.Parallel()
		
		buf := AcquireTelemetryBuffer()
		defer buf.Decref()

		body := io.NopCloser(strings.NewReader("hello world"))
		
		trc := &teeReadCloser{
			original: body,
			buffer:   buf,
			limit:    5,
		}

		out, err := io.ReadAll(trc)
		if err != nil {
			t.Fatalf("Failed to read: %v", err)
		}

		if string(out) != "hello world" {
			t.Errorf("Expected to read 'hello world', got '%s'", string(out))
		}

		if buf.Buffer.String() != "hello" {
			t.Errorf("Expected telemetry buffer to contain 'hello', got '%s'", buf.Buffer.String())
		}
		
		if trc.read != 11 {
			t.Errorf("Expected bytes read to be 11, got %d", trc.read)
		}
	})

	t.Run("Close decreases ref count", func(t *testing.T) {
		t.Parallel()
		
		buf := AcquireTelemetryBuffer()
		body := io.NopCloser(bytes.NewReader(nil))
		
		trc := &teeReadCloser{
			original: body,
			buffer:   buf,
			limit:    1024,
		}

		err := trc.Close()
		if err != nil {
			t.Fatalf("Failed to close: %v", err)
		}
		
		if buf.refs.Load() != 0 {
			t.Errorf("Expected refs to be 0 after close, got %d", buf.refs.Load())
		}
	})
}
