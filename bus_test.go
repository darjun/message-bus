package messagebus

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	bus := New(runtime.NumCPU())

	if bus == nil {
		t.Fail()
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := New(runtime.NumCPU())

	handler := func(interface{}) {}

	bus.Subscribe("test", handler)

	if err := bus.Unsubscribe("test", handler); err != nil {
		fmt.Println(err)
		t.Fail()
	}

	if err := bus.Unsubscribe("non-existed", func(interface{}) {}); err == nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestClose(t *testing.T) {
	bus := New(runtime.NumCPU())

	handler := func(interface{}) {}

	bus.Subscribe("test", handler)

	original, ok := bus.(*messageBus)
	if !ok {
		fmt.Println("Could not cast message bus to its original type")
		t.Fail()
	}

	if 0 == len(original.handlers) {
		fmt.Println("Did not subscribed handler to topic")
		t.Fail()
	}

	bus.Close("test")

	if 0 != len(original.handlers) {
		fmt.Println("Did not unsubscribed handlers from topic")
		t.Fail()
	}
}

func TestPublish(t *testing.T) {
	bus := New(runtime.NumCPU())

	var wg sync.WaitGroup
	wg.Add(2)

	first := false
	second := false

	bus.Subscribe("topic", func(v interface{}) {
		defer wg.Done()
		first = v.(bool)
	})

	bus.Subscribe("topic", func(v interface{}) {
		defer wg.Done()
		second = v.(bool)
	})

	bus.Publish("topic", true)

	wg.Wait()

	if first == false || second == false {
		t.Fail()
	}
}

func TestHandleError(t *testing.T) {
	bus := New(runtime.NumCPU())
	bus.Subscribe("topic", func(arg interface{}) {
		out := arg.(chan error)
		out <- errors.New("I do throw error")
	})

	out := make(chan error)
	defer close(out)

	bus.Publish("topic", out)

	if <-out == nil {
		t.Fail()
	}
}
