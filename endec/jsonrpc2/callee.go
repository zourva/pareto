package jsonrpc2

type ChannelBinder interface {
	// Bind binds router channels to physical impl.
	Bind(channels map[string]Dispatcher) error
}

// Server defines the JSON-RPC server provider.
type Server struct {
	router Router
}

// RegisterHandler registers a handler for the
// given method, and replace the old one if exists,
// in the default channel.
//
// Deprecated: Use server.Router().Register instead.
//func (s *Server) RegisterHandler(channel, method string, handler Handler) {
//	s.router.DefaultGroup(channel).Add(method, handler)
//}

// Serve binds to underlying router.
//
// NOTE: this method is not blocking.
func (s *Server) Serve() error {
	return s.router.enableBindings()
}

func (s *Server) Router() Router {
	return s.router
}

func NewServer(router Router) *Server {
	return &Server{
		router: router,
	}
}
