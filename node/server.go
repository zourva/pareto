package node

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"github.com/zourva/pareto/env"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/stats"
	"math"
	"net"
	"time"
)

type ServerSideCallback func(*Node)

// ServerSideHooks defines callbacks exposed on server side.
type ServerSideHooks struct {
	//called when a certain node is successfully authed
	OnNodeJoin ServerSideCallback

	//called when stream message received
	OnNodeNasMsg func(*StreamMessage)

	//called when a certain node checked out
	OnNodeLeave ServerSideCallback
}

// option func-closure pattern
type ServerOption func(agent *Server)

// serverOptions used by server
type serverOptions struct {
	network  string
	endpoint string //listen endpoint
	hooks    ServerSideHooks
	cluster  bool
	raftAddr string
	peerAddr []string
}

func defaultServerOptions() serverOptions {
	return serverOptions{
		network:  tcpNetwork,
		endpoint: listenEndpoint,
		hooks:    ServerSideHooks{},
		cluster:  false,
	}
}

// Server models node of the server side.
// Based on grpc and protocol buffer v3,
// we define the s1 interface procedures.
type Server struct {
	server  *grpc.Server
	options serverOptions
	confMgr ServerConfManager
	ssnMgr  *sessionManager
}

func WithServerHooks(cbs ServerSideHooks) ServerOption {
	return func(s *Server) {
		s.options.hooks = cbs
	}
}

// ep is raft protocol listen address of this server node.
// peers contains addresses of other raft server nodes.
func WithClusterMode(ep string, peers []string) ServerOption {
	return func(s *Server) {
		s.options.cluster = true
		s.options.peerAddr = peers
		s.options.raftAddr = ep
	}
}

func NewServer(endpoint string, opts ...ServerOption) *Server {
	if !box.ValidateEndpoint(endpoint) {
		return nil
	}

	s := &Server{
		options: defaultServerOptions(),
		confMgr: NewServerConfManager(env.GetExecFilePath() + "/../etc/node.db"),
	}

	s.ssnMgr = newSessionManager(s)
	s.options.endpoint = endpoint

	for _, opt := range opts {
		opt(s)
	}

	log.Infoln("new node server with endpoint", endpoint)

	return s
}

// Start starts the node server and blocks till stopped.
func (s *Server) Start() error {
	lis, err := net.Listen(s.options.network, s.options.endpoint)
	if err != nil {
		log.Errorln("node server failed to listen:", err)
		return err
	}

	srv := grpc.NewServer(
		grpc.StatsHandler(s),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     time.Duration(math.MaxInt64),
			MaxConnectionAge:      time.Duration(math.MaxInt64),
			MaxConnectionAgeGrace: time.Duration(math.MaxInt64),
			Time:                  2 * time.Hour,
			Timeout:               30 * time.Second,
		}),
		/*grpc.ChainStreamInterceptor(a.streamInterceptorLogger),
		grpc.InitialWindowSize(64*1024),
		grpc.InitialConnWindowSize(16*64*1024),
		grpc.ChainUnaryInterceptor(a.validator)*/)
	RegisterS1ServiceServer(srv, NewServerProto(s))
	reflection.Register(srv)

	s.server = srv

	log.Infoln("node server started")

	// blocks here
	if err := srv.Serve(lis); err != nil {
		log.Errorln("node server serve failed:", err)
		return err
	}

	return nil
}

// Stop stops the server,
// and blocks until all the pending RPCs are finished if graceful is true,
// otherwise stops immediately by cancelling the pending RPCs.
func (s *Server) Stop(graceful bool) {
	if graceful {
		s.server.GracefulStop()
	} else {
		s.server.Stop()
	}

	log.Infoln("node server stopped")
}

func (s *Server) TagRPC(ctx context.Context, tag *stats.RPCTagInfo) context.Context {
	//log.Traceln("see rpc call:", tag.FullMethodName)
	return ctx
}

func (s *Server) HandleRPC(ctx context.Context, stats stats.RPCStats) {
	//log.Traceln("handle rpc call:", ctx.Value(r.ConnId))
	return
}

// TagConn prepares a key, using the underlying *stats.ConnTagInfo, for a new session.
func (s *Server) TagConn(ctx context.Context, connTag *stats.ConnTagInfo) context.Context {
	log.Infoln("server see connection pair:",
		connTag.RemoteAddr.String(), "-",
		connTag.LocalAddr.String())

	// piggyback a <session key id, session key> pair when a new connection is created
	return context.WithValue(ctx, sessionKeyId, connTag)
}

// HandleConn handles creation and deletion of connection sessions (phase I),
// using the session key prepared by TagConn callback.
func (s *Server) HandleConn(ctx context.Context, connStats stats.ConnStats) {
	key := s.ssnMgr.getSessionKey(ctx)
	if key == nil {
		log.Errorln("illegal state: no connection tag found")
		return
	}

	switch connStats.(type) {
	case *stats.ConnBegin:
		s.ssnMgr.save(key)
		log.Infof("begin conn %p, #connections = %d", key, s.ssnMgr.size())
	case *stats.ConnEnd:
		s.ssnMgr.delete(key)
		log.Infof("end conn %p, #connections = %d", key, s.ssnMgr.size())
	default:
		log.Infoln("illegal connStats type")
	}

	return
}
