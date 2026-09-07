package runtime

import (
	"sync"
	"testing"
	"time"
)

func TestEventBroker(t *testing.T) {
	t.Parallel()

	t.Run("Concurrent Publish and Subscribe Dual-Mode", func(t *testing.T) {
		t.Parallel()
		broker := NewEventBroker()
		defer broker.Stop()

		jsonCh := broker.SubscribeJSON(FilterOptions{ProjectName: "test-proj"})
		rawCh := broker.SubscribeRaw(FilterOptions{ProjectName: "test-proj"})
		otherCh := broker.SubscribeJSON(FilterOptions{ProjectName: "other-proj"})

		// Let the subscriber loop register them
		time.Sleep(10 * time.Millisecond)

		var wg sync.WaitGroup
		concurrency := 100

		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func(id int) {
				defer wg.Done()
				buf := AcquireTelemetryBuffer()
				buf.Buffer.WriteString(`{"id": 1}`)
				event := &RequestEvent{ProjectNameField: "test-proj"}
				broker.Publish(event, buf)
			}(i)
		}

		wg.Wait()

		// Drain jsonCh
		jsonCount := 0
		rawCount := 0
		timeout := time.After(500 * time.Millisecond)

	drainLoop:
		for {
			select {
			case msg := <-jsonCh:
				jsonCount++
				if msg.Payload != nil {
					msg.Payload.Decref()
				}
				if jsonCount == concurrency && rawCount == concurrency {
					break drainLoop
				}
			case msg := <-rawCh:
				rawCount++
				msg.Decref()
				if jsonCount == concurrency && rawCount == concurrency {
					break drainLoop
				}
			case <-timeout:
				break drainLoop
			}
		}

		if jsonCount != concurrency {
			t.Errorf("Expected %d messages in jsonCh, got %d", concurrency, jsonCount)
		}
		if rawCount != concurrency {
			t.Errorf("Expected %d messages in rawCh, got %d", concurrency, rawCount)
		}

		// Check otherCh for cross-talk
		select {
		case <-otherCh:
			t.Errorf("otherCh should not receive messages for test-proj")
		default:
		}

		broker.UnsubscribeJSON(jsonCh)
		broker.UnsubscribeRaw(rawCh)
		broker.UnsubscribeJSON(otherCh)
	})

	t.Run("Ring-Buffer Eviction", func(t *testing.T) {
		t.Parallel()
		broker := NewEventBroker()
		defer broker.Stop()

		// Fill the subscriber channel (limit is 100)
		ch := broker.SubscribeJSON(FilterOptions{ProjectName: "test-proj"})
		time.Sleep(10 * time.Millisecond)

		for i := 0; i < 105; i++ {
			buf := AcquireTelemetryBuffer()
			buf.Buffer.WriteString("data")
			event := &RequestEvent{
				ProjectNameField: "test-proj",
				SequenceID:       string(rune(i)),
			}
			broker.Publish(event, buf)
		}

		time.Sleep(50 * time.Millisecond)

		count := 0
		timeout := time.After(100 * time.Millisecond)

		for {
			select {
			case msg := <-ch:
				count++
				if msg.Payload != nil {
					msg.Payload.Decref()
				}
			case <-timeout:
				goto EndDrain
			}
		}
	EndDrain:
		if count != 100 {
			t.Errorf("Expected exactly 100 messages to be retained after eviction, got %d", count)
		}

		broker.UnsubscribeJSON(ch)
	})

	t.Run("Stop explicitly drains inputChan", func(t *testing.T) {
		t.Parallel()
		broker := NewEventBroker()

		buf := AcquireTelemetryBuffer()
		event := &RequestEvent{ProjectNameField: "test-proj"}
		broker.Publish(event, buf)

		broker.Stop()

		if len(broker.inputChan) != 0 {
			t.Errorf("Expected inputChan to be drained, got %d", len(broker.inputChan))
		}
	})
}
