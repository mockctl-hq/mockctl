package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

const maxTelemetryBodySize = 1 * 1024 * 1024 // 1MB

// correlationCounter ensures unique UUID-like correlation IDs for requests
var correlationCounter atomic.Int64

// TelemetryMiddleware generates real-time HTTP events and pushes them to the EventBroker.
func TelemetryMiddleware(broker EventPublisher, projectName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			seqID := strconv.FormatInt(broker.GetNextSequenceID(), 10)
			corrID := strconv.FormatInt(correlationCounter.Add(1), 10)

			// Fast deep copy of headers
			reqHeaderCopy := r.Header.Clone()
			reqQueryCopy := r.URL.Query()

			// Prepare early metadata event
			baseEvent := &RequestEvent{
				SequenceID:       seqID,
				CorrelationID:    corrID,
				Timestamp:        startTime,
				ProjectNameField: projectName,
				HTTPMethod:       r.Method,
				Path:             r.URL.Path,
				RequestHeaders:   reqHeaderCopy,
				QueryParameters:  reqQueryCopy,
				// RequestSizeBytes is ContentLength; it might be -1 if unknown
				RequestSizeBytes: r.ContentLength,
			}

			// Immediately publish 'request_started' to prevent Slow-Reader Invisibility
			publishEvent(broker, baseEvent)

			// Setup Tee-Reader for Request Body
			reqBuffer := AcquireTelemetryBuffer()
			reqBuffer.Buffer.Reset()

			// Transfer ownership of ref to event
			baseEvent.RequestBody = reqBuffer

			customReadCloser := &teeReadCloser{
				original: r.Body,
				buffer:   reqBuffer,
				limit:    maxTelemetryBodySize,
			}
			r.Body = customReadCloser

			// Setup Tee-Writer for Response
			resBuffer := AcquireTelemetryBuffer()
			resBuffer.Buffer.Reset()

			// Transfer ownership of ref to event
			baseEvent.ResponseBody = resBuffer

			interceptor := &interceptorResponseWriter{
				ResponseWriter: w,
				buffer:         resBuffer,
				limit:          maxTelemetryBodySize,
			}

			// Setup Hijack Callback
			interceptor.hijackFn = func() {
				// Fire early hijack event to prevent blindspot
				hijackEvent := *baseEvent
				hijackEvent.IsHijacked = true
				hijackEvent.StatusCode = http.StatusSwitchingProtocols
				publishEvent(broker, &hijackEvent)
			}

			// Deferred Panic Recovery and Final Telemetry Publication
			defer func() {
				// Prevent Chaostic Goroutine Write-After-Free
				interceptor.closed.Store(true)
				customReadCloser.closed.Store(true)

				// Panic Recovery
				if rec := recover(); rec != nil {
					baseEvent.PanicError = fmt.Sprintf("%v", rec)
					if !interceptor.wroteHeader {
						interceptor.statusCode = http.StatusInternalServerError
					}
					// Must re-panic so Go's default handler prints the stack trace
					defer panic(rec)
				}

				// Ghost Status Fix: Implicit 200 if handler returned without writing body
				if !interceptor.wroteHeader && !interceptor.isHijacked {
					interceptor.statusCode = http.StatusOK
				}

				// Handler Override Leak Fix
				if r.Body != nil {
					if r.Body != customReadCloser {
						_ = customReadCloser.Close()
					}
				}

				// Check Base64 encodings
				reqBytes := reqBuffer.Buffer.Bytes()
				if len(reqBytes) > 0 && !utf8.Valid(reqBytes) {
					baseEvent.IsRequestBodyBase64 = true
				}

				resBytes := resBuffer.Buffer.Bytes()
				if len(resBytes) > 0 && !utf8.Valid(resBytes) {
					baseEvent.IsResponseBodyBase64 = true
				}

				// Check truncations
				baseEvent.IsRequestBodyTruncated = reqBuffer.Buffer.Len() == maxTelemetryBodySize
				baseEvent.IsResponseBodyTruncated = resBuffer.Buffer.Len() == maxTelemetryBodySize

				// Telemetry Blindspot Fix (Unread body)
				if r.ContentLength > 0 && customReadCloser.read == 0 {
					baseEvent.IsRequestBodyIgnored = true
				}

				baseEvent.StatusCode = interceptor.statusCode
				baseEvent.LatencyMs = time.Since(startTime).Milliseconds()
				baseEvent.IsHijacked = interceptor.isHijacked

				// Deep copy response headers
				baseEvent.ResponseHeaders = interceptor.Header().Clone()

				// Publish completed event
				// We don't publish if hijacked, since it was published early and would cause duplicate keys.
				if !interceptor.isHijacked {
					publishEvent(broker, baseEvent)
				} else {
					// We must manually decref because we skipped Publish
					baseEvent.Decref()
				}
			}()

			// Execute the mock handler
			next.ServeHTTP(interceptor, r)
		})
	}
}

// publishEvent asynchronously encodes and pushes the event to the broker
func publishEvent(broker EventPublisher, event *RequestEvent) {
	// If no internal connections exist, no one receives the raw event.
	// But BboltPersister should always be active.
	// Wait, we always publish because BboltPersister needs it, even if no Remote SSE.

	var jsonBuf *RefCountedBuffer

	// Only allocate and encode JSON if there are active SSE web clients
	if broker.ActiveRemoteConnections() > 0 {
		jsonBuf = AcquireTelemetryBuffer()
		jsonBuf.Buffer.Reset()

		encoder := json.NewEncoder(jsonBuf.Buffer)
		encoder.SetEscapeHTML(false) // CRITICAL: JSON HTML Escaping Mutilation

		err := encoder.Encode(event)
		if err == nil {
			// JSON Newline Parse Error Fix
			b := jsonBuf.Buffer.Bytes()
			if len(b) > 0 && b[len(b)-1] == '\n' {
				jsonBuf.Buffer.Truncate(len(b) - 1)
			}
		} else {
			jsonBuf.Decref()
			jsonBuf = nil
		}
	}

	// broker.Publish handles raw fan-out natively.
	// It accepts jsonBuf which may be nil.
	broker.Publish(event, jsonBuf)
}
