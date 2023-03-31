package ipc

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"reflect"
	"sync"
	"time"
)

// InProcRPCBroker holds handlers of rpc service.
type InProcRPCBroker struct {
	sync.Mutex
	handlers map[string]reflect.Value
}

// Resolve get registered handler of the given name. return nil if not found.
func (r *InProcRPCBroker) Resolve(name string, args []reflect.Value) (reflect.Value, error) {
	return r.handlers[name], nil
}

func (r *InProcRPCBroker) register(name string, fn interface{}) {
	r.Lock()
	defer r.Unlock()

	r.handlers[name] = reflect.ValueOf(fn)
}

// InProcRPC implements RPC interface, including both server and client.
type InProcRPC struct {
	//network  string
	//endpoint string
	broker *InProcRPCBroker
}

// Expose exposes a service by associating a function handler.
func (r *InProcRPC) Expose(name string, fn interface{}) {
	r.broker.register(name, fn)
}

func (r *InProcRPC) ExposeV2(name string, handler CalleeHandler) error {
	// TODO: implement me.
	return nil
}

// Call calls an remote service identified by its name with the given args.
func (r *InProcRPC) Call(name string, args ...interface{}) (reflect.Value, error) {
	fn, ok := r.broker.handlers[name]
	if !ok {
		return reflect.Value{}, fmt.Errorf("rpc name %s not found", name)
	}

	funcType := fn.Type()
	arguments := make([]reflect.Value, len(args))
	for i, v := range args {
		if v == nil {
			arguments[i] = reflect.New(funcType.In(i)).Elem()
		} else {
			arguments[i] = reflect.ValueOf(v)
		}
	}

	ret := fn.Call(arguments)
	if len(ret) > 0 {
		return ret[0], nil
	}

	return reflect.Value{}, nil
}

func (r *InProcRPC) CallV2(name string, data []byte, timeout time.Duration) ([]byte, error) {
	_, ok := r.broker.handlers[name]
	if !ok {
		return nil, fmt.Errorf("rpc name %s not found", name)
	}

	// TODO: implement me.

	return nil, nil
}

func NewInProcRPC(conf *RPCConf) RPC {
	inst := &InProcRPC{
		//network:  "unix",
		//endpoint: env.GetExecFilePath() + "/rpc.sock",
	}

	inst.broker = &InProcRPCBroker{
		Mutex:    sync.Mutex{},
		handlers: make(map[string]reflect.Value),
	}

	log.Infoln("in-proc rpc service started")

	return inst
}
