package ntop

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"reflect"
	"sync"
)

// RPCServer defines callee side of an RPC service.
type RPCServer interface {
	//Expose exposes an service by associating a function handler.
	Expose(name string, fn interface{})
}

// RPCClient defines caller side of an RPC service.
type RPCClient interface {
	//Call calls an remote service identified by its name with the given args.
	Call(name string, args ...interface{}) (reflect.Value, error)
}

// RPC implements both sides of RPC service.
type RPC interface {
	RPCServer
	RPCClient
}

// RPCMethod defines the identity of an rpc method.
type RPCMethod struct {
	Service string
	Object  string
	Method  string
}

// SerializedName returns the dotted name of the method, i.e.:
//  service.object.method
// e.g.:
//  webserver.cookie.get   #
func (r *RPCMethod) SerializedName() string {
	return fmt.Sprintf("%s.%s.%s", r.Service, r.Object, r.Method)
}

// Resolver holds handlers of rpc service.
type Resolver struct {
	sync.Mutex
	handlers map[string]reflect.Value
}

// Resolve get registered handler of the given name. return nil if not found.
func (r *Resolver) Resolve(name string, args []reflect.Value) (reflect.Value, error) {
	return r.handlers[name], nil
}

func (r *Resolver) register(name string, fn interface{}) {
	r.Lock()
	defer r.Unlock()

	r.handlers[name] = reflect.ValueOf(fn)
}

// RPCImpl implements RPC interface, including both server and client.
type RPCImpl struct {
	//network  string
	//endpoint string
	resolver *Resolver
}

// NewRPC creates a new RPC server and client.
func NewRPC() RPC {
	inst := &RPCImpl{
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

//Expose exposes an service by associating a function handler.
func (r *RPCImpl) Expose(name string, fn interface{}) {
	r.resolver.register(name, fn)
}

//Call calls an remote service identified by its name with the given args.
func (r *RPCImpl) Call(name string, args ...interface{}) (reflect.Value, error) {
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
