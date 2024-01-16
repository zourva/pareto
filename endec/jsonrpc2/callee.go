package jsonrpc2

// Server defines the JSON-RPC server provider.
type Server struct {
	router Router
}

func (s *Server) route(reqBuf []byte) ([]byte, error) {
	req, err := ParseRequest(reqBuf)
	if err != nil {
		return nil, err
	}

	handler, err2 := s.router.GetHandler(req.Method)
	if err2 != nil {
		return nil, err2
	}

	rsp := handler(req)

	return rsp.Marshal()
}

// RegisterHandler registers a handler for the
// given method, and replace the old one if exists.
//
// NOTE: not goroutine-safe.
func (s *Server) RegisterHandler(method string, handler Handler) {
	s.router.Register(method, handler)
}

// Serve binds to underlying router.
//
// NOTE: this method is not blocking.
func (s *Server) Serve() error {
	return s.router.Binder().Bind(s.route)
}

func NewServer(router Router) *Server {
	return &Server{
		router: router,
	}
}
