package jsonrpc2

import (
	"errors"
	log "github.com/sirupsen/logrus"
)

type Handler func(*RPCRequest) *RPCResponse

type RouterBinder interface {
	// Bind binds router to an underlying channel.
	Bind(func([]byte) ([]byte, error)) error
}

// Router defines underlying invoker
// of the JSON-RPC server side.
type Router interface {
	Register(method string, handler Handler)
	GetHandler(method string) (Handler, error)
	Binder() RouterBinder
}

func NewRouter(binder RouterBinder) Router {
	if binder == nil {
		log.Fatalln("binder must not be nil")
	}

	return &router{
		binder:  binder,
		handler: make(map[string]Handler),
	}
}

type router struct {
	binder  RouterBinder
	handler map[string]Handler
}

func (r *router) Binder() RouterBinder {
	return r.binder
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
