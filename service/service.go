package service

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"github.com/zourva/pareto/endec/jsonrpc2"
	"github.com/zourva/pareto/ipc"
	"time"
)

// Service abstracts the service clients.
// A service comprises a messager, a registerer and hooks.
// A service has the lifecycle of:
//
//	lifecycle:  offline -> starting -> servicing -> stopping -> stopped
//	liveness:   dead    -> starting ->   alive   -> stopping -> dead
//	readiness:  false   ->  false   -> true/false-> false    -> false
type Service interface {
	Messaging

	// Name returns the name of the service.
	Name() string

	//Messager returns internal messager instance.
	Messager() *ipc.Messager

	//Registrar returns the delegator associated with this service.
	Registrar() Registrar

	// Status returns the liveness and readiness states
	// this service exposed.
	//
	// Registrar will periodically call this method,
	// based on the value of StatusConf.Interval, to get
	// the status object and report it to the registry.
	//
	// The content exported and its format are
	// implementation-specific and should be carefully
	// designed by service provider and consumer.
	Status() *Status
	StatusConf() *StatusConf
	MarshalStatus() []byte

	// SetState changes liveness state of
	// this service manually. Use it carefully
	// since liveness may cause cascade failures.
	SetState(state State)

	// State returns liveness of this service.
	State() State

	// SetReady changes readiness state of
	// this service.
	SetReady(ready bool)

	// Ready returns readiness of this service.
	Ready() bool

	// Startup should be called by user after a service is created
	// to enable built-in functions such as status export.
	//
	// This method is expected to be overwritten.
	Startup() bool

	// Shutdown should be called before a service is destroyed
	// to disable built-in functions.
	//
	// This method is expected to be overwritten.
	Shutdown()

	//AfterRegistered is called when the service finish registration.
	AfterRegistered()

	//BeforeStarting is called before the service instance is started.
	BeforeStarting()

	//AfterStarting is called when the service finish initialization.
	AfterStarting()

	//CheckRecovery is called AfterStarting and Before Servicing.
	CheckRecovery(list *StatusList)

	//BeforeStopping is called when the service is about to stop.
	BeforeStopping()

	//AfterStopping is called when the service finish shutdown.
	AfterStopping()
}

// MetaService implements the Service interface and
// provides a bunch of methods for inheritance.
type MetaService struct {
	name        string //name of this service
	registry    string //registry this service registered to
	enableTrace bool   //enable trace of service messaging

	conf   *StatusConf //status report config
	status *Status     //status snapshot

	invoker  *jsonrpc2.Client  //JSON rpc caller
	exposer  *jsonrpc2.Server  //JSON rpc callee
	messager *ipc.Messager     //raw messager bound to
	handler  ipc.CalleeHandler //private RR channel handler

	registrar Registrar //registry client

	watched []string //watched service list, not thread-safe
	//locker  sync.Locker
}

func (s *MetaService) Name() string {
	return s.name
}

func (s *MetaService) Messager() *ipc.Messager {
	return s.messager
}

func (s *MetaService) Registrar() Registrar {
	return s.registrar
}

func (s *MetaService) RpcClient() *jsonrpc2.Client {
	return s.invoker
}

