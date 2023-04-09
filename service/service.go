package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/ipc"
	"time"
)

// Service abstracts the service clients.
// A service comprises a messager, a registerer and hooks.
// A service has the lifecycle of:
//
//	server side				client side
//	created/enable
//					<----	BeforeStarting()
//	starting
//					<----	AfterStarting()
//	running
//					<----	BeforeStopping()
//	stopping
//					<----	AfterStopping()
//	stopped
//					<----	BeforeDestroyed()
//	destroyed/disable
type Service interface {
	// Name returns the name of the service
	Name() string

	//Messager returns the messager instance associated with this service.
	Messager() *ipc.Messager

	//Registerer returns the delegator associated with this service.
	Registerer() *Registerer

	//BeforeStarting is called after the service instance is created.
	BeforeStarting()

	//AfterStarting is called when the service finish initialization.
	AfterStarting()

	//BeforeStopping is called when the service is about to stop.
	BeforeStopping()

	//AfterStopping is called when the service finish shutdown.
	AfterStopping()

	//BeforeDestroyed is called before the service instance is destroyed.
	BeforeDestroyed()
}

// Config sums up info necessary to define a new service.
type Config struct {
	//Name of the service, mandatory.
	Name string

	//Messager to communicate with service server, mandatory.
	Messager *ipc.Messager

	//Registerer as a delegator to interact with service server, mandatory.
	Registerer *Registerer
}

// Endpoint defines the identity of a bus endpoint or rpc channel.
type Endpoint struct {
	Service string
	Object  string
	Method  string
}

// SerializedName returns the path-like name of the method, i.e.:
//
//	service.object.method
//
// e.g.:
//
//	webserver/cookie/get
func (r *Endpoint) SerializedName() string {
	return fmt.Sprintf("%s/%s/%s", r.Service, r.Object, r.Method)
}

// MetaService implements the Service interface and
// provides a bunch of methods for inheritance.
type MetaService struct {
	conf *Config
}

func (s *MetaService) Name() string {
	return s.conf.Name
}

func (s *MetaService) Messager() *ipc.Messager {
	return s.conf.Messager
}

func (s *MetaService) Registerer() *Registerer {
	return s.conf.Registerer
}

func (s *MetaService) BeforeStarting() {
	log.Infoln("ready to initialize service", s.Name())
}

func (s *MetaService) AfterStarting() {
	log.Infoln("finish initializing service", s.Name())
}

func (s *MetaService) BeforeStopping() {
	log.Infoln("ready to shutdown service", s.Name())
}

func (s *MetaService) AfterStopping() {
	log.Infoln("finish shutdown service", s.Name())
}

func (s *MetaService) BeforeDestroyed() {
	log.Infoln("about to destroy service", s.Name())
}

// Listen binds a handler to a subscribed topic.
// Old handler will be replaced if already bounded.
func (s *MetaService) Listen(topic string, fn ipc.Handler) error {
	log.Infof("%s subscribe to %s", s.Name(), topic)
	return s.Messager().Subscribe(topic, fn)
}

// Notify broadcasts a notice message to all subscribers and assumes no replies.
func (s *MetaService) Notify(topic string, data []byte) error {
	log.Debugf("%s publish to %s", s.Name(), topic)
	return s.Messager().Publish(topic, data)
}

// ExposeMethod registers a server-side method, identified by name, with the given handler.
func (s *MetaService) ExposeMethod(name string, fn ipc.CalleeHandler) error {
	log.Debugf("%s expose method %s", s.Name(), name)
	return s.Messager().ExposeV2(name, fn)
}

// CallMethod calls a remote method identified by id.
func (s *MetaService) CallMethod(name string, data []byte, to time.Duration) ([]byte, error) {
	log.Debugf("%s invoke rpc %s", s.Name(), name)
	return s.Messager().CallV2(name, data, to)
}

// NewMetaService creates a new meta service with the given conf.
// The newly created service is registered automatically to the server.
func NewMetaService(conf *Config) *MetaService {
	if conf == nil {
		log.Errorln("service config must not be nil")
		return nil
	} else {
		if len(conf.Name) == 0 ||
			conf.Messager == nil ||
			conf.Registerer == nil {
			log.Errorln("service config members must not be nil or empty")
			return nil
		}
	}

	s := &MetaService{
		conf: conf,
	}

	if !conf.Registerer.Register(s) {
		log.Errorf("register service %s failed", conf.Name)
		return nil
	}

	return s
}
