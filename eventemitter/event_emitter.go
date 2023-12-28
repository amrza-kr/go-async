package eventemitter

import (
	"sync"
)

const DEFAULT_MAX_LISTENERS uint = 10

type Listener[T any] struct {
	Name    string
	Once    bool
	Handler func(T)
}

type eventData[T any] struct {
	onceListenersCount int
	listeners          []Listener[T]
}

type EventEmitter[T any] interface {
	// On adds the listener function to the end of the listeners array for the event named eventName.
	// No checks are made to see if the listener has already been added.
	// Multiple calls passing the same combination of eventName and handler will result in the listener being added, and called, multiple times.
	// Returns a reference to the EventEmitter, so that calls can be chained.
	On(eventName string, listenerName string, handler func(T)) EventEmitter[T]

	// AddListener is an alias for emitter.On(eventName, listenerName, handler).
	AddListener(eventName string, listenerName string, handler func(T)) EventEmitter[T]

	// Off is an alias for emitter.RemoveListener().
	Off(eventName string, listenerName string) EventEmitter[T]

	// RemoveListener removes the specified listener with matching listenerName from the listeners array for the event named eventName.
	// RemoveListener() will remove, at most, one instance of a listener from the listeners array.
	// If any single listener has been added multiple times to the listeners array for the specified eventName, then RemoveListener() must be called multiple times to remove each instance.
	// Once an event is emitted, all listeners attached to it at the time of emitting are called in order.
	// This implies that any RemoveListener() or RemoveAllListeners() calls after emitting and before the last listener finishes execution will not remove them from Emit() in progress. Subsequent events behave as expected.
	// Returns a reference to the EventEmitter, so that calls can be chained.
	RemoveListener(eventName string, listenerName string) EventEmitter[T]

	// Emit synchronously calls each of the listeners registered for the event named eventName, in the order they were registered, passing the supplied arguments to each.
	// Returns true if the event had listeners, false otherwise.
	Emit(eventName string, data T) bool

	// EmitAsync calls each of the listeners registered for the event named eventName in a separate goroutine, in the order they were registered, passing the supplied arguments to each.
	// Returns true if the event had listeners, false otherwise.
	EmitAsync(eventName string, data T) bool

	// EventNames returns an array listing the events for which the emitter has registered listeners. The values in the array are strings.
	EventNames() []string

	// ListenersCount returns the number of listeners listening for the event named eventName.
	ListenersCount(eventName string) int

	// ListenerCount returns the number of listeners listening for the event named eventName with matching listenerName.
	ListenerCount(eventName string, listenertName string) int

	// Listeners returns a copy of the array of listeners for the event named eventName.
	Listeners(eventName string) []Listener[T]

	// Once adds a one-time listener function for the event named eventName. The
	// next time eventName is triggered, this listener is invoked and then removed.
	// Returns a reference to the EventEmitter, so that calls can be chained.
	Once(eventName string, listenerName string, handler func(T)) EventEmitter[T]

	// PrependListener adds the listener function to the beginning of the listeners array for the event named eventName.
	// No checks are made to see if the listener has already been added.
	// Multiple calls passing the same combination of eventName and handler will result in the listener being added, and called, multiple times.
	// Returns a reference to the EventEmitter, so that calls can be chained.
	PrependListener(eventName string, listenerName string, handler func(T)) EventEmitter[T]

	// PrependOnceListener adds a one-time listener function for the event named eventName to the beginning of the listeners array.
	// The next time eventName is triggered, this listener is removed, and then invoked.
	// Returns a reference to the EventEmitter, so that calls can be chained.
	PrependOnceListener(eventName string, listenerName string, handler func(T)) EventEmitter[T]

	// RemoveAllListeners removes all listeners.
	// It is bad practice to remove listeners added elsewhere in the code, particularly when the EventEmitter instance was created by some other package.
	// Returns a reference to the EventEmitter, so that calls can be chained.
	RemoveAllListeners() EventEmitter[T]

	// RemoveAllEventListeners removes all listeners of the specified eventName.
	// It is bad practice to remove listeners added elsewhere in the code, particularly when the EventEmitter instance was created by some other package.
	// Returns a reference to the EventEmitter, so that calls can be chained.
	RemoveAllEventListeners(eventName string) EventEmitter[T]
}

type eventEmitter[T any] struct {
	sync.RWMutex
	events map[string]eventData[T]
}

func NewEventEmitter[T any]() EventEmitter[T] {
	return &eventEmitter[T]{
		events: make(map[string]eventData[T]),
	}
}

func (ee *eventEmitter[T]) On(eventName string, listenerName string, handler func(T)) EventEmitter[T] {
	ee.addListener(eventName, listenerName, handler, false, false)

	return ee
}

func (ee *eventEmitter[T]) AddListener(eventName string, listenerName string, handler func(T)) EventEmitter[T] {
	return ee.On(eventName, listenerName, handler)
}

func (ee *eventEmitter[T]) Off(eventName string, listenerName string) EventEmitter[T] {
	return ee.RemoveListener(eventName, listenerName)
}

