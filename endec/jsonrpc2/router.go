package jsonrpc2

import (
	"errors"
	log "github.com/sirupsen/logrus"
)

// Handler defines protocol layer handler for rpc.
type Handler func(*RPCRequest) *RPCResponse

// ChannelHandler defines underlying channel routing impl for rpc.
type ChannelHandler = func(req []byte) (rsp []byte, err error)

// Router defines underlying invoker
// of the JSON-RPC server side.
type Router interface {
	AllChannels() map[string]ChannelHandler
	AddChannel(name string)
	AddChannelWithHandler(name string, handler ChannelHandler) error
	RemoveChannel(name string)

	//Register registers handler on a routing channel.
	Register(channel, method string, handler Handler)

	//GetHandler returns handler on a routing channel.
	GetHandler(channel, method string) (Handler, error)

	Run() error
}

func NewRouter(binder ChannelBinder) Router {
	if binder == nil {
		log.Fatalln("binder must not be nil")
		return nil
	}

	return &router{
		channelBinder:   binder,
		channelHandlers: make(map[string]ChannelHandler),
		methodHandlers:  make(map[string]map[string]Handler),
	}
}

// router manages routing endpoints based
// /service/channel/method hierarchy path.
type router struct {
	channelHandlers map[string]ChannelHandler
	methodHandlers  map[string]map[string]Handler
	channelBinder   ChannelBinder
}

//func (r *router) dispatch(channel string, reqBuf []byte) ([]byte, error) {
//	req, err := ParseRequest(reqBuf)
//	if err != nil {
//		return nil, err
//	}
//
//	handler, err2 := r.GetHandler(channel, req.Method)
//	if err2 != nil {
//		return nil, err2
//	}
//
//	rsp := handler(req)
//
//	return rsp.Marshal()
//}

func (r *router) Run() error {
	return r.channelBinder.Bind(r.channelHandlers)
}

// AllChannels
//
// NOTE: Not goroutine safe.
func (r *router) AllChannels() map[string]ChannelHandler {
	return r.channelHandlers
}

// AddChannel adds a routing channel(routing group) to the
// router and the default handler is used if handler is nil.
//
// NOTE: Not goroutine safe.
func (r *router) AddChannel(channel string) {
	r.channelHandlers[channel] = func(reqBuf []byte) ([]byte, error) {
		req, err := ParseRequest(reqBuf)
		if err != nil {
			return nil, err
		}

		handler, err2 := r.GetHandler(channel, req.Method)
		if err2 != nil {
			return nil, err2
		}

		rsp := handler(req)

		return rsp.Marshal()
	}
}

func (r *router) AddChannelWithHandler(name string, handler ChannelHandler) error {
	if handler == nil {
		log.Errorln("channel handler must not be nil")
		return errors.New("channel handler is nil")
	}

	r.channelHandlers[name] = handler
	return nil
}

// RemoveChannel removes the channel, if exists, from the router.
//
// NOTE: Not goroutine safe.
func (r *router) RemoveChannel(name string) {
	delete(r.channelHandlers, name)
}

func (r *router) GetHandler(channel, method string) (Handler, error) {
	handler, ok := r.methodHandlers[channel][method]
	if !ok {
		return nil, errors.New(ErrCodeString[ErrServerMethodNotFound])
	}

	return handler, nil
}

// Register registers a handler for the
// given method, and replace if exists.
// If channel is not added yet, it will be registered before
// registering method handler.
//
// NOTE: not goroutine-safe.
func (r *router) Register(channel, method string, handler Handler) {
	if _, ok := r.channelHandlers[channel]; !ok {
		r.AddChannel(channel)
	}

	if _, ok := r.methodHandlers[channel]; !ok {
		r.methodHandlers[channel] = make(map[string]Handler)
	}

	r.methodHandlers[channel][method] = handler

	log.Infof("expose method %s at %s", method, channel)
}