func (s *MetaService) RpcServer() *jsonrpc2.Server {
	return s.exposer
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

// SetState changes liveness state of
// this service manually. Use it carefully
// since liveness may cause cascade failures.
func (s *MetaService) SetState(state State) {
	s.status.State = state
}

// State returns liveness of this service.
func (s *MetaService) State() State {
	return s.status.State
}

func (s *MetaService) SetReady(r bool) {
	s.status.Ready = r
}

// Ready returns readiness of this service.
func (s *MetaService) Ready() bool {
	return s.status.Ready
}

func (s *MetaService) Shutdown() {
}

func (s *MetaService) AfterRegistered() {
	if s.handler != nil {
		err := s.ExposeMethod(EndpointServiceRRHandlePrefix+s.Name(), s.handler)
		if err != nil {
			log.Fatalf("expose service handle to registry failed: %v", err)
		}
	}

	log.Debugln("finish registering service", s.Name())
}

func (s *MetaService) BeforeStarting() {
	log.Debugln("ready to initialize service", s.Name())
}

func (s *MetaService) AfterStarting() {
	log.Debugln("finish initializing service", s.Name())
}

func (s *MetaService) CheckRecovery(list *StatusList) {
	//override expected if necessary
	log.Debugln("finish recovery checking for", s.Name())
}

func (s *MetaService) BeforeStopping() {
	log.Debugln("ready to shutdown service", s.Name())
}

func (s *MetaService) AfterStopping() {
	log.Debugln("finish shutdown service", s.Name())
}

func (s *MetaService) MarshalStatus() []byte {
	s.status.Time = box.TimeNowMs()
	//s.status.Conf = s.conf
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
	log.Infof("%s expose method at %s", s.Name(), name)
	return s.Messager().ExposeV2(name, fn)
}

// CallMethod calls a remote method identified by id.
func (s *MetaService) CallMethod(name string, data []byte, to time.Duration) ([]byte, error) {
	log.Tracef("%s invoke rpc %s", s.Name(), name)
	return s.Messager().CallV2(name, data, to)
}

func (s *MetaService) AddCallerForwarder(sink string, to time.Duration) Forwarder {
	return func(data []byte) ([]byte, error) {
		return s.CallMethod(sink, data, to)
	}
}

func (s *MetaService) AddCalleeForwarder(sink string, f Forwarder) error {
	return s.ExposeMethod(sink, f)
}

// Watch registers an observation for a given service.
// The watch function will be invoked when registry detect any
// service state change. If whitelist is provided, watch is invoked iff
// state of services in the list changed.
func (s *MetaService) Watch(watch func(status *Status), whitelist ...string) error {
	//if len(spec.Watched) == 0 {
	//	return errors.New("target service name must not be empty")
	//}

	//if len(spec.Channel) == 0 {
	//	log.Tracef("no notify channel provided, ignored")
	//	return nil
	//}

	//spec.Watcher = s.name
	//
	//if len(spec.TargetStates) == 0 {
	//	spec.TargetStates = append(spec.TargetStates, defaultWatchedStates...)
	//}

	//buf, err := json.Marshal(spec)
	//if err != nil {
	//	log.Errorf("watch service %s failed", err)
	//	return err
	//}

	//_, err = s.CallMethod(EndpointServiceNotice, buf, time.Second)
	//if err != nil {
	//	log.Errorf("watch service %s failed", err)
	//	return err
	//}

	if len(whitelist) != 0 {
		//s.locker.Lock()
		//defer s.locker.Unlock()
		s.watched = whitelist
	}

	err := s.Listen(EndpointServiceNotice, func(data []byte) {
		if len(whitelist) == 0 {
			return
		}

		status := &Status{}
		if err := json.Unmarshal(data, status); err != nil {
			log.Errorf("service %s: json unmarshal failed: %v", s.name, err)
			return
		}

		if s.serviceWatched(status.Name) {
			watch(status)
		}
	})

	return err
}

func (s *MetaService) serviceWatched(name string) bool {
	for _, w := range s.watched {
		if w == name {
			return true
		}
	}

	return false
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
		reg := NewRegistrar(s)
		if reg == nil {
			log.Errorln("create registrar failed")
			return false
		}

		s.registrar = reg
	}

	s.invoker = jsonrpc2.NewClient(NewJsonRpcInvoker(s))
	s.exposer = jsonrpc2.NewServer(jsonrpc2.NewRouter(NewJsonRpcBinder(s)))

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

	if !s.Registrar().Register() {
		log.Errorf("register service %s failed", s.Name())
		return false
	}

	s.AfterRegistered()

	s.SetState(Starting)

	s.BeforeStarting()
	if !s.Startup() {
		return false
	}
	s.AfterStarting()

	s.CheckRecovery(s.Registrar().StatusList())

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
		//locker:      concurrent.NewSpinLock(),
		status: &Status{
			Name:  name,
			State: Offline,
			Time:  box.TimeNowMs(),
			Ready: false,
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
