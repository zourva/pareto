package jsonrpc2

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"runtime/debug"
	"sync"
)

// Hook should be either Interceptor or PostHandler
type Hook = any

// Handler defines message layer handler for rpc.
type Handler = func(*RPCRequest) *RPCResponse

// Interceptor defines interceptor for message handlers.
//
// If an interceptor returns non-nil RPCResponse, then
// the intercepting chain will be terminated.
type Interceptor = func(*RPCRequest) *RPCResponse

// PostHandler defines post-handling interceptors for rpc method.
type PostHandler = func(*RPCRequest, *RPCResponse)

// Dispatcher defines underlying channel message dispatcher for rpc.
type Dispatcher = func(req []byte) (rsp []byte, err error)

type MethodDispatcher = func(raw []byte) (req *RPCRequest, rsp *RPCResponse)

// Router defines underlying invoker of the JSON-RPC server side.
// It has a routing arch of:
//
//	one Router manager
//	 \- many Channel routers and interceptors
//	   \- many Group routers and interceptors
//	     \- many method Handlers
//
// Note: all methods are goroutine-safe.
type Router interface {
	// AddChannel adds a routing channel to the
	// router with the given endpoint and handlers associated.
	AddChannel(endpoint string, handlers map[string]Handler, opts ...ChannelOption) RouterChannel

	// Channel returns the channel identified by name.
	Channel(name string) RouterChannel

	// Group returns or creates a routing group identified
	// by the group name and the channel bounded to.
	//Group(channel, group string) RouterGroup

	// GetHandler returns handler for a method on the
	// routing channel, and nil if not found.
	GetHandler(channel, method string) (Handler, error)

	// EnableTrace enables debug level req-rsp pair log printing if
	// set to true and clipLimit will be used to clip request body
	// and response body to the limit if needed.
	//
	// clipLimit is valid iff trace is enabled.
	// Set clipLimit to 0 means no clipping.
	// Set clipLimit <= 0 means using the default, which is 1024.
	EnableTrace(on bool, clipLimit int)

	// genMethodDispatcher generates a default dispatcher for the given channel.
	//genMethodDispatcher(channel string) Dispatcher

	//enableBindings enable binding registered channels
	//in a delayed way.
	enableBindings() error
}

func NewRouter(binder ChannelBinder) Router {
	if binder == nil {
		log.Fatalln("binder must not be nil")
		return nil
	}

	return &router{
		trace:         false,
		traceClip:     1024,
		channelBearer: binder,
		dispatchers:   make(map[string]Dispatcher),
		channels:      make(map[string]*routerChannel),
	}
}

// RouterChannel defines channel behavior for rpc.
//
// Note: all methods are goroutine-safe.
type RouterChannel interface {
	// Channel returns endpoint of this channel.
	Channel() string

	// MethodDispatcher returns dispatcher of this channel.
	MethodDispatcher() MethodDispatcher

	SetDispatcher(dispatcher MethodDispatcher)

	// Add registers handler for the given method,
	// replaced if already exists.
	Add(method string, handler Handler) RouterChannel

	// AddMap merges a map of handlers with existing ones.
	AddMap(handlers map[string]Handler) RouterChannel

	// Remove removes handler for the given method, if any.
	Remove(method string)

	// Handler returns the handler registered for method,
	// or nil if not exist.
	Handler(method string) Handler

	// AddInterceptors adds interceptors chained before invoking handler.
	AddInterceptors(interceptors ...Interceptor) RouterChannel

	// AddPostHandlers adds interceptors chained after invoked of handler.
	AddPostHandlers(postHandlers ...PostHandler) RouterChannel

	// Interceptors returns chained before-phase interceptors for a method handler.
	Interceptors(method string) []Hook

	// PostHandlers returns chained after-phase interceptors for a method handler.
	// After-phase interceptors are valid iff the handler is successfully invoked.
	PostHandlers(method string) []Hook

	// Group returns group of the given name,
	// which will be created if not exist and
	// will be merged if already exists.
	Group(name string, handlers map[string]Handler) RouterGroup
}

// RouterGroup defines group behavior for more than one method.
//
// Note: all methods are goroutine-safe.
type RouterGroup interface {
	// Name returns name of the group.
	Name() string

	// AddInterceptors configures interceptors chained before invoking handler.
	AddInterceptors(interceptors ...Interceptor) RouterGroup

	// AddPostHandlers configures interceptors chained after invoked of handler.
	AddPostHandlers(postHandlers ...PostHandler) RouterGroup
}

// router manages routing endpoints based
// /service/channel/method hierarchy path.
type router struct {
	sync.Mutex
	channelBearer ChannelBinder
	channels      map[string]*routerChannel
	dispatchers   map[string]Dispatcher

	trace     bool
	traceClip int
}

var _ Router = &router{}

