package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/endec/jsonrpc2"
	"github.com/zourva/pareto/ipc"
	"time"
)

type RpcSpec int

const (
	JsonRpc RpcSpec = iota
	InvalidRpc
)

// Messaging defines RR and PS messaging patterns.
//
//	API Layer: RPC Caller/Callee functions
//	Specification Layer: Naming & Identification
//	Encapsulation Layer: Wire Protocol Message Serializer/Deserializer
//	Transportation Layer: Uni-cast Protocol
type Messaging interface {
	// Listen and Notify defines PS-mode messaging methods.
	Listen(topic string, fn ipc.Handler) error
	Notify(topic string, data []byte) error

	// ExposeMethod and CallMethod defines RR-mode messaging methods.
	ExposeMethod(name string, fn ipc.CalleeHandler) error
	CallMethod(name string, data []byte, to time.Duration) ([]byte, error)

	// RpcClient returns the built-in rpc client
	RpcClient() *jsonrpc2.Client

	// RpcServer returns the built-in rpc server
	RpcServer() *jsonrpc2.Server
}

// JsonRpcBinder implements jsonrpc2.ChannelBinder
// using service-framework messaging mechanism.
type JsonRpcBinder struct {
	service Service
}

func (b *JsonRpcBinder) Bind(channels map[string]jsonrpc2.ChannelHandler) error {
	for name, handler := range channels {
		if err := b.service.ExposeMethod(name, handler); err != nil {
			return err
		}
	}

	return nil
}

func NewJsonRpcBinder(service Service) *JsonRpcBinder {
	if service == nil {
		log.Fatalln("service must not be nil")
	}

	return &JsonRpcBinder{
		service: service,
	}
}

// JsonRpcInvoker implements jsonrpc2.Invoker
// using service-framework messaging mechanism.
type JsonRpcInvoker struct {
	service Service
}

func (i *JsonRpcInvoker) Call(channel string, data []byte, to time.Duration) ([]byte, error) {
	return i.service.CallMethod(channel, data, to)
}

func NewJsonRpcInvoker(service Service) *JsonRpcInvoker {
	if service == nil {
		log.Fatalln("service must not be nil")
	}

	return &JsonRpcInvoker{service: service}
}
