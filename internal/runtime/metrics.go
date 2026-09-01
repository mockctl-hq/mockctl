package runtime

import (
	"context"
	"encoding/json"
	"runtime"
	"sync"
	"time"
)

// MetricsCollector periodically gathers runtime memory and goroutine stats
// and broadcasts them via the EventBroker to power the dashboard metrics.
// CRITICAL: This MUST be a singleton to prevent GC paralysis from 1000s of projects.
type MetricsCollector struct {
	broker    EventPublisher
	frequency time.Duration
	wg        sync.WaitGroup
	cancel    context.CancelFunc
}

// NewMetricsCollector initializes the global metrics generator.
func NewMetricsCollector(broker EventPublisher) *MetricsCollector {
	return &MetricsCollector{
		broker:    broker,
		frequency: 2 * time.Second,
	}
}

// Start begins the background metrics generation loop.
func (m *MetricsCollector) Start(ctx context.Context) {
	ctx, m.cancel = context.WithCancel(ctx)
	m.wg.Add(1)
	go m.run(ctx)
}

// Stop initiates graceful shutdown.
func (m *MetricsCollector) Stop() {
	if m.cancel != nil {
		m.cancel()
		m.wg.Wait()
	}
}

func (m *MetricsCollector) run(ctx context.Context) {
	defer m.wg.Done()
	ticker := time.NewTicker(m.frequency)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Vestigial Deadlock Fix: Only generate metrics if someone is watching
			if m.broker.ActiveConnections() == 0 {
				continue
			}

			// Read runtime stats
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			numGoroutines := runtime.NumGoroutine()

			event := &MetricEvent{
				Timestamp:        time.Now(),
				ProjectNameField: "", // Global
				NumGoroutine:     numGoroutines,
				AllocBytes:       memStats.Alloc,
				SysBytes:         memStats.Sys,
			}

			// Synchronously encode and publish
			jsonBuf := AcquireTelemetryBuffer()
			jsonBuf.Buffer.Reset()

			if err := json.NewEncoder(jsonBuf.Buffer).Encode(event); err == nil {
				// JSON Newline Truncation
				b := jsonBuf.Buffer.Bytes()
				if len(b) > 0 && b[len(b)-1] == '\n' {
					jsonBuf.Buffer.Truncate(len(b) - 1)
				}
				m.broker.Publish(event, jsonBuf)
			} else {
				jsonBuf.Decref()
			}
		}
	}
}
