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

	//Registrar returns the delegator associated with this service.
	Registrar() *Registrar

	// Status returns the internal states this
	// service want to export to external world.
	//
	// Registrar will periodically call this method,
	// based on the value of Config.Interval, to get
	// the status object and export to the endpoint
	// defined by Config.Endpoint.
	//
	// The exported content and its format is
	// implementation-specific and should be carefully
	// designed by service provider and consumer.
	Status() *Status
	StatusConf() *StatusConf
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

// MetaService implements the Service interface and
// provides a bunch of methods for inheritance.
type MetaService struct {
	name        string        //name of this service
	registry    string        //registry this service registered to
	messager    *ipc.Messager //messager node or peer
	registrar   *Registrar    //registry client
	enableTrace bool          //enable trace of service messaging

	conf   *StatusConf //status report config
	status *Status     //status snapshot
}

func (s *MetaService) Name() string {
	return s.name
}

func (s *MetaService) Messager() *ipc.Messager {
	return s.messager
}

func (s *MetaService) Registrar() *Registrar {
	return s.registrar
}

func (s *MetaService) Status() *Status {
	return s.status
}

func (s *MetaService) StatusConf() *StatusConf {
	return s.conf
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
	s.status.Time = box.TimeNowMs()
	s.status.Health = s.conf
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
	if s.enableTrace {
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

func (s *MetaService) initialize() bool {
	if s.messager == nil { // create a default messager
		busName := fmt.Sprintf("%s-bus", s.name)
		rpcName := fmt.Sprintf("%s-rpc", s.name)
		messager, err := ipc.NewMessager(&ipc.MessagerConf{
			BusConf: &ipc.BusConf{Name: busName, Type: ipc.InterProcBus, Broker: s.registry},
			RpcConf: &ipc.RPCConf{Name: rpcName, Type: ipc.InterProcRpc, Broker: s.registry},
		})
		if messager == nil || err != nil {
			log.Errorln("create messager failed", err)
			return false
		}

		s.messager = messager
	}

	if s.registrar == nil { // create default registrar
		registrar := NewRegisterer(s.messager)
		if registrar == nil {
			log.Errorln("create registrar failed")
			return false
		}

		s.registrar = registrar
	}

	if s.conf == nil {
		s.conf = getDefaultStatusConf()
	}

	return true
}

// Start starts the given service in the following sequence:
//  1. registers the service to manager.
//  2. invokes the user callback service.Start and related hooks.
//  3. starts the status exporter of Registrar.
func Start(s Service) bool {
	s.SetState(Offline)

	if !s.Registrar().Register(s) {
		log.Errorf("register service %s failed", s.Name())
		return false
	}

	s.Registrar().EnableStatusReport()

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
//  1. stops the status exporter of Registrar.
//  2. invokes the user callback service.Stop and related hooks.
//  3. unregisters the service from manager.
func Stop(s Service) {
	s.SetState(Stopping)

	s.BeforeStopping()
	s.Shutdown()
	s.AfterStopping()

	s.SetState(Stopped)

	s.Registrar().DisableStatusExport()

	s.Registrar().Unregister()
}

// New creates a service with the given name, registry and options.
func New(desc *Descriptor, options ...Option) Service {
	return NewMetaService(desc, options...)
}

// NewMetaService creates a generic meta service with the given name
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
func NewMetaService(desc *Descriptor, options ...Option) *MetaService {
	name := desc.Name
	reg := desc.Registry
	if len(name) == 0 || len(reg) == 0 {
		log.Errorln("service name/registry must not be empty")
		return nil
	}

	s := &MetaService{
		name:        name,
		registry:    reg,
		enableTrace: false,
		status: &Status{
			Name:  name,
			State: Offline,
			Time:  box.TimeNowMs(),
		},
	}

	for _, fn := range options {
		fn(s)
	}

	if !s.initialize() {
		return nil
	}

	return s
}