// AddChannel adds a routing channel to the
// router with the given dispatcher.
func (r *router) AddChannel(endpoint string, handlers map[string]Handler, opts ...ChannelOption) RouterChannel {
	r.Lock()
	defer r.Unlock()

	channel, ok := r.channels[endpoint]
	if ok {
		log.Debugln("channel already exists")
		return channel
	}

	c := newRouterChannel(endpoint, handlers, opts...)
	if c.MethodDispatcher() == nil {
		c.SetDispatcher(r.genMethodDispatcher(endpoint))
		//log.Debugln("use default dispatcher")
	}

	// save channel
	r.channels[endpoint] = c

	// save dispatchers for binder
	r.dispatchers[endpoint] = r.dispatcher(c)

	log.Debugf("register channel %s with %d methods", endpoint, len(handlers))

	return c
}

func (r *router) Channel(name string) RouterChannel {
	r.Lock()
	defer r.Unlock()

	return r.getChannel(name)
}

func (r *router) GetHandler(channel, method string) (Handler, error) {
	r.Lock()
	defer r.Unlock()

	c := r.getChannel(channel)
	if c == nil {
		return nil, errors.New(ErrCodeString[ErrServerMethodNotFound])
	}

	h := c.Handler(method)
	if h == nil {
		return nil, errors.New(ErrCodeString[ErrServerMethodNotFound])
	}

	return h, nil
}

func (r *router) EnableTrace(on bool, clipLimit int) {
	r.trace = on

	if clipLimit < 0 {
		clipLimit = 1024
	}

	r.traceClip = clipLimit
}

// dispatcher creates a dispatcher for a channel binder.
func (r *router) dispatcher(c *routerChannel) Dispatcher {
	return func(raw []byte) (data []byte, err error) {
		method := ""
		defer func() {
			if r.trace {
				log.Debugf("jsonrpc response: %s, %s", method, r.clip(string(data)))
			}

			if p := recover(); p != nil {
				stack := debug.Stack()
				log.Errorln("router dispatcher recovered from:", p)
				fmt.Print("Show stack:\n", string(stack))
			}
		}()

		if r.trace {
			log.Debugf("jsonrpc request: %s", string(raw))
		}

		req, rsp := c.MethodDispatcher()(raw)
		if req == nil {
			if rsp != nil {
				// rsp is expected to be an error response
				return rsp.Marshal()
			} else {
				return NewErrorResponse(ErrServerInternal, "implementation error").Marshal()
			}
		}

		method = req.Method
		r.applyPostHandlers(c.PostHandlers(method), req, rsp)

		if rsp == nil {
			return NewErrorResponse(ErrServerInternal, "implementation error").Marshal()
		} else {
			return rsp.Marshal()
		}
	}
}

// DefaultDispatcher generates a default handler for the given channel.
func (r *router) genMethodDispatcher(channel string) MethodDispatcher {
	return func(reqRaw []byte) (req *RPCRequest, rsp *RPCResponse) {
		// parse and validate request
		req, err := ParseRequest(reqRaw)
		if err != nil {
			return nil, NewErrorResponse(err.Code, err.Message)
		}

		// get handler if any
		handler, err2 := r.GetHandler(channel, req.Method)
		if err2 != nil {
			return req, NewErrorResponse(ErrServerInternal, err2.Error())
		}

		// get channel mounted
		ch := r.Channel(channel)
		if ch == nil {
			return req, NewErrorResponse(ErrServerInternal, "channel:"+channel+" not found")
		}

		// invoke before-interceptors
		if bail := r.applyInterceptors(ch.Interceptors(req.Method), req); bail != nil {
			return req, bail
		}

		rsp = handler(req)

		return req, rsp
	}
}

// returns nil if all interceptors applied, and non-nil
// if any error occurred and the chained calls are terminated.
func (r *router) applyInterceptors(interceptors []Hook, req *RPCRequest) *RPCResponse {
	for _, interceptor := range interceptors {
		rsp := interceptor.(Interceptor)(req)

		// stop applying and return
		if rsp != nil {
			return rsp
		}
	}

	return nil
}

func (r *router) applyPostHandlers(handlers []Hook, req *RPCRequest, rsp *RPCResponse) {
	for _, handler := range handlers {
		go handler.(PostHandler)(req, rsp)
	}
}

// getChannel returns a RouterChannel of the given name.
func (r *router) getChannel(channel string) RouterChannel {
	if c, ok := r.channels[channel]; !ok {
		return nil
	} else {
		return c
	}
}

func (r *router) enableBindings() error {
	return r.channelBearer.Bind(r.dispatchers)
}

func (r *router) clip(s string) string {
	if len(s) > r.traceClip {
		return s[0:r.traceClip] + "..."
	}

	return s
}

type routerChannel struct {
	sync.RWMutex
	channel    string
	dispatcher MethodDispatcher
	handlers   map[string]Handler
	groups     map[string]*routerGroup

	interceptors []Hook
	postHandlers []Hook
}

type ChannelOption func(channel RouterChannel)

func UseDispatcher(dispatcher MethodDispatcher) ChannelOption {
	return func(c RouterChannel) {
		c.SetDispatcher(dispatcher)
	}
}

var _ RouterChannel = &routerChannel{}

