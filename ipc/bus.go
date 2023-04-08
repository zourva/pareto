package ipc

import log "github.com/sirupsen/logrus"

type BusType int

const (
	InterProcBus BusType = iota + 1
	InnerProcBus
)

type BusConf struct {
	//Name of the bus, optional but recommended.
	Name string

	//Type defines broker type managing the bus.
	Type BusType

	//Broker is the address used as a mediator-pattern endpoint.
	Broker string
}

type Handler func([]byte)

// Bus provides a message bus and expose API using pub/sub pattern.
type Bus interface {
	// Publish publishes data on the given topic.
	//  This method is goroutine-safe.
	Publish(topic string, data []byte) error

	// Subscribe subscribes data on the given topic by registering a callback.
	//  This method is goroutine-safe.
	Subscribe(topic string, fn Handler) error

	// SubscribeOnce calls Unsubscribe
	//  This method is goroutine-safe.
	SubscribeOnce(topic string, fn Handler) error

	// Unsubscribe removes handler registered for a topic.
	// Returns error if there are no handlers subscribed to the topic.
	//  This method is goroutine-safe.
	Unsubscribe(topic string, fn Handler) error
}

// NewBus returns a new Bus endpoint and connects itself to the given broker.
// If conf is nil, bus endpoint will not be created.
func NewBus(conf *BusConf) Bus {
	if conf == nil {
		conf = &BusConf{
			//Name:   "bus listener",
			Type:   InnerProcBus,
			Broker: "",
		}
	} else {
		if conf.Type == InterProcBus {
			// broker address must be provided
			if len(conf.Broker) == 0 {
				log.Errorln("broker address is necessary when the bus type is inter-proc")
				return nil
			}
		}
	}

	switch conf.Type {
	case InterProcBus:
		return NewNatsBus(conf)
	case InnerProcBus:
		fallthrough
	default:
		return NewEventBus(conf)
	}
}
