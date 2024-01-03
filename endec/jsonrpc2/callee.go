package jsonrpc2

import "errors"

type Handler func(*RPCRequest) *RPCResponse

// Router defines underlying bearer
// of the JSON-RPC server side.
type Router interface {
	Bind(func([]byte) ([]byte, error)) error
}

// Server defines the JSON-RPC server provider.
type Server struct {
	router  Router
	handler map[string]Handler
}

func (s *Server) route(reqBuf []byte) ([]byte, error) {
	req, err := ParseRequest(reqBuf)
	if err != nil {
		return nil, err
	}

	handler, ok := s.handler[req.Method]
	if !ok {
		return nil, errors.New(ErrCodeString[ErrServerMethodNotFound])
	}

	rsp := handler(req)

	return rsp.Marshal()
}

// RegisterHandler registers a handler for the
// given method, and replace the old one if exists.
//
// NOTE: not goroutine-safe.
func (s *Server) RegisterHandler(method string, handler Handler) {
	s.handler[method] = handler
}

// Serve binds to underlying router.
//
// NOTE: this method is not blocking.
func (s *Server) Serve() error {
	return s.router.Bind(s.route)
}

func NewServer(router Router) *Server {
	return &Server{
		router:  router,
		handler: make(map[string]Handler),
	}
}
