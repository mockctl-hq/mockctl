package runtime

import (
	"time"
)

// EventType defines the type of SSE telemetry event.
type EventType string

const (
	EventTypeRequestLog  EventType = "request_log"
	EventTypeMetric      EventType = "metric"
	EventTypeDropWarning EventType = "drop_warning"
)

// Event is the base interface for all telemetry events broadcast by the EventBroker.
// This allows the broker to filter events by project name without inspecting the full payload.
type Event interface {
	EventType() EventType
	ProjectName() string
}

// RequestEvent represents a single HTTP request/response cycle intercepted by the TelemetryMiddleware.
// It is serialized to JSON and broadcast to connected SSE clients.
type RequestEvent struct {
	// SequenceID is monotonically increasing for drop detection.
	// CRITICAL: Serialized as string to prevent JavaScript precision loss (JS max int is 2^53 - 1).
	SequenceID string `json:"sequence_id,string"`
	
	Timestamp             time.Time           `json:"timestamp"`
	ProjectNameField      string              `json:"project_name"`
	HTTPMethod            string              `json:"http_method"`
	Path                  string              `json:"path"`
	StatusCode            int                 `json:"status_code"`
	LatencyMs             int64               `json:"latency_ms"`
	RequestSizeBytes      int64               `json:"request_size_bytes"`
	ResponseSizeBytes     int64               `json:"response_size_bytes"`
	MatchedRouteID        string              `json:"matched_route_id,omitempty"`
	RequestHeaders        map[string][]string `json:"request_headers,omitempty"`
	ResponseHeaders       map[string][]string `json:"response_headers,omitempty"`
	QueryParameters       map[string][]string `json:"query_parameters,omitempty"`
	RequestBody           string              `json:"request_body,omitempty"`
	ResponseBody          string              `json:"response_body,omitempty"`

	// Edge case flags
	IsHijacked               bool `json:"is_hijacked"`
	IsRequestBodyTruncated   bool `json:"is_request_body_truncated"`
	IsResponseBodyTruncated  bool `json:"is_response_body_truncated"`
	IsRequestBodyIgnored     bool `json:"is_request_body_ignored"`

	// Chaos & Panic metrics
	ChaosLatencyMs     int64  `json:"chaos_latency_ms,omitempty"`
	ChaosErrorInjected string `json:"chaos_error_injected,omitempty"`
	PanicError         string `json:"panic_error,omitempty"`
}

// EventType implements the Event interface.
func (e *RequestEvent) EventType() EventType {
	return EventTypeRequestLog
}

// ProjectName implements the Event interface.
func (e *RequestEvent) ProjectName() string {
	return e.ProjectNameField
}

// MetricEvent represents server health metrics periodically broadcast to the dashboard.
type MetricEvent struct {
	Timestamp        time.Time `json:"timestamp"`
	ProjectNameField string    `json:"project_name"` // Usually empty/global for server metrics
	NumGoroutine     int       `json:"num_goroutine"`
	AllocBytes       uint64    `json:"alloc_bytes"`
	SysBytes         uint64    `json:"sys_bytes"`
}

func (e *MetricEvent) EventType() EventType { return EventTypeMetric }
func (e *MetricEvent) ProjectName() string  { return e.ProjectNameField }

// DropWarningEvent is injected by the broker when a slow client drops messages.
type DropWarningEvent struct {
	Timestamp        time.Time `json:"timestamp"`
	ProjectNameField string    `json:"project_name"`
	DroppedCount     int64     `json:"dropped_count"`
}

func (e *DropWarningEvent) EventType() EventType { return EventTypeDropWarning }
func (e *DropWarningEvent) ProjectName() string  { return e.ProjectNameField }
