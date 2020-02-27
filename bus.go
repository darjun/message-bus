package messagebus

import (
	"fmt"
	"reflect"
	"sync"
)

// MessageBus implements publish/subscribe messaging paradigm
type MessageBus interface {
	// Publish publishes arguments to the given topic subscribers
	// Publish block only when the buffer of one of the subscribers is full.
	Publish(topic string, arg interface{})
	// Close unsubscribe all handlers from given topic
	Close(topic string)
	// Subscribe subscribes to the given topic
	Subscribe(topic string, fn callback)
	// Unsubscribe unsubscribe handler from the given topic
	Unsubscribe(topic string, fn callback) error
}

type callback func (interface{})

type handlersMap map[string][]*handler

type handler struct {
	callback callback
	rvalue   reflect.Value
	queue    chan interface{}
}

type messageBus struct {
	handlerQueueSize int
	mtx              sync.RWMutex
	handlers         handlersMap
}

func (b *messageBus) Publish(topic string, arg interface{}) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()

	if hs, ok := b.handlers[topic]; ok {
		for _, h := range hs {
			h.queue <- arg
		}
	}
}

func (b *messageBus) Subscribe(topic string, fn callback) {
	h := &handler{
		callback: fn,
		rvalue:   reflect.ValueOf(fn),
		queue:    make(chan interface{}, b.handlerQueueSize),
	}

	go func() {
		for arg := range h.queue {
			h.callback(arg)
		}
	}()

	b.mtx.Lock()
	defer b.mtx.Unlock()

	b.handlers[topic] = append(b.handlers[topic], h)
}

func (b *messageBus) Unsubscribe(topic string, fn callback) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	rvalue := reflect.ValueOf(fn)

	if _, ok := b.handlers[topic]; ok {
		for i, h := range b.handlers[topic] {
			if h.rvalue == rvalue {
				close(h.queue)

				b.handlers[topic] = append(b.handlers[topic][:i], b.handlers[topic][i+1:]...)
			}
		}

		return nil
	}

	return fmt.Errorf("topic %s doesn't exist", topic)
}

func (b *messageBus) Close(topic string) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	if _, ok := b.handlers[topic]; ok {
		for _, h := range b.handlers[topic] {
			close(h.queue)
		}

		delete(b.handlers, topic)

		return
	}
}

// New creates new MessageBus
// handlerQueueSize sets buffered channel length per subscriber
func New(handlerQueueSize int) MessageBus {
	if handlerQueueSize == 0 {
		panic("handlerQueueSize has to be greater then 0")
	}

	return &messageBus{
		handlerQueueSize: handlerQueueSize,
		handlers:         make(handlersMap),
	}
}
