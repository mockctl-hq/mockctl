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

type jsonSubscriber struct {
	ch      chan *EventMessage
	options FilterOptions
}

type rawSubscriber struct {
	ch      chan Event
	options FilterOptions
}

// EventBroker manages pub/sub for real-time telemetry.
// It supports Dual-Mode Fan-Out for embedded (raw) and remote (JSON) clients.
type EventBroker struct {
	isStopped           atomic.Bool
	inputChan           chan *EventMessage
	addJsonChan         chan *jsonSubscriber
	removeJsonChan      chan chan *EventMessage
	addRawChan          chan *rawSubscriber
	removeRawChan       chan chan Event
	quitChan            chan struct{}
	remoteConnections   atomic.Int32
	internalConnections atomic.Int32
	sequenceCounter     atomic.Int64 // Embedded sequence ID for requests
	wg                  sync.WaitGroup
}

// NewEventBroker initializes a concurrent-safe, lock-free event broker.
func NewEventBroker() *EventBroker {
	b := &EventBroker{
		inputChan:      make(chan *EventMessage, 1000), // CRITICAL: Heavy buffer for global firehose
		addJsonChan:    make(chan *jsonSubscriber, 100),
		removeJsonChan: make(chan chan *EventMessage, 100),
		addRawChan:     make(chan *rawSubscriber, 100),
		removeRawChan:  make(chan chan Event, 100),
		quitChan:       make(chan struct{}),
	}
	b.wg.Add(1)
	go b.distributor()
	return b
}

func (b *EventBroker) distributor() {
	defer b.wg.Done()

	jsonSubs := make(map[*jsonSubscriber]struct{})
	rawSubs := make(map[*rawSubscriber]struct{})

	// CRITICAL (Lock-Free Paradox Fix & Shutdown Memory Leak Fix):
	defer func() {
		for sub := range jsonSubs {
			close(sub.ch)
		}
		for sub := range rawSubs {
			close(sub.ch)
		}
		// Drain the input channel on shutdown
		for {
			select {
			case msg := <-b.inputChan:
				if msg.Payload != nil {
					msg.Payload.Decref()
				}
				msg.Event.Decref()
			default:
				return
			}
		}
	}()

	for {
		select {
		case <-b.quitChan:
			return

		case sub := <-b.addJsonChan:
			jsonSubs[sub] = struct{}{}

		case ch := <-b.removeJsonChan:
			for sub := range jsonSubs {
				if sub.ch == ch {
					delete(jsonSubs, sub)
					close(sub.ch)
					break
				}
			}

		case sub := <-b.addRawChan:
			rawSubs[sub] = struct{}{}

		case ch := <-b.removeRawChan:
			for sub := range rawSubs {
				if sub.ch == ch {
					delete(rawSubs, sub)
					close(sub.ch)
					break
				}
			}

		case msg := <-b.inputChan:
			// Process JSON Subscribers (Remote HTTP)
			var activeJson []*jsonSubscriber
			for sub := range jsonSubs {
				if sub.options.ProjectName == "" || sub.options.ProjectName == msg.Event.ProjectName() {
					activeJson = append(activeJson, sub)
				}
			}

			if len(activeJson) > 0 && msg.Payload != nil {
				msg.Payload.AddRef(int32(len(activeJson))) // #nosec G115

				for _, sub := range activeJson {
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
			}

			// Decref JSON base count (held by HTTP handler)
			if msg.Payload != nil {
				msg.Payload.Decref()
			}

			// Process Raw Subscribers (Embedded TUI / BBolt)
			var activeRaw []*rawSubscriber
			for sub := range rawSubs {
				if sub.options.ProjectName == "" || sub.options.ProjectName == msg.Event.ProjectName() {
					activeRaw = append(activeRaw, sub)
				}
			}

			if len(activeRaw) > 0 {
				for _, sub := range activeRaw {
					// CloneRaw creates a shallow copy of the Event struct and increments buffer refs.
					clonedEvent := msg.Event.CloneRaw()

					select {
					case sub.ch <- clonedEvent:
					default:
						// Ring buffer eviction
						select {
						case oldEvent := <-sub.ch:
							oldEvent.Decref()
						default:
						}

						select {
						case sub.ch <- clonedEvent:
						default:
							clonedEvent.Decref()
						}
					}
				}
			}

			// Decref raw event base count (held by HTTP handler)
			msg.Event.Decref()
		}
	}
}

// Publish pushes a new event into the broker. It uses a default case to prevent blocking.
func (b *EventBroker) Publish(event Event, jsonBuf *RefCountedBuffer) {
	if b.isStopped.Load() {
		if jsonBuf != nil {
			jsonBuf.Decref()
		}
		event.Decref()
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
		event.Decref()
	}
}

// SubscribeJSON opens a new event stream channel for remote SSE clients.
func (b *EventBroker) SubscribeJSON(options FilterOptions) chan *EventMessage {
	if b.isStopped.Load() {
		return nil
	}
	sub := &jsonSubscriber{ch: make(chan *EventMessage, 100), options: options}
	select {
	case <-b.quitChan:
		return nil
	case b.addJsonChan <- sub:
		b.remoteConnections.Add(1)
		return sub.ch
	}
}

// UnsubscribeJSON closes the event stream channel and removes it from the broker.
func (b *EventBroker) UnsubscribeJSON(ch chan *EventMessage) {
	if ch == nil || b.isStopped.Load() {
		return
	}
	select {
	case <-b.quitChan:
		return
	case b.removeJsonChan <- ch:
		b.remoteConnections.Add(-1)
	}
}

// SubscribeRaw opens a new event stream channel for embedded clients (TUI, BBolt).
func (b *EventBroker) SubscribeRaw(options FilterOptions) chan Event {
	if b.isStopped.Load() {
		return nil
	}
	sub := &rawSubscriber{ch: make(chan Event, 100), options: options}
	select {
	case <-b.quitChan:
		return nil
	case b.addRawChan <- sub:
		b.internalConnections.Add(1)
		return sub.ch
	}
}

// UnsubscribeRaw closes the raw stream channel.
func (b *EventBroker) UnsubscribeRaw(ch chan Event) {
	if ch == nil || b.isStopped.Load() {
		return
	}
	select {
	case <-b.quitChan:
		return
	case b.removeRawChan <- ch:
		b.internalConnections.Add(-1)
	}
}

func (b *EventBroker) ActiveRemoteConnections() int32   { return b.remoteConnections.Load() }
func (b *EventBroker) ActiveInternalConnections() int32 { return b.internalConnections.Load() }
func (b *EventBroker) GetNextSequenceID() int64         { return b.sequenceCounter.Add(1) }

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
	ActiveRemoteConnections() int32
	ActiveInternalConnections() int32
	GetNextSequenceID() int64
}

// EventSubscriber defines the interface for subscribing to events.
type EventSubscriber interface {
	SubscribeJSON(options FilterOptions) chan *EventMessage
	UnsubscribeJSON(ch chan *EventMessage)
	SubscribeRaw(options FilterOptions) chan Event
	UnsubscribeRaw(ch chan Event)
}
