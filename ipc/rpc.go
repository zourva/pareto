package ipc

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"reflect"
	"time"
)

// CalleeHandler abstracts the RPC server side universal handler.
type CalleeHandler func(data []byte) ([]byte, error)

// RPCServer defines callee side of an RPC service.
type RPCServer interface {
	//Expose register a method to rpc server by associating a function handler.
	Expose(name string, fn interface{})

	//ExposeV2 register a method to rpc server by associating a handler.
	//Serialization based style.
	ExposeV2(name string, handler CalleeHandler) error
}

// RPCClient defines caller side of an RPC service.
type RPCClient interface {
	//Call calls a remote method identified by its name with the given args.
	Call(name string, args ...interface{}) (reflect.Value, error)

	//CallV2 calls a remote service identified by its name with the given args
	//and expects response data or error, in the time limited by timeout.
	CallV2(name string, data []byte, timeout time.Duration) ([]byte, error)
}

// RPC implements both sides of RPC service.
type RPC interface {
	RPCServer
	RPCClient
}

type RpcType int

const (
	InterProcRpc RpcType = iota + 1
	InnerProcRpc
)

type RPCConf struct {
	//Name of the RPC stream, optional but recommended.
	Name string

	//Type defines the carrier used to exchange RPC messages.
	Type RpcType

	//Broker is the address used as a mediator-pattern endpoint.
	Broker string
}

// NewRPC creates a bidirectional RPC-pattern messager.
func NewRPC(conf *RPCConf) (RPC, error) {
	if conf == nil {
		conf = &RPCConf{
			Type:   InnerProcRpc,
			Broker: "",
		}
	} else {
		if conf.Type == InterProcRpc {
			// broker address must be provided
			if len(conf.Broker) == 0 {
				log.Errorln("broker address is necessary when rpc type is inter-proc")
				return nil, errors.New("broker address is invalid")
			}
		}
	}

	switch conf.Type {
	case InterProcRpc:
		return NewNatsRPC(conf)
	case InnerProcRpc:
		fallthrough
	default:
		return NewInProcRPC(conf)
	}
}
