package ipc

import (
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"reflect"
	"sync"
)

type descriptor struct {
	sub *nats.Subscription
	fn  interface{}
}

// NatsBus implements the Bus interface, thus can be used as a publisher, a subscriber or both.
// It creates a connection to a nats broker and use the connection object as
// the underlying carrier for bus messaging patterns.
type NatsBus struct {
	*nats.Conn
	conf *BusConf

	subs map[string][]*descriptor // topic - array of nats.Subscription map
	lock sync.RWMutex             // lock for the *nats.Subscription map
}

// NewNatsBus creates a Bus endpoint
// according to the conf.
//
// Returns nil and any error when failed.
func NewNatsBus(conf *BusConf) (Bus, error) {
	if len(conf.Name) == 0 {
		conf.Name = "nats-based bus"
	}

	nc, err := nats.Connect(conf.Broker,
		nats.Name(conf.Name),
		nats.MaxReconnects(-1),
		nats.ClosedHandler(func(conn *nats.Conn) {
			id, _ := conn.GetClientID()
			log.Infof("nats connection %d closed", id)
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			id, _ := conn.GetClientID()
			log.Infof("nats %d disconnected: %v", id, err)
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			id, _ := conn.GetClientID()
			log.Infof("nats reconnected, id = %d", id)
		}))
	if err != nil {
		log.Errorln("connect to broker failed:", err)
		return nil, err
	}

	bus := &NatsBus{
		Conn: nc,
		conf: conf,
		subs: make(map[string][]*descriptor),
		lock: sync.RWMutex{},
	}

	return bus, nil
}

func (n *NatsBus) Publish(topic string, data []byte) error {
	return n.Conn.Publish(topic, data)
}

func (n *NatsBus) Subscribe(topic string, fn Handler) error {
	s, err := n.Conn.Subscribe(topic, func(msg *nats.Msg) {
		//log.Debugln("recv subscribed:", msg.Data)
		fn(msg.Data)
	})

	if err != nil {
		return err
	}

	n.lock.Lock()
	defer n.lock.Unlock()

	//log.Tracef("subscribe to %s with %v", topic, fn)
	n.subs[topic] = append(n.subs[topic], &descriptor{s, fn})

	return nil
}

func (n *NatsBus) SubscribeOnce(topic string, fn Handler) error {
	// no need to save to n.subs since it will unsubscribe immediately
	var err error
	var s *nats.Subscription
	if s, err = n.Conn.Subscribe(topic, func(msg *nats.Msg) {
		//log.Debugln("recv subscribed:", msg.Data)
		fn(msg.Data)

		_ = s.Unsubscribe()
	}); err != nil {
		return err
	}

	return nil
}

// Unsubscribe
//
//	This method is goroutine-safe.
func (n *NatsBus) Unsubscribe(topic string, fn Handler) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	s, ok := n.subs[topic]
	if !ok {
		return nil
	}

	l := len(n.subs[topic])

	for i, desc := range s {
		ref1 := reflect.ValueOf(desc.fn)
		ref2 := reflect.ValueOf(fn)
		if ref1.Pointer() == ref2.Pointer() {
			// copy & move & overwrite & nullify
			copy(n.subs[topic][i:], n.subs[topic][i+1:])
			n.subs[topic][l-1] = nil // or the zero value of T
			n.subs[topic] = n.subs[topic][:l-1]
			break
		}
	}

	return nil
}
