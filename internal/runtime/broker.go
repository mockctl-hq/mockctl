package runtime

import (
	"sync"
	"sync/atomic"
)

// EventMessage groups the routing metadata (Event interface) with its
// pre-encoded JSON payload (RefCountedBuffer).
type EventMessage struct {
	Event   Event
	Payload *RefCountedBuffer
}

// FilterOptions allows clients to specify which events they want to receive.
type FilterOptions struct {
	ProjectName string
}

type subscriber struct {
	ch      chan *EventMessage
	options FilterOptions
}

// EventBroker manages pub/sub for real-time telemetry via SSE.
type EventBroker struct {
	isStopped         atomic.Bool
	inputChan         chan *EventMessage
	addChan           chan *subscriber
	removeChan        chan chan *EventMessage
	quitChan          chan struct{}
	activeConnections atomic.Int32
	wg                sync.WaitGroup
}

// NewEventBroker initializes a concurrent-safe, lock-free event broker.
func NewEventBroker() *EventBroker {
	b := &EventBroker{
		// CRITICAL: 1000 buffer to prevent saturation dropping all traffic.
		inputChan:  make(chan *EventMessage, 1000),
		addChan:    make(chan *subscriber),
		removeChan: make(chan chan *EventMessage),
		quitChan:   make(chan struct{}),
	}
	b.wg.Add(1)
	go b.distributor()
	return b
}

// distributor is the lock-free background Goroutine responsible for Fan-Out routing.
func (b *EventBroker) distributor() {
	defer b.wg.Done()
	subscribers := make(map[*subscriber]struct{})

	// CRITICAL (Lock-Free Paradox Fix & Shutdown Memory Leak Fix):
	// Drain the input channel on shutdown and close all active subscribers.
	defer func() {
		for sub := range subscribers {
			close(sub.ch)
		}
		for {
			select {
			case msg := <-b.inputChan:
				if msg.Payload != nil {
					msg.Payload.Decref()
				}
			default:
				return
			}
		}
	}()

	for {
		select {
		case <-b.quitChan:
			return
		case sub := <-b.addChan:
			subscribers[sub] = struct{}{}
		case ch := <-b.removeChan:
			for sub := range subscribers {
				if sub.ch == ch {
					delete(subscribers, sub)
					close(sub.ch)
					break
				}
			}
		case msg := <-b.inputChan:
			// Filter and Fan-Out
			var activeSubs []*subscriber
			for sub := range subscribers {
				if sub.options.ProjectName == "" || sub.options.ProjectName == msg.Event.ProjectName() {
					activeSubs = append(activeSubs, sub)
				}
			}

			if len(activeSubs) == 0 {
				if msg.Payload != nil {
					msg.Payload.Decref()
				}
				continue
			}

			// AddRef for all active subscribers
			if msg.Payload != nil {
				msg.Payload.AddRef(int32(len(activeSubs)))
			}

			for _, sub := range activeSubs {
				select {
				case sub.ch <- msg:
					// Success
				default:
					// Buffer full, Ring-Buffer Eviction
					select {
					case oldMsg := <-sub.ch:
						if oldMsg.Payload != nil {
							oldMsg.Payload.Decref() // Prevent leak of dropped message
						}
					default:
					}
					// Try sending again
					select {
					case sub.ch <- msg:
					default:
						if msg.Payload != nil {
							msg.Payload.Decref()
						}
					}
				}
			}

			// Decref base count (held by HTTP handler)
			if msg.Payload != nil {
				msg.Payload.Decref()
			}
		}
	}
}

// Publish pushes a new event into the broker. It uses a default case to prevent blocking.
func (b *EventBroker) Publish(event Event, jsonBuf *RefCountedBuffer) {
	if b.isStopped.Load() {
		if jsonBuf != nil {
			jsonBuf.Decref()
		}
		return
	}

	msg := &EventMessage{
		Event:   event,
		Payload: jsonBuf,
	}

	select {
	case b.inputChan <- msg:
	default:
		// Drop on global saturation
		if jsonBuf != nil {
			jsonBuf.Decref()
		}
	}
}

// Subscribe opens a new event stream channel.
func (b *EventBroker) Subscribe(options FilterOptions) chan *EventMessage {
	if b.isStopped.Load() {
		return nil
	}

	sub := &subscriber{
		ch:      make(chan *EventMessage, 100), // OOM Limit per client
		options: options,
	}

	select {
	case <-b.quitChan:
		return nil
	case b.addChan <- sub:
		b.activeConnections.Add(1)
		return sub.ch
	}
}

// Unsubscribe closes the event stream channel and removes it from the broker.
func (b *EventBroker) Unsubscribe(ch chan *EventMessage) {
	if ch == nil || b.isStopped.Load() {
		return
	}

	select {
	case <-b.quitChan:
		return
	case b.removeChan <- ch:
		b.activeConnections.Add(-1)
	}
}

// ActiveConnections returns the number of active SSE clients.
func (b *EventBroker) ActiveConnections() int32 {
	return b.activeConnections.Load()
}

// Stop initiates a graceful shutdown of the EventBroker.
func (b *EventBroker) Stop() {
	if b.isStopped.Swap(true) {
		return
	}
	close(b.quitChan)
	b.wg.Wait()
}

// EventPublisher defines the interface for publishing events (narrow interface principle).
type EventPublisher interface {
	Publish(event Event, jsonBuf *RefCountedBuffer)
	ActiveConnections() int32
}

// EventSubscriber defines the interface for subscribing to events.
type EventSubscriber interface {
	Subscribe(options FilterOptions) chan *EventMessage
	Unsubscribe(ch chan *EventMessage)
}
