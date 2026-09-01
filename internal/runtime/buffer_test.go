package runtime

import (
	"sync"
	"testing"
)

func TestRefCountedBuffer(t *testing.T) {
	t.Parallel()

	t.Run("Acquire and Decref returns to pool", func(t *testing.T) {
		t.Parallel()
		buf := AcquireTelemetryBuffer()
		if buf.refs.Load() != 1 {
			t.Errorf("Expected refs to be 1, got %d", buf.refs.Load())
		}
		if buf.Buffer == nil {
			t.Errorf("Expected buffer to be initialized")
		}

		// Write some data
		buf.Buffer.WriteString("hello")

		// Decref should reset buffer and return to pool
		buf.Decref()

		if buf.refs.Load() != 0 {
			t.Errorf("Expected refs to be 0, got %d", buf.refs.Load())
		}

		// Re-acquire and check if it's clean
		buf2 := AcquireTelemetryBuffer()
		defer buf2.Decref()

		if buf2.Buffer.Len() != 0 {
			t.Errorf("Expected buffer to be empty after reset, got length %d", buf2.Buffer.Len())
		}
	})

	t.Run("Concurrent AddRef and Decref", func(t *testing.T) {
		t.Parallel()
		buf := AcquireTelemetryBuffer()
		
		var wg sync.WaitGroup
		concurrency := 100

		// Add multiple refs
		buf.AddRef(int32(concurrency))

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				buf.Decref()
			}()
		}
		
		wg.Wait()
		
		if buf.refs.Load() != 1 {
			t.Errorf("Expected refs to be 1 (base ref), got %d", buf.refs.Load())
		}
		
		buf.Decref() // Release base ref
	})
}
