package events

import (
	"sync"

	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/bluetuith-org/api-native/api/errorkinds"
	"github.com/olebedev/emitter"
)

// Event represents a general event.
type Event[T EventDataConstraint] struct {
	// ID holds the event ID.
	ID EventID

	// Action holds the corresponding action associated
	// with this event.
	Action EventAction

	// Data holds the actual event data.
	Data T
}

// EventDataConstraint defines a set of possible event data types.
type EventDataConstraint interface {
	errorkinds.GenericError |
		bluetooth.AdapterEventData | bluetooth.DeviceEventData |
		bluetooth.MediaEventData | bluetooth.FileTransferEventData
}

// eventEmitter represents an event emitter.
type eventEmitter struct {
	*emitter.Emitter

	mu sync.Mutex
}

// EventID represents a unique event ID.
type EventID int

// The different types of event IDs.
const (
	EventNone EventID = iota // The zero value for this type.
	EventError
	EventAdapter
	EventDevice
	EventFileTransfer
	EventMediaPlayer
)

// EventAction describes an action that is associated with an event.
type EventAction string

// The different types of event actions.
const (
	EventActionNone    EventAction = "none"
	EventActionUpdated EventAction = "updated"
	EventActionAdded   EventAction = "added"
	EventActionRemoved EventAction = "removed"
)

// eventNames holds names of different events.
var (
	eventNames = map[EventID]string{
		EventAdapter:      "adapter",
		EventDevice:       "device",
		EventFileTransfer: "file transfer",
		EventMediaPlayer:  "media player",
	}

	ev eventEmitter
)

// Name returns the name of the event ID.
func (e EventID) Name() string {
	return eventNames[e]
}

// Publish publishes the event to the event stream.
func (e Event[T]) Publish(data T) {
	event(func(em *emitter.Emitter) {
		e.Data = data
		em.Emit(e.ID.Name(), e)
	})
}

// Subscribe listens to the event stream and subscribes to the event.
func (e Event[T]) Subscribe() chan Event[T] {
	c := make(chan Event[T], 10)

	go func(ch chan Event[T]) {
		var emitterEvent <-chan emitter.Event

		event(func(em *emitter.Emitter) {
			emitterEvent = em.On(e.ID.Name())
		})

		for v := range emitterEvent {
			select {
			case c <- v.Args[0].(Event[T]):
			default:
			}
		}

		close(ch)
	}(c)

	return c
}

// AdapterEvent returns an event interface to publish/subscribe to adapter events.
func AdapterEvent(action ...EventAction) Event[bluetooth.AdapterEventData] {
	eventAction := EventActionNone
	if action != nil {
		eventAction = action[0]
	}

	return Event[bluetooth.AdapterEventData]{ID: EventAdapter, Action: eventAction}
}

// DeviceEvent returns an event interface to publish/subscribe to device events.
func DeviceEvent(action ...EventAction) Event[bluetooth.DeviceEventData] {
	eventAction := EventActionNone
	if action != nil {
		eventAction = action[0]
	}

	return Event[bluetooth.DeviceEventData]{ID: EventDevice, Action: eventAction}
}

// MediaEvent returns an event interface to publish/subscribe to media events.
func MediaEvent(action ...EventAction) Event[bluetooth.MediaEventData] {
	eventAction := EventActionNone
	if action != nil {
		eventAction = action[0]
	}

	return Event[bluetooth.MediaEventData]{ID: EventMediaPlayer, Action: eventAction}
}

// FileTransferEvent returns an event interface to publish/subscribe to file transfer events.
func FileTransferEvent(action ...EventAction) Event[bluetooth.FileTransferEventData] {
	eventAction := EventActionNone
	if action != nil {
		eventAction = action[0]
	}

	return Event[bluetooth.FileTransferEventData]{ID: EventFileTransfer, Action: eventAction}
}

// ErrorEvent returns an event interface to publish/subscribe to error events.
func ErrorEvent() Event[errorkinds.GenericError] {
	return Event[errorkinds.GenericError]{ID: EventError, Action: EventActionAdded}
}

// event initializes and accesses the event emitter.
func event(f func(e *emitter.Emitter)) {
	ev.mu.Lock()
	defer ev.mu.Unlock()

	if ev.Emitter == nil {
		ev.Emitter = emitter.New(100)
		ev.Use(EventNone.Name(), emitter.Sync)
	}

	f(ev.Emitter)
}
