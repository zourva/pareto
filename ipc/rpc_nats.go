package ipc

import (
	"errors"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"reflect"
	"sync"
	"time"
)

// NatsRPC implements the RPC interface, thus can be used
// as an RPCServer, an RPCClient or both.
// It creates a connection to a nats broker and use
// the connection object as the underlying carrier for RR messaging patterns.
type NatsRPC struct {
	*nats.Conn
	conf *RPCConf

	subs map[string][]*nats.Subscription // topic - array of nats.Subscription map
	lock sync.RWMutex                    // lock for the *nats.Subscription map
}

// Expose exposes a service by associating a handler.
func (r *NatsRPC) Expose(name string, fn interface{}) {
}

// ExposeV2 exposes a service by associating a handler.
func (r *NatsRPC) ExposeV2(name string, handler CalleeHandler) error {
	if handler == nil {
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

// CallV2 calls a remote service identified by its name with the given args and expects
// response data or error, in the time limited by timeout.
func (r *NatsRPC) CallV2(name string, data []byte, timeout time.Duration) ([]byte, error) {
	m, err := r.Request(name, data, timeout)
	if err != nil {
		log.Errorf("rpc caller call %s failed: %v", name, err)
		return nil, err
	}

	return m.Data, nil
}

// NewNatsRPC creates an RPC channel
// according to the conf.
//
// Returns nil and any error when failed.
func NewNatsRPC(conf *RPCConf) (RPC, error) {
	if len(conf.Name) == 0 {
		conf.Name = "nats-based rpc"
	}

	nc, err := nats.Connect(conf.Broker,
		nats.Name(conf.Name),
		nats.MaxReconnects(-1),
		nats.ClosedHandler(func(conn *nats.Conn) {
			id, _ := conn.GetClientID()
			log.Infof("nats client %d connection closed", id)
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			id, _ := conn.GetClientID()
			log.Infof("nats client %d disconnected: %v", id, err)
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			id, _ := conn.GetClientID()
			log.Infof("nats client %d reconnected", id)
		}))
	if err != nil {
		log.Errorln("connect to broker failed:", err)
		return nil, err
	}

	rpc := &NatsRPC{
		Conn: nc,
		conf: conf,
		subs: make(map[string][]*nats.Subscription),
		lock: sync.RWMutex{},
	}

	return rpc, nil
}
