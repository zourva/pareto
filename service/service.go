package service

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
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
	// Name returns the name of the service.
	Name() string

	//Messager returns the messager instance associated with this service.
	Messager() *ipc.Messager

	//Registerer returns the delegator associated with this service.
	Registerer() *Registerer

	//Config returns the configuration info of this service.
	Config() *Config

	// Status returns the internal states this
	// service want to export to external world.
	//
	// Registerer will periodically call this method,
	// based on the value of Config.Interval, to get
	// the status object and export to the endpoint
	// defined by Config.Endpoint.
	//
	// The exported content and its format is
	// implementation-specific and should be carefully
	// designed by service provider and consumer.
	Status() *Status
	MarshalStatus() []byte
	SetState(state State)

	// Startup should be called by user after a service is created
	// to enable built-in functions such as status export.
	//
	// This method should be overwritten by every implementer.
	Startup() bool

	// Shutdown should be called before a service is destroyed
	// to disable built-in functions.
	//
	// This method should be overwritten by every implementer.
	Shutdown()

	//BeforeStarting is called before the service instance is started.
	BeforeStarting()

	//AfterStarting is called when the service finish initialization.
	AfterStarting()

	//BeforeStopping is called when the service is about to stop.
	BeforeStopping()

	//AfterStopping is called when the service finish shutdown.
	AfterStopping()

	////AfterCreated is called after the service instance is created.
	//AfterCreated()
	////BeforeDestroyed is called before the service instance is destroyed.
	//BeforeDestroyed()
}

// Config sums up info necessary to define a new service.
type Config struct {
	//Name of the service, mandatory.
	Name string

	//Messager to communicate with service server, mandatory.
	Messager *ipc.Messager

	//Registerer as a delegator to interact with service server, mandatory.
	Registerer *Registerer

	//EnableTrace
	EnableTrace bool

	//Status report config
	Status *StatusConf
}

