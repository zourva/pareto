package jsonrpc2

import (
	"errors"
)

type Handler func(*RPCRequest) *RPCResponse

// Router defines underlying invoker
// of the JSON-RPC server side.
type Router interface {
	Register(method string, handler Handler)
	GetHandler(method string) (Handler, error)
}

func NewRouter() Router {
	return &router{
		handler: make(map[string]Handler),
	}
}

type router struct {
	handler map[string]Handler
}

func (r *router) GetHandler(method string) (Handler, error) {
	handler, ok := r.handler[method]
	if !ok {
		return nil, errors.New(ErrCodeString[ErrServerMethodNotFound])
	}

	return handler, nil
}

// Register registers a handler for the
// given method, and replace if exists.
//
// NOTE: not goroutine-safe.
func (r *router) Register(method string, handler Handler) {
	r.handler[method] = handler
}
