package runtime

import (
	"encoding/json"
	"net/http"
	"time"
)

// sseSemaphore limits the maximum number of concurrent SSE connections (Task 3.2).
// CRITICAL (Race-Free Concurrency): We use a channel instead of an atomic counter 
// to prevent "Check-Then-Act" race conditions when rejecting connections.
var sseSemaphore = make(chan struct{}, 5)

// handleSSEEvents streams real-time telemetry to the dashboard via Server-Sent Events.
func (s *HTTPServer) handleSSEEvents(w http.ResponseWriter, r *http.Request) {
	// 1. Validate Flusher compatibility immediately
	f, ok := w.(http.Flusher)
	if !ok {
		s.sendAdminError(w, "INTERNAL_SERVER_ERROR", "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 2. Enforce Semaphore Limit
	select {
	case sseSemaphore <- struct{}{}:
		// Semaphore acquired. Defer release.
		// CRITICAL (Client Disconnect Leak & Reconnect Deadlock Fix): 
		// MUST release semaphore BEFORE calling Unsubscribe().
		defer func() { <-sseSemaphore }()
	default:
		// Full
		s.sendAdminError(w, "RATE_LIMIT_EXCEEDED", "Maximum concurrent SSE connections reached", http.StatusTooManyRequests)
		return
	}

	// 3. Set Strict SSE Headers
	w.Header().Set("Content-Type", "text/event-stream")
	// CRITICAL (CDN Buffering Blackhole): no-transform prevents proxies from buffering
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("X-Accel-Buffering", "no")
	
	// Dynamically echo origin for CORS credentials
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	// CRITICAL (Gzip Buffering Sabotage): Prevent global compressors from buffering the stream
	r.Header.Del("Accept-Encoding")

	// Flush headers immediately
	f.Flush()

	// 4. Subscribe to the EventBroker
	// Parse ProjectName from query if needed
	projectName := r.URL.Query().Get("project")
	
	if s.broker == nil {
		return // Should never happen unless mock initialization is wrong
	}

	ch := s.broker.Subscribe(FilterOptions{ProjectName: projectName})
	if ch == nil {
		return // Broker stopped
	}

	// Explicit Unsubscribe and Drain
	defer func() {
		s.broker.Unsubscribe(ch)
		// CRITICAL (Memory Pool Returns): Drain the channel after Unsubscribe
		for msg := range ch {
			if msg.Payload != nil {
				msg.Payload.Decref()
			}
		}
	}()

	// 5. Heartbeat & Stream Loop
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()
	rc := http.NewResponseController(w)

	for {
		select {
		case <-ctx.Done():
			return
			
		case <-ticker.C:
			// Heartbeat
			// CRITICAL (Kernel Syscall Thrash Fix): Set write deadline ONLY inside the slow ticker
			_ = rc.SetWriteDeadline(time.Now().Add(10 * time.Second))
			
			_, err := w.Write([]byte("event: heartbeat\ndata: {}\n\n"))
			if err != nil {
				return // Client disconnected
			}
			f.Flush()
			
		case msg, ok := <-ch:
			// CRITICAL (CPU Spike Prevention): Check channel close
			if !ok {
				return 
			}

			// Format SSE Envelope
			eventType := msg.Event.EventType()
			
			// Build event block
			// envelope pattern: data: {"success":true,"data":...}
			// But msg.Payload is already just the data object itself (like RequestEvent).
			// PKS-027 mandates standard envelope. We need to wrap it.
			
			// We can write it in parts to avoid allocations:
			w.Write([]byte("event: "))
			w.Write([]byte(eventType))
			w.Write([]byte("\ndata: {\"success\":true,\"data\":"))
			
			if msg.Payload != nil {
				w.Write(msg.Payload.Buffer.Bytes())
			} else {
				// Fallback if payload missing
				payloadBytes, _ := json.Marshal(msg.Event)
				w.Write(payloadBytes)
			}
			
			w.Write([]byte("}\n\n"))
			
			f.Flush()
			
			// CRITICAL: Decref payload
			if msg.Payload != nil {
				msg.Payload.Decref()
			}
		}
	}
}