func newRouterChannel(channel string, handlers map[string]Handler, opts ...ChannelOption) *routerChannel {
	c := &routerChannel{
		channel:  channel,
		handlers: handlers,
		groups:   make(map[string]*routerGroup),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (r *routerChannel) Channel() string {
	return r.channel
}

func (r *routerChannel) MethodDispatcher() MethodDispatcher {
	return r.dispatcher
}

func (r *routerChannel) Group(name string, handlers map[string]Handler) RouterGroup {
	r.Lock()
	defer r.Unlock()

	g, ok := r.groups[name]
	if !ok {
		g = newRouterGroup(name)
		r.groups[name] = g
	}

	if handlers != nil {
		for method, handler := range handlers {
			g.handlers[method] = handler
			r.handlers[method] = handler
		}
	}

	return g
}

func (r *routerChannel) SetDispatcher(dispatcher MethodDispatcher) {
	r.Lock()
	defer r.Unlock()

	r.dispatcher = dispatcher
}

func (r *routerChannel) Add(method string, handler Handler) RouterChannel {
	r.Lock()
	defer r.Unlock()

	if handler != nil {
		r.handlers[method] = handler
	}

	return r
}

func (r *routerChannel) Handler(method string) Handler {
	r.Lock()
	defer r.Unlock()

	if h, ok := r.handlers[method]; !ok {
		return nil
	} else {
		return h
	}
}

// AddMap merge handlers into the routing table.
func (r *routerChannel) AddMap(handlers map[string]Handler) RouterChannel {
	r.Lock()
	defer r.Unlock()

	if handlers != nil {
		for k, v := range handlers {
			r.handlers[k] = v
		}
	}

	return r
}

func (r *routerChannel) Remove(method string) {
	r.Lock()
	defer r.Unlock()

	delete(r.handlers, method)
}

func (r *routerChannel) AddInterceptors(interceptors ...Interceptor) RouterChannel {
	r.Lock()
	defer r.Unlock()

	for _, interceptor := range interceptors {
		if interceptor != nil {
			r.interceptors = append(r.interceptors, interceptor)
		}
	}

	log.Debugf("add %d intercepotrs up to %d for channel %s",
		len(interceptors), len(r.interceptors), r.channel)

	return r
}

func (r *routerChannel) AddPostHandlers(postHandlers ...PostHandler) RouterChannel {
	r.Lock()
	defer r.Unlock()

	for _, handler := range postHandlers {
		if handler != nil {
			r.postHandlers = append(r.postHandlers, handler)
		}
	}

	log.Debugf("add %d post handlers up to %d for channel %s",
		len(postHandlers), len(r.postHandlers), r.channel)

	return r
}

func (r *routerChannel) Interceptors(method string) []Hook {
	return r.extractHooks(method, r.interceptors,
		func(group *routerGroup) []Hook {
			var interceptors []Hook
			for _, ai := range group.interceptors {
				interceptors = append(interceptors, ai)
			}

			return interceptors
		})
}

func (r *routerChannel) PostHandlers(method string) []Hook {
	return r.extractHooks(method, r.postHandlers,
		func(group *routerGroup) []Hook {
			var handlers []Hook
			for _, ai := range group.postHandlers {
				handlers = append(handlers, ai)
			}

			return handlers
		})
}

func (r *routerChannel) extractHooks(method string, from []Hook,
	groupExtractor func(*routerGroup) []Hook) []Hook {
	r.Lock()
	defer r.Unlock()

	var hooks []Hook

	// extract global hooks
	for _, interceptor := range from {
		hooks = append(hooks, interceptor)
	}

	// extract group level hooks.
	// all hooks of groups containing
	// the method will be extracted.
	for _, group := range r.groups {
		if group == nil {
			continue
		}

		for name := range group.handlers {
			if method != name {
				continue
			}

			gh := groupExtractor(group)
			if len(gh) > 0 {
				hooks = append(hooks, gh...)
			}

			// quit group extraction when found
			// since method is unique within a group
			break
		}
	}

	return hooks
}

type routerGroup struct {
	sync.Mutex
	name     string
	handlers map[string]Handler // only handler names are used

	interceptors []Hook
	postHandlers []Hook
}

var _ RouterGroup = &routerGroup{}

func newRouterGroup(name string) *routerGroup {
	g := &routerGroup{
		name:         name,
		handlers:     make(map[string]Handler),
		interceptors: []Hook{},
		postHandlers: []Hook{},
	}
	return g
}

func (r *routerGroup) AddInterceptors(interceptors ...Interceptor) RouterGroup {
	r.Lock()
	defer r.Unlock()

	for _, interceptor := range interceptors {
		if interceptor != nil {
			r.interceptors = append(r.interceptors, interceptor)
		}
	}

	return r
}

func (r *routerGroup) AddPostHandlers(postHandlers ...PostHandler) RouterGroup {
	r.Lock()
	defer r.Unlock()

	for _, handler := range postHandlers {
		if handler != nil {
			r.postHandlers = append(r.postHandlers, handler)
		}
	}

	return r
}

func (r *routerGroup) Name() string {
	return r.name
}
