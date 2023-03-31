package ipc

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"reflect"
	"sync"
)

// EventBus - box for handlers and callbacks.
// based on https://github.com/asaskevich/EventBus/blob/master/event_bus.go
type EventBus struct {
	handlers map[string][]*eventHandler
	lock     sync.RWMutex // a lock for the map
}

type eventHandler struct {
	callBack   reflect.Value
	flagOnce   bool
	sync.Mutex // lock for an event handler - useful for running async callbacks serially
}

func (bus *EventBus) check(fn interface{}) error {
	if !(reflect.TypeOf(fn).Kind() == reflect.Func) {
		return fmt.Errorf("%s is not of type reflect.Func", reflect.TypeOf(fn).Kind())
	}

	return nil
}

// Subscribe subscribes to a topic.
// Returns error if `fn` is not a function.
func (bus *EventBus) Subscribe(topic string, fn interface{}) error {
	if err := bus.check(fn); err != nil {
		return err
	}

	return bus.doSubscribe(topic, &eventHandler{
		callBack: reflect.ValueOf(fn), flagOnce: false, Mutex: sync.Mutex{},
	})
}

// SubscribeOnce subscribes to a topic once. Handler will be removed after executing.
// Returns error if `fn` is not a function.
func (bus *EventBus) SubscribeOnce(topic string, fn interface{}) error {
	if err := bus.check(fn); err != nil {
		return err
	}

	return bus.doSubscribe(topic, &eventHandler{
		callBack: reflect.ValueOf(fn), flagOnce: true, Mutex: sync.Mutex{},
	})
}

// Unsubscribe removes callback defined for a topic.
// Returns error if there are no callbacks subscribed to the topic.
func (bus *EventBus) Unsubscribe(topic string, handler interface{}) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	if _, ok := bus.handlers[topic]; ok && len(bus.handlers[topic]) > 0 {
		bus.removeHandler(topic, bus.findHandlerIdx(topic, reflect.ValueOf(handler)))
		return nil
	}

	return fmt.Errorf("topic %s doesn't exist", topic)
}

// Publish executes callback defined for a topic. Any additional argument will be transferred to the callback.
func (bus *EventBus) Publish(topic string, args ...interface{}) {
	bus.lock.Lock() // will unlock if handler is not found or always after makeArgs
	defer bus.lock.Unlock()

	if handlers, ok := bus.handlers[topic]; ok && 0 < len(handlers) {
		// Handlers slice may be changed by removeHandler and Unsubscribe during iteration,
		// so make a copy and iterate the copied slice.
		copyHandlers := make([]*eventHandler, len(handlers))
		copy(copyHandlers, handlers)

		//log.Debugln("number subscribers to publish:", len(copyHandlers))

		for i, handler := range copyHandlers {
			if handler.flagOnce {
				bus.removeHandler(topic, i)
			}

			log.Tracef("publish to %s with %v", topic, handler.callBack)
			//TODO: try goroutine pooling
			go bus.doPublish(handler, args...)
		}
	}
}

// doSubscribe handles the subscription logic and is utilized by the public Subscribe functions
func (bus *EventBus) doSubscribe(topic string, handler *eventHandler) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	//log.Tracef("subscribe to %s with %v", topic, handler.callBack)
	bus.handlers[topic] = append(bus.handlers[topic], handler)
	return nil
}

func (bus *EventBus) doPublish(handler *eventHandler, args ...interface{}) {
	passedArguments := bus.setupPublish(handler, args...)
	handler.callBack.Call(passedArguments)
}

func (bus *EventBus) removeHandler(topic string, idx int) {
	if _, ok := bus.handlers[topic]; !ok {
		return
	}
	l := len(bus.handlers[topic])

	if !(0 <= idx && idx < l) {
		return
	}

	copy(bus.handlers[topic][idx:], bus.handlers[topic][idx+1:])
	bus.handlers[topic][l-1] = nil // or the zero value of T
	bus.handlers[topic] = bus.handlers[topic][:l-1]
}

func (bus *EventBus) findHandlerIdx(topic string, callback reflect.Value) int {
	if _, ok := bus.handlers[topic]; ok {
		for idx, handler := range bus.handlers[topic] {
			if handler.callBack.Type() == callback.Type() &&
				handler.callBack.Pointer() == callback.Pointer() {
				return idx
			}
		}
	}
	return -1
}

func (bus *EventBus) setupPublish(callback *eventHandler, args ...interface{}) []reflect.Value {
	funcType := callback.callBack.Type()
	passedArguments := make([]reflect.Value, len(args))
	for i, v := range args {
		if v == nil {
			passedArguments[i] = reflect.New(funcType.In(i)).Elem()
		} else {
			passedArguments[i] = reflect.ValueOf(v)
		}
	}

	return passedArguments
}

func NewEventBus(conf *BusConf) Bus {
	b := &EventBus{
		handlers: make(map[string][]*eventHandler),
		lock:     sync.RWMutex{},
	}
	return Bus(b)
}
