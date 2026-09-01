package runtime

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
	"unicode/utf8"
	"unsafe"
)

const maxTelemetryBodySize = 1 * 1024 * 1024 // 1MB

// globalSequenceID generates monotonically increasing sequence IDs for drop detection.
var globalSequenceID atomic.Int64

// TelemetryMiddleware generates real-time HTTP events and pushes them to the EventBroker.
func TelemetryMiddleware(broker EventPublisher, projectName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			seqID := strconv.FormatInt(globalSequenceID.Add(1), 10)

			// Fast deep copy of headers
			reqHeaderCopy := r.Header.Clone()
			reqQueryCopy := r.URL.Query()

			// Prepare early metadata event
			baseEvent := &RequestEvent{
				SequenceID:       seqID,
				Timestamp:        startTime,
				ProjectNameField: projectName,
				HTTPMethod:       r.Method,
				Path:             r.URL.Path,
				RequestHeaders:   reqHeaderCopy,
				QueryParameters:  reqQueryCopy,
				// RequestSizeBytes is ContentLength; it might be -1 if unknown
				RequestSizeBytes: r.ContentLength,
			}

			// Extract ProjectName from request context if Engine populates it (or via Chi route)
			// Assuming there's a convention for it, left as "unknown" for now.

			// Immediately publish 'request_started' to prevent Slow-Reader Invisibility
			publishEvent(broker, baseEvent)

			// Setup Tee-Reader for Request Body
			reqBuffer := AcquireTelemetryBuffer()
			reqBuffer.Buffer.Reset()
			
			customReadCloser := &teeReadCloser{
				original: r.Body,
				buffer:   reqBuffer,
				limit:    maxTelemetryBodySize,
			}
			r.Body = customReadCloser

			// Setup Tee-Writer for Response
			resBuffer := AcquireTelemetryBuffer()
			resBuffer.Buffer.Reset()

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
				if r.Body != customReadCloser {
					customReadCloser.Close()
				}

				// Extract buffers as zero-allocation strings safely
				reqBytes := customReadCloser.buffer.Buffer.Bytes()
				resBytes := interceptor.buffer.Buffer.Bytes()

				// Binary Payload Destruction Fix
				if utf8.Valid(reqBytes) {
					baseEvent.RequestBody = unsafe.String(unsafe.SliceData(reqBytes), len(reqBytes))
				} else if len(reqBytes) > 0 {
					baseEvent.RequestBody = base64.StdEncoding.EncodeToString(reqBytes)
				}

				if utf8.Valid(resBytes) {
					baseEvent.ResponseBody = unsafe.String(unsafe.SliceData(resBytes), len(resBytes))
				} else if len(resBytes) > 0 {
					baseEvent.ResponseBody = base64.StdEncoding.EncodeToString(resBytes)
				}

				// Check truncations
				baseEvent.IsRequestBodyTruncated = customReadCloser.buffer.Buffer.Len() == maxTelemetryBodySize
				baseEvent.IsResponseBodyTruncated = interceptor.buffer.Buffer.Len() == maxTelemetryBodySize

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
				publishEvent(broker, baseEvent)

				// Cleanup Response buffer (since ResponseWriter has no Close method)
				resBuffer.Decref()
			}()

			// Execute the mock handler
			next.ServeHTTP(interceptor, r)
		})
	}
}

// publishEvent asynchronously encodes and pushes the event to the broker
func publishEvent(broker EventPublisher, event *RequestEvent) {
	if broker.ActiveConnections() == 0 {
		return
	}

	jsonBuf := AcquireTelemetryBuffer()
	jsonBuf.Buffer.Reset()
	
	// Synchronous encoding leveraging handler's goroutine
	err := json.NewEncoder(jsonBuf.Buffer).Encode(event)
	if err == nil {
		// JSON Newline Parse Error Fix
		b := jsonBuf.Buffer.Bytes()
		if len(b) > 0 && b[len(b)-1] == '\n' {
			jsonBuf.Buffer.Truncate(len(b) - 1)
		}
		
		// jsonBuf base count (ref=1) is passed to Publish
		// It will be decref'd by the broker when appropriate
		broker.Publish(event, jsonBuf)
	} else {
		jsonBuf.Decref()
	}
}
