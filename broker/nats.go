package broker

import (
	"errors"
	"github.com/nats-io/nats-server/v2/server"
	log "github.com/sirupsen/logrus"
	"time"
)

// EmbeddedNats embeds the nats MQ server.
//
// see https://docs.nats.io/running-a-nats-service/clients#embedding-nats
// and https://dev.to/karanpratapsingh/embedding-nats-in-go-19o.
type EmbeddedNats struct {
	ns *server.Server

	host    string //service host, "127.0.0.1" by default
	port    int    //service port, 4222 by default
	mHost   string //monitor host, "127.0.0.1" by default
	mPort   int    //monitor port, 8222 by default
	retry   time.Duration
	retries int
	token   string
	logger  string //stdout/stderr or file path
}

// Startup starts the broker.
func (s *EmbeddedNats) Startup() error {
	//ns.Start is blocking
	go s.ns.Start()

	//wait for ready
	for i := 1; i <= s.retries; i++ {
		if s.ns.ReadyForConnections(s.retry) {
			break
		}

		log.Warnf("nats server is not ready yet, %d retry", i)

		if i == s.retries {
			//failed, shutdown and quit
			log.Errorln("nats broker startup timeout")
			s.ns.Shutdown()
			return errors.New("startup timeout")
		}
	}

	log.Infoln("nats broker started")

	return nil
}

// Shutdown stops the broker.
func (s *EmbeddedNats) Shutdown() error {
	s.ns.Shutdown()

	log.Infoln("nats broker stopped")
	return nil
}

func (s *EmbeddedNats) natsOptions() *server.Options {
	opt := &server.Options{}

	opt.Host = s.host
	opt.Port = s.port
	opt.HTTPHost = s.mHost
	opt.HTTPPort = s.mPort
	opt.Authorization = s.token
	if s.logger == "stdout" || s.logger == "stderr" {
		opt.LogFile = ""
	} else {
		opt.LogFile = s.logger
	}

	return opt
}

type Option = func(*EmbeddedNats)

// WithHost overrides default host "127.0.0.1".
func WithHost(host string) Option {
	return func(mq *EmbeddedNats) {
		mq.host = host
	}
}

// WithMonitorHost overrides default host "127.0.0.1".
func WithMonitorHost(host string) Option {
	return func(mq *EmbeddedNats) {
		mq.mHost = host
	}
}

// WithPort overrides default port 4222.
func WithPort(port int) Option {
	return func(mq *EmbeddedNats) {
		if port < 0 {
			port = 4222
		}
		mq.port = port
	}
}

// WithMonitorPort overrides default port 8222.
func WithMonitorPort(port int) Option {
	return func(mq *EmbeddedNats) {
		if port < 0 {
			port = 8222
		}
		mq.mPort = port
	}
}

// WithRetryDuration overrides default retry duration
// which is 5 seconds.
func WithRetryDuration(d time.Duration) Option {
	return func(mq *EmbeddedNats) {
		mq.retry = d
	}
}

// WithRetryCount overrides default retry count, 3.
func WithRetryCount(c int) Option {
	return func(mq *EmbeddedNats) {
		mq.retries = c
	}
}

func WithAuthorizationToken(t string) Option {
	return func(mq *EmbeddedNats) {
		mq.token = t
	}
}

// WithLoggerFile defines logger for embedded nats,
// file can be "stdout", "stderr" or any file path.
// Logging is disabled if file is "", which is the default.
func WithLoggerFile(file string) Option {
	return func(mq *EmbeddedNats) {
		mq.logger = file
	}
}

// NewEmbeddedNats creates and initialize an embedded nats
// server using the given options.
func NewEmbeddedNats(opts ...Option) (*EmbeddedNats, error) {
	srv := &EmbeddedNats{
		host:    "127.0.0.1",
		port:    4222,
		mHost:   "127.0.0.1",
		mPort:   8222,
		retry:   5 * time.Second,
		retries: 3,
		token:   "dag0HTXl4RGg7dXdaJwbC8",
		logger:  "", //empty means disabled
	}

	for _, opt := range opts {
		opt(srv)
	}

	ns, err := server.NewServer(srv.natsOptions())
	if err != nil {
		log.Errorln("create nats broker failed:", err)
		return nil, err
	}

	if len(srv.logger) != 0 {
		ns.ConfigureLogger()
	}

	srv.ns = ns

	log.Infoln("nats broker created")

	return srv, nil
}
