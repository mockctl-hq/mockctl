package runtime

import (
	"sync"
	"testing"
	"time"
)

func TestEventBroker(t *testing.T) {
	t.Parallel()

	t.Run("Concurrent Publish and Subscribe", func(t *testing.T) {
		t.Parallel()
		broker := NewEventBroker()
		defer broker.Stop()

		ch1 := broker.Subscribe(FilterOptions{ProjectName: "test-proj"})
		ch2 := broker.Subscribe(FilterOptions{ProjectName: "other-proj"})

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

		// Drain ch1 and count
		count := 0
		timeout := time.After(500 * time.Millisecond)

	drainLoop:
		for {
			select {
			case msg := <-ch1:
				count++
				if msg.Payload != nil {
					msg.Payload.Decref()
				}
				if count == concurrency {
					break drainLoop
				}
			case <-timeout:
				break drainLoop
			}
		}

		if count != concurrency {
			t.Errorf("Expected %d messages in ch1, got %d", concurrency, count)
		}

		// Check ch2 for cross-talk
		select {
		case <-ch2:
			t.Errorf("ch2 should not receive messages for test-proj")
		default:
		}

		broker.Unsubscribe(ch1)
		broker.Unsubscribe(ch2)
	})

	t.Run("Ring-Buffer Eviction", func(t *testing.T) {
		t.Parallel()
		broker := NewEventBroker()
		defer broker.Stop()

		// Fill the subscriber channel (limit is 100)
		ch := broker.Subscribe(FilterOptions{ProjectName: "test-proj"})
		time.Sleep(10 * time.Millisecond)

		// Publish 105 messages.
		// Subscriber channel capacity is 100.
		// The oldest 5 messages should be evicted.
		for i := 0; i < 105; i++ {
			buf := AcquireTelemetryBuffer()
			buf.Buffer.WriteString("data")
			event := &RequestEvent{
				ProjectNameField: "test-proj",
				SequenceID:       string(rune(i)),
			}
			broker.Publish(event, buf)
		}

		// Wait for distributor to process
		time.Sleep(50 * time.Millisecond)

		// Drain and verify we only have 100 messages, and they are the LATEST ones
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

		broker.Unsubscribe(ch)
	})

	t.Run("Stop explicitly drains inputChan", func(t *testing.T) {
		t.Parallel()
		broker := NewEventBroker()

		buf := AcquireTelemetryBuffer()
		event := &RequestEvent{ProjectNameField: "test-proj"}
		broker.Publish(event, buf)

		// Stop immediately, the background goroutine should drain the inputChan
		broker.Stop()

		// If it drained properly, inputChan should be empty
		if len(broker.inputChan) != 0 {
			t.Errorf("Expected inputChan to be drained, got %d", len(broker.inputChan))
		}
	})
}