// MetaService implements the Service interface and
// provides a bunch of methods for inheritance.
type MetaService struct {
	conf   *Config
	status *Status
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

func (s *MetaService) Config() *Config {
	return s.conf
}

func (s *MetaService) Status() *Status {
	return s.status
}

func (s *MetaService) Startup() bool {
	return true
}

func (s *MetaService) SetState(state State) {
	s.status.State = state
}

func (s *MetaService) Shutdown() {
}

func (s *MetaService) BeforeStarting() {
	log.Debugln("ready to initialize service", s.Name())
}

func (s *MetaService) AfterStarting() {
	log.Debugln("finish initializing service", s.Name())
}

func (s *MetaService) BeforeStopping() {
	log.Debugln("ready to shutdown service", s.Name())
}

func (s *MetaService) AfterStopping() {
	log.Debugln("finish shutdown service", s.Name())
}

func (s *MetaService) MarshalStatus() []byte {
	buf, err := json.Marshal(s.status)
	if err != nil {
		return []byte("")
	}

	return buf
}

//
//func (s *MetaService) BeforeDestroyed() {
//	log.Infoln("about to destroy service", s.Name())
//}

// Listen binds a handler to a subscribed topic.
// Old handler will be replaced if already bounded.
func (s *MetaService) Listen(topic string, fn ipc.Handler) error {
	log.Infof("%s subscribe to %s", s.Name(), topic)
	return s.Messager().Subscribe(topic, fn)
}

// Notify broadcasts a notice message to all subscribers and assumes no replies.
func (s *MetaService) Notify(topic string, data []byte) error {
	if s.conf.EnableTrace {
		log.Tracef("%s publish to %s", s.Name(), topic)
	}

	return s.Messager().Publish(topic, data)
}

// ExposeMethod registers a server-side method, identified by name, with the given handler.
func (s *MetaService) ExposeMethod(name string, fn ipc.CalleeHandler) error {
	log.Infof("%s expose method %s", s.Name(), name)
	return s.Messager().ExposeV2(name, fn)
}

// CallMethod calls a remote method identified by id.
func (s *MetaService) CallMethod(name string, data []byte, to time.Duration) ([]byte, error) {
	log.Tracef("%s invoke rpc %s", s.Name(), name)
	return s.Messager().CallV2(name, data, to)
}

// NewMetaService creates a new meta service with the given conf.
// The newly created service is registered automatically to the server.
// Returns nil when conf is invalid or when the registration failed.
func NewMetaService(conf *Config) *MetaService {
	if conf == nil {
		log.Errorln("service config must not be nil")
		return nil
	}

	if len(conf.Name) == 0 ||
		conf.Messager == nil ||
		conf.Registerer == nil {
		log.Errorln("service config members must not be nil or empty")
		return nil
	}

	if conf.Status == nil {
		conf.Status = getDefaultStatusConf()
	} else {
		if conf.Status.Interval == 0 {
			conf.Status.Interval = StatusReportInterval
		}

		if conf.Status.Threshold == 0 {
			conf.Status.Threshold = StatusLostThreshold
		}

		//if len(conf.Status.endpoint) == 0 {
		//	//conf.endpoint = fmt.Sprintf("%s/status", conf.Name)
		//	conf.Status.endpoint = EndpointServiceStatus
		//}
	}

	s := &MetaService{
		conf: conf,
		status: &Status{
			Name:  conf.Name,
			State: Offline,
			Time:  box.TimeNowMs(),
		},
	}

	return s
}

// NewGenericMetaService creates a generic meta service with the given name
// and use default values for other config items.
//
// A default messager is created with:
//
//  1. both BUS and RPC capabilities enabled,
//  2. both BUS and RPC using inter-proc pattern with the same broker endpoint,
//  3. names for BUS & RPC created from the service name with a format of
//     {service name}-bus and {service name}-rpc
//
// A default register is also created associating with the default messager.
func NewGenericMetaService(name, broker string) *MetaService {
	if len(name) == 0 || len(broker) == 0 {
		log.Errorln("service name/broker must not be empty")
		return nil
	}

	// create default messager
	busName := fmt.Sprintf("%s-bus", name)
	rpcName := fmt.Sprintf("%s-rpc", name)
	messager, err := ipc.NewMessager(&ipc.MessagerConf{
		BusConf: &ipc.BusConf{
			Name:   busName,
			Type:   ipc.InterProcBus,
			Broker: broker},
		RpcConf: &ipc.RPCConf{
			Name:   rpcName,
			Type:   ipc.InterProcRpc,
			Broker: broker},
	})
	if messager == nil || err != nil {
		log.Errorln("create default messager failed", err)
		return nil
	}

	// create default registerer
	registerer := NewRegisterer(messager)
	if registerer == nil {
		log.Errorln("create default registerer failed")
		return nil
	}

	conf := &Config{
		Name:        name,
		Messager:    messager,
		Registerer:  registerer,
		EnableTrace: false,
		Status:      getDefaultStatusConf(),
	}

	return NewMetaService(conf)
}

// Start starts the given service in the following sequence:
//  1. registers the service to manager.
//  2. invokes the user callback service.Start and related hooks.
//  3. starts the status exporter of Registerer.
func Start(s Service) bool {
	s.SetState(Offline)

	if !s.Registerer().Register(s) {
		log.Errorf("register service %s failed", s.Name())
		return false
	}

	s.Registerer().EnableStatusExport()

	s.SetState(Starting)

	s.BeforeStarting()
	if !s.Startup() {
		return false
	}
	s.AfterStarting()

	s.SetState(Servicing)

	return true
}

// Stop stops the given service in the following sequence:
//  1. stops the status exporter of Registerer.
//  2. invokes the user callback service.Stop and related hooks.
//  3. unregisters the service from manager.
func Stop(s Service) {
	s.SetState(Stopping)

	s.BeforeStopping()
	s.Shutdown()
	s.AfterStopping()

	s.SetState(Stopped)

	s.Registerer().DisableStatusExport()

	s.Registerer().Unregister()
}
