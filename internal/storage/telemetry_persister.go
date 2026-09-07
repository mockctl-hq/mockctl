package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mockctl-hq/mockctl/internal/runtime"
	"go.etcd.io/bbolt"
)

// TelemetryPersister asynchronously writes telemetry events to the bbolt database.
// It subscribes to the EventBroker's raw stream and batches events to maximize SSD performance (EDL-050).
type TelemetryPersister struct {
	broker    runtime.EventSubscriber
	db        *bbolt.DB
	batchSize int
	flushFreq time.Duration
	wg        sync.WaitGroup
	cancel    context.CancelFunc
}

func NewTelemetryPersister(broker runtime.EventSubscriber, db *bbolt.DB) *TelemetryPersister {
	return &TelemetryPersister{
		broker:    broker,
		db:        db,
		batchSize: 100,             // Batch limit per transaction
		flushFreq: 2 * time.Second, // Yielding Backoff limit
	}
}

func (p *TelemetryPersister) Start(ctx context.Context) {
	ctx, p.cancel = context.WithCancel(ctx)
	p.wg.Add(1)
	go p.run(ctx)

	// Start Compaction Routine (Task 1.5)
	p.wg.Add(1)
	go p.compactionRoutine(ctx)
}

func (p *TelemetryPersister) Stop() {
	if p.cancel != nil {
		p.cancel()
		p.wg.Wait()
	}
}

func (p *TelemetryPersister) run(ctx context.Context) {
	defer p.wg.Done()

	// Subscribe to raw stream (no project filter, we want everything)
	ch := p.broker.SubscribeRaw(runtime.FilterOptions{})
	if ch == nil {
		return
	}
	defer p.broker.UnsubscribeRaw(ch)

	buffer := make([]runtime.Event, 0, p.batchSize)
	ticker := time.NewTicker(p.flushFreq)
	defer ticker.Stop()

	flush := func() {
		if len(buffer) == 0 {
			return
		}

		err := p.db.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket(bucketTelemetry)
			if bucket == nil {
				return fmt.Errorf("telemetry bucket missing")
			}

			for _, ev := range buffer {
				// Task 1.5: Serialize the struct
				payload, err := json.Marshal(ev)
				if err == nil {
					key := []byte(time.Now().UTC().Format(time.RFC3339Nano) + "_" + string(ev.EventType()))
					_ = bucket.Put(key, payload)
				}
				// Decref the raw buffers!
				ev.Decref()
			}
			return nil
		})

		if err != nil {
			// If update fails, we still must decref to prevent leaks!
			for _, ev := range buffer {
				ev.Decref()
			}
		}

		buffer = buffer[:0]
	}

	defer func() {
		flush()
		// Drain the channel to prevent Broker from leaking memory
		for ev := range ch {
			ev.Decref()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			buffer = append(buffer, ev)
			if len(buffer) >= p.batchSize {
				flush()
				// Yielding Backoff - Reset ticker so we don't immediately flush again
				ticker.Reset(p.flushFreq)
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (p *TelemetryPersister) compactionRoutine(ctx context.Context) {
	defer p.wg.Done()

	// Run compaction every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Compaction Routine (Task 1.5): Prevent fragmented mmap crashes
			// We delete telemetry older than 24 hours.
			_ = p.db.Update(func(tx *bbolt.Tx) error {
				bucket := tx.Bucket(bucketTelemetry)
				if bucket == nil {
					return nil
				}

				cutoff := []byte(time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339Nano))

				c := bucket.Cursor()
				for k, _ := c.First(); k != nil; k, _ = c.Next() {
					if string(k) < string(cutoff) {
						_ = c.Delete()
					} else {
						// Keys are chronological, so we can stop scanning
						break
					}
				}
				return nil
			})
		}
	}
}
