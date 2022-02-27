package ntop

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"reflect"
	"sync"
)

// RpcServer defines callee side of an RPC service.
type RpcServer interface {
	//Expose exposes an service by associating a function handler.
	Expose(name string, fn interface{})
}

// RpcClient defines caller side of an RPC service.
type RpcClient interface {
	//Call calls an remote service identified by its name with the given args.
	Call(name string, args ...interface{}) (reflect.Value, error)
}

// Rpc implements both sides of RPC service.
type Rpc interface {
	RpcServer
	RpcClient
}

// RpcMethod defines the identity of an rpc method.
type RpcMethod struct {
	Service string
	Object  string
	Method  string
}

// SerializedName returns the dotted name of the method, i.e.:
//  service.object.method
// e.g.:
//  webserver.cookie.get   #
func (r *RpcMethod) SerializedName() string {
	return fmt.Sprintf("%s.%s.%s", r.Service, r.Object, r.Method)
}

type Resolver struct {
	sync.Mutex
	handlers map[string]reflect.Value
}

func (r *Resolver) Resolve(name string, args []reflect.Value) (reflect.Value, error) {
	return r.handlers[name], nil
}

func (r *Resolver) register(name string, fn interface{}) {
	r.Lock()
	defer r.Unlock()

	r.handlers[name] = reflect.ValueOf(fn)
}

type RpcImpl struct {
	//network  string
	//endpoint string
	resolver *Resolver
}

func NewRpc() Rpc {
	inst := &RpcImpl{
		//network:  "unix",
		//endpoint: env.GetExecFilePath() + "/rpc.sock",
	}

	inst.resolver = &Resolver{
		Mutex:    sync.Mutex{},
		handlers: make(map[string]reflect.Value),
	}

	log.Infoln("rpc server & client started")

	return inst
}

func (r *RpcImpl) Expose(name string, fn interface{}) {
	r.resolver.register(name, fn)
}

func (r *RpcImpl) Call(name string, args ...interface{}) (reflect.Value, error) {
	//return r.client.SendV(name, args)
	fn, ok := r.resolver.handlers[name]
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
