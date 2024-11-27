package eventbus

import (
	"sync"

	"github.com/cskr/pubsub/v2"
)

// nilEventHandler represents a disabled event handler.
type nilEventHandler struct{}

// defaultEventHandler represents an internal event handler.
type defaultEventHandler struct {
	*pubsub.PubSub[uint, any]
}

// EventPublisher represents an interface that provides an event publisher.
type EventPublisher interface {
	// Publish publishes an event to the event stream.
	Publish(id uint, name string, data any)
}

// EventSubscriber represents an interface that provides an event subscriber.
type EventSubscriber interface {
	// Subscribe subscribes to an event from the event stream.
	Subscribe(id uint, name string) SubscriberID
}

// EventHandler represents an interface that provides an event publisher and subscriber.
type EventHandler interface {
	EventPublisher
	EventSubscriber
}

// eventHandler represents the main event handler.
type eventHandler struct {
	p EventPublisher
	s EventSubscriber

	mu sync.RWMutex
}

var eventEmitter eventHandler

func init() {
	RegisterEventHandler(DefaultHandler())
}

// RegisterEventHandler registers the event handler interface.
func RegisterEventHandler(eh EventHandler) {
	if eh == nil {
		return
	}

	eventEmitter.mu.Lock()
	defer eventEmitter.mu.Unlock()

	eventEmitter.p = eh.(EventPublisher)
	eventEmitter.s = eh.(EventSubscriber)
}

// RegisterEventHandlers registers the event publisher and subscriber interfaces separately.
// To disable an EventPublisher or EventSubscriber, pass 'nil' as the parameter.
// For example: `RegisterEventHandlers(&eventPublisher{}, nil)` can be called to only register
// an event publisher.
func RegisterEventHandlers(p EventPublisher, s EventSubscriber) {
	eventEmitter.mu.Lock()
	defer eventEmitter.mu.Unlock()

	if p == nil {
		p = &nilEventHandler{}
	}
	if s == nil {
		s = &nilEventHandler{}
	}

	eventEmitter.p = p
	eventEmitter.s = s
}

// DisableEvents unregisters the event handler.
func DisableEvents() {
	RegisterEventHandler(&nilEventHandler{})
}

// Publish calls the registered publisher handler.
func Publish(id EventID, data any) {
	if id == nil {
		return
	}

	eventEmitter.mu.RLock()
	p := eventEmitter.p
	eventEmitter.mu.RUnlock()

	p.Publish(id.Value(), id.String(), data)
}

// Subscribe calls the registered subscriber handler.
func Subscribe(id EventID) SubscriberID {
	if id == nil {
		return (&nilEventHandler{}).Subscribe(0, "")
	}

	eventEmitter.mu.RLock()
	s := eventEmitter.s
	eventEmitter.mu.RUnlock()

	return s.Subscribe(id.Value(), id.String())
}

// DefaultHandler returns the default event handler.
func DefaultHandler() *defaultEventHandler {
	return &defaultEventHandler{PubSub: pubsub.New[uint, any](10)}
}

// NilHandler returns a disabled event handler.
func NilHandler() *nilEventHandler {
	return &nilEventHandler{}
}

// Publish publishes an event to the event stream.
func (d *defaultEventHandler) Publish(id uint, name string, data any) {
	d.TryPub(data, id)
}

// Subscribe subscribes to an event from the event stream.
func (d *defaultEventHandler) Subscribe(id uint, name string) SubscriberID {
	ch := d.Sub(id)
	return SubscriberID{
		C:      ch,
		active: true,
		unsub: func() {
			go d.Unsub(ch, id)
		},
	}
}

// Publish does not do anything.
func (n *nilEventHandler) Publish(uint, string, any) {
}

// Subscribe does not do anything.
func (n *nilEventHandler) Subscribe(uint, string) SubscriberID {
	ch := make(chan any)
	close(ch)
	return SubscriberID{C: ch}
}