func (ee *eventEmitter[T]) RemoveListener(eventName string, listenerName string) EventEmitter[T] {
	ee.Lock()
	defer ee.Unlock()

	event := ee.events[eventName]
	if len(event.listeners) == 0 {
		return ee
	}

	for k, v := range event.listeners {
		if v.Name == listenerName {
			ee.events[eventName] = eventData[T]{
				onceListenersCount: event.onceListenersCount,
				listeners:          append(event.listeners[:k], event.listeners[k+1:]...),
			}
			break
		}
	}

	return ee
}

func (ee *eventEmitter[T]) Emit(eventName string, data T) bool {
	ee.Lock()
	defer ee.Unlock()

	if event := ee.events[eventName]; len(event.listeners) > 0 {
		for _, event := range event.listeners {
			event.Handler(data)
		}

		ee.events[eventName] = ee.permanentListeners(event)

		return true
	}

	return false
}

func (ee *eventEmitter[T]) EmitAsync(eventName string, data T) bool {
	ee.Lock()
	defer ee.Unlock()

	if event := ee.events[eventName]; len(event.listeners) > 0 {
		for _, event := range event.listeners {
			go event.Handler(data)
		}

		ee.events[eventName] = ee.permanentListeners(event)

		return true
	}

	return false
}

func (ee *eventEmitter[T]) EventNames() []string {
	ee.RLock()
	defer ee.RUnlock()

	eventName := make([]string, 0, len(ee.events))

	for k := range ee.events {
		eventName = append(eventName, k)
	}

	return eventName
}

func (ee *eventEmitter[T]) ListenersCount(eventName string) int {
	ee.RLock()
	defer ee.RUnlock()

	return len(ee.events[eventName].listeners)
}

func (ee *eventEmitter[T]) ListenerCount(eventName string, listenertName string) int {
	ee.RLock()
	defer ee.RUnlock()

	listeners := ee.events[eventName].listeners
	if len(listeners) == 0 {
		return 0
	}

	count := 0
	for _, v := range listeners {
		if v.Name == listenertName {
			count++
		}
	}

	return count
}

func (ee *eventEmitter[T]) Listeners(eventName string) []Listener[T] {
	ee.RLock()
	defer ee.RUnlock()

	listeners := ee.events[eventName].listeners
	if len(listeners) == 0 {
		return []Listener[T]{}
	}

	listenersCopy := make([]Listener[T], len(listeners))
	for k, v := range listeners {
		listenersCopy[k] = v
	}

	return listenersCopy
}

func (ee *eventEmitter[T]) Once(eventName string, listenerName string, handler func(T)) EventEmitter[T] {
	ee.addListener(eventName, listenerName, handler, true, false)

	return ee
}

func (ee *eventEmitter[T]) PrependListener(eventName string, listenerName string, handler func(T)) EventEmitter[T] {
	ee.addListener(eventName, listenerName, handler, false, true)

	return ee
}

func (ee *eventEmitter[T]) PrependOnceListener(eventName string, listenerName string, handler func(T)) EventEmitter[T] {
	ee.addListener(eventName, listenerName, handler, true, true)

	return ee
}

func (ee *eventEmitter[T]) RemoveAllListeners() EventEmitter[T] {
	ee.Lock()
	defer ee.Unlock()

	for k := range ee.events {
		ee.events[k] = eventData[T]{
			onceListenersCount: 0,
			listeners:          make([]Listener[T], 0),
		}
	}

	return ee
}

func (ee *eventEmitter[T]) RemoveAllEventListeners(eventName string) EventEmitter[T] {
	ee.Lock()
	defer ee.Unlock()

	if _, exists := ee.events[eventName]; exists {
		ee.events[eventName] = eventData[T]{
			onceListenersCount: 0,
			listeners:          make([]Listener[T], 0),
		}
	}

	return ee
}

// permanentListeners Would remove any Once listeners from the event listeners
func (ee *eventEmitter[T]) permanentListeners(event eventData[T]) eventData[T] {
	l := len(event.listeners) - event.onceListenersCount
	if l == 0 {
		return eventData[T]{
			onceListenersCount: 0,
			listeners:          []Listener[T]{},
		}
	}

	listeners := make([]Listener[T], 0, l)

	for _, v := range event.listeners {
		if !v.Once {
			listeners = append(listeners, v)
		}
	}

	return eventData[T]{
		onceListenersCount: 0,
		listeners:          listeners,
	}
}

func (ee *eventEmitter[T]) addListener(eventName, listenerName string, handler func(T), once, prepend bool) {
	if handler == nil {
		return
	}

	ee.Lock()
	defer ee.Unlock()

	onceListenerCount := 0
	if once {
		onceListenerCount = 1
	}

	if event, exists := ee.events[eventName]; exists {
		onceListenerCount = event.onceListenersCount
		if once {
			onceListenerCount++
		}

		listener := Listener[T]{
			Name:    listenerName,
			Once:    once,
			Handler: handler,
		}

		listeners := append(event.listeners, listener)
		if prepend {
			copy(listeners[1:], listeners)
			listeners[0] = listener
		}

		ee.events[eventName] = eventData[T]{
			onceListenersCount: onceListenerCount,
			listeners:          listeners,
		}
		return
	}

	ee.events[eventName] = eventData[T]{
		onceListenersCount: onceListenerCount,
		listeners: []Listener[T]{
			{
				Name:    listenerName,
				Once:    once,
				Handler: handler,
			},
		},
	}

	return
}
