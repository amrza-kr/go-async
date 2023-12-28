package eventemitter

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEventEmitter(t *testing.T) {
	emitter := NewEventEmitter[int]()
	assert.NotNil(t, emitter, "NewEventEmitter() should not return nil")
}

func TestAddAndEmitEvent(t *testing.T) {
	emitter := NewEventEmitter[int]()
	called := false
	emitter.On("testEvent", "testListener", func(data int) {
		called = true
	})

	assert.True(t, emitter.Emit("testEvent", 123), "Emit should report true when listeners are added")
	assert.True(t, called, "Listener should have been called")
}

func TestRemoveListener(t *testing.T) {
	emitter := NewEventEmitter[int]()
	called := false
	handler := func(data int) { called = true }
	emitter.On("testEvent", "testListener", handler)
	emitter.RemoveListener("testEvent", "testListener")

	assert.False(t, emitter.Emit("testEvent", 123), "Emit should report false after listener is removed")
	assert.False(t, called, "Removed listener should not have been called")
}

func TestOnceListener(t *testing.T) {
	emitter := NewEventEmitter[int]()
	callCount := 0
	emitter.Once("testEvent", "testListener", func(data int) {
		callCount++
	})

	assert.True(t, emitter.Emit("testEvent", 123), "Emit should report true for Once listener")
	assert.False(t, emitter.Emit("testEvent", 123), "Emit should report false after Once listener is called")
	assert.Equal(t, 1, callCount, "Once listener should be called exactly once")
}

func TestEmitAsync(t *testing.T) {
	emitter := NewEventEmitter[int]()
	var wg sync.WaitGroup
	called := false

	wg.Add(1)
	emitter.On("asyncEvent", "asyncListener", func(data int) {
		called = true
		wg.Done()
	})

	assert.True(t, emitter.EmitAsync("asyncEvent", 123), "EmitAsync should report true when listeners are added")
	wg.Wait()
	assert.True(t, called, "Async listener should have been called")
}

func TestPrependListener(t *testing.T) {
	emitter := NewEventEmitter[int]()
	order := []int{}

	emitter.On("orderEvent", "listener1", func(data int) {
		order = append(order, 1)
	})
	emitter.PrependListener("orderEvent", "listener2", func(data int) {
		order = append(order, 2)
	})

	emitter.Emit("orderEvent", 0)
	expectedOrder := []int{2, 1}
	assert.Equal(t, expectedOrder, order, "PrependListener should add listener to the beginning")
}

func TestRemoveAllListeners(t *testing.T) {
	emitter := NewEventEmitter[int]()
	called := false
	emitter.On("testEvent", "listener1", func(data int) {
		called = true
	})
	emitter.On("testEvent", "listener2", func(data int) {
		called = true
	})

	emitter.RemoveAllListeners()

	assert.False(t, emitter.Emit("testEvent", 123), "Emit should report false after listener is removed")
	assert.False(t, called, "Listeners should have been removed by RemoveAllListeners")
}

func TestRemoveAllEventListeners(t *testing.T) {
	emitter := NewEventEmitter[int]()
	called := false
	emitter.On("testEvent", "listener1", func(data int) {
		called = true
	})
	emitter.On("testEvent", "listener2", func(data int) {
		called = true
	})

	emitter.RemoveAllEventListeners("testEvent")

	assert.False(t, emitter.Emit("testEvent", 123), "Emit should report false after listener is removed")
	assert.False(t, called, "Event listeners should have been removed by RemoveAllEventListeners")
}

func TestListenerNameUniqueness(t *testing.T) {
	emitter := NewEventEmitter[int]()
	called1 := false
	called2 := false

	emitter.On("event1", "commonListener", func(data int) {
		called1 = true
	})
	emitter.On("event2", "commonListener", func(data int) {
		called2 = true
	})

	emitter.Emit("event1", 123)
	assert.True(t, called1, "Listener for event1 should have been called")
	assert.False(t, called2, "Listener for event2 should not have been called")
}

// ... [previous test functions]

func TestEventNames(t *testing.T) {
	emitter := NewEventEmitter[int]()
	emitter.On("event1", "listener1", func(data int) {})
	emitter.Once("event2", "listener2", func(data int) {})

	eventNames := emitter.EventNames()
	assert.ElementsMatch(t, []string{"event1", "event2"}, eventNames, "EventNames should return all event names")
}

func TestListeners(t *testing.T) {
	emitter := NewEventEmitter[int]()
	emitter.On("event1", "listener1", func(data int) {})
	emitter.Once("event1", "listener2", func(data int) {})
	emitter.On("event1", "listener3", func(data int) {})

	listeners := emitter.Listeners("event1")
	assert.Len(t, listeners, 3, "Listeners should return all listeners for an event")

	assert.Equal(t, "listener1", listeners[0].Name, "First listener should have the correct name")
	assert.Equal(t, "listener2", listeners[1].Name, "Second listener should have the correct name")
	assert.Equal(t, "listener3", listeners[2].Name, "Third listener should have the correct name")
	assert.False(t, listeners[0].Once, "First listener is permanent")
	assert.True(t, listeners[1].Once, "Second listener isn't permanent")
	assert.False(t, listeners[2].Once, "Third listener is permanent")

	emitter.Emit("event1", 5)

	listeners = emitter.Listeners("event1")
	assert.Len(t, listeners, 2, "Listeners should return all listeners for an event")

	assert.Equal(t, "listener1", listeners[0].Name, "First listener should have the correct name")
	assert.Equal(t, "listener3", listeners[1].Name, "Second listener should have the correct name")
	assert.False(t, listeners[0].Once, "First listener is permanent")
	assert.False(t, listeners[1].Once, "Second listener is permanent")
}

func TestListenersCount(t *testing.T) {
	emitter := NewEventEmitter[int]()
	emitter.On("event1", "listener1", func(data int) {})
	emitter.On("event1", "listener2", func(data int) {})

	count := emitter.ListenersCount("event1")
	assert.Equal(t, 2, count, "ListenersCount should return the correct number of listeners for an event")
}

// Additional tests for concurrent access can be added here.
