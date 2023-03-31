package ipc

import (
	"errors"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"reflect"
	"sync"
	"time"
)

// NatsBus implements the Bus interface, thus can be used as a publisher, a subscriber or both.
// It creates a connection to a nats broker and use the connection object as
// the underlying carrier for bus messaging patterns.
type NatsRPC struct {
	*nats.Conn
	conf *RPCConf

	subs map[string][]*nats.Subscription // topic - array of nats.Subscription map
	lock sync.RWMutex                    // lock for the *nats.Subscription map
}

// Expose exposes a service by associating a handler.
func (r *NatsRPC) Expose(name string, fn interface{}) {
}

// Expose exposes a service by associating a handler.
func (r *NatsRPC) ExposeV2(name string, handler CalleeHandler) error {
	if handler != nil {
		return errors.New("handler must not be nil")
	}

	_, err := r.Subscribe(name, func(msg *nats.Msg) {
		// invoke handler with msg.Data
		if rsp, e := handler(msg.Data); e != nil {
			log.Errorf("invoke rpc callee handler for %s failed: %v", name, e)
		} else {
			if e = msg.Respond(rsp); e != nil {
				log.Errorf("respond to rpc caller %s failed: %v", name, e)
				return
			}
		}
	})

	if err != nil {
		log.Errorln("expose method failed:", err)
		return err
	}

	return nil
}

// Call calls a remote service identified by its name with the given args.
func (r *NatsRPC) Call(name string, args ...interface{}) (reflect.Value, error) {
	return reflect.Value{}, nil
}

// Call calls a remote service identified by its name with the given args and expects
// response data or error, in the time limited by timeout.
func (r *NatsRPC) CallV2(name string, data []byte, timeout time.Duration) ([]byte, error) {
	m, err := r.Request(name, data, timeout)
	if err != nil {
		log.Errorf("rpc caller call %s failed: %v", name, err)
		return nil, err
	}

	return m.Data, nil
}

func NewNatsRPC(conf *RPCConf) RPC {
	if len(conf.Name) == 0 {
		conf.Name = "nats-based rpc"
	}

	nc, err := nats.Connect(conf.Broker, nats.Name(conf.Name))
	if err != nil {
		log.Errorf("connect to broker %s failed: %v\n", conf.Broker, err)
		return nil
	}

	rpc := &NatsRPC{
		Conn: nc,
		conf: conf,
		subs: make(map[string][]*nats.Subscription),
		lock: sync.RWMutex{},
	}

	return rpc
}
