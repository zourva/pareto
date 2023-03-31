package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"github.com/zourva/pareto/ipc"
	"sync"
)

//// RegistryServer manages all services implementing the Service interface.
//type RegistryServer interface {
//	ipc.Bus
//	ipc.RPC
//
//	// Startup starts the manager.
//	Startup() error
//
//	// Shutdown stops the manager and all services.
//	Shutdown()
//
//	// GetService returns the service associated with the
//	// given name or nil if not found.
//	GetService(name string) Service
//
//	// Up registers a service with the given name and instance.
//	// Instance will be substituted if already exists.
//	Up(name string, svc Service)
//
//	// Down detaches the service, identified by the name, from the manager.
//	// Detached services will not be joined when service manager exits.
//	// Does nothing when the service is not found.
//	Down(name string)
//
//	// Detached returns true if the service is detached from the manager already.
//	// return false when the service is not found.
//	Detached(name string) bool
//}

// service registry state
type state int

const (
	online state = iota
	pending
	offline
	unknown
)

// service registry info
type registry struct {
	//service registry format:
	//  {service name}:{messager name}
	//  e.g.:
	// 		ip:port
	//		system-monitor:monitor-rpc-messager
	name string
	//server-side-perceived state
	state state
	//timestamp in us when up
	onlineTime uint64
	//timestamp in us when down
	offlineTime uint64
	//timestamp in us of last heartbeat
	aliveTime uint64
}

// RegistryServer manages all services as service clients.
type RegistryServer struct {
	messager *ipc.Messager

	//services repository
	services sync.Map

	//join for all
	//done chan struct{}
}

// GetService returns the service associated with the
// given name or nil if not found.
//
//	This method is goroutine-safe.
func (s *RegistryServer) GetService(name string) Service {
	if sd, ok := s.services.Load(name); ok {
		return sd.(Service)
	}

	return nil
}

// Registered returns true if the service is registered to the server and
// false when not found.
//
//	This method is goroutine-safe.
func (s *RegistryServer) Registered(name string) bool {
	if _, ok := s.services.Load(name); ok {
		return true
	}

	return false
}

// Up registers a service with the given name and set
// the its state to online.
//
//	This method is goroutine-safe.
func (s *RegistryServer) Up(name string) {
	t := box.TimeNowUs()
	s.services.Store(name, &registry{
		name:       name,
		state:      online,
		onlineTime: t,
		aliveTime:  t,
	})
}

// Down de-registers a service and set its state to offline.
// Does nothing when the service is not found.
//
//	This method is goroutine-safe.
func (s *RegistryServer) Down(name string) {
	if _, ok := s.services.Load(name); ok {
		s.services.Store(name, &registry{
			name:        name,
			state:       offline,
			offlineTime: box.TimeNowUs(),
		})
	}
}

// Shutdown stops the server.
// It notifies all the registered and alive service clients before quit.
func (s *RegistryServer) Shutdown() {
	//notify all registered services
	//s.messager.Publish(res.ServiceStop)
	//
	////wait for all attached services to quit,
	////making a graceful shutdown
	//if s.serviceCount() == 0 {
	//	close(s.done)
	//} else {
	//	select {
	//	case <-s.done:
	//		log.Infoln("all attached services quit")
	//		break
	//	}
	//}

	log.Infoln("service manager shutdown")
}

// Startup starts the server.
func (s *RegistryServer) Startup() error {
	////enable service join when exiting
	//if err := s.Listen(res.ServiceDown, s.onServiceDown); err != nil {
	//	return err
	//}
	//
	////transfer control to services themselves
	//if err := s.Notify(res.ServiceStart); err != nil {
	//	return err
	//}
	//
	//s.Messager().Publish(topic, args...)

	log.Infoln("service manager started")

	return nil
}

func (s *RegistryServer) serviceCount() int {
	var alive = 0
	s.services.Range(func(key, value interface{}) bool {
		alive++
		return true
	})

	return alive
}

func (s *RegistryServer) onServiceDown(serviceName string) {
	log.Infof("[%s] quit acknowledged", serviceName)

	service := s.GetService(serviceName)
	if service != nil {
		service.BeforeDestroyed()

		s.services.Delete(serviceName)
	}

	if s.serviceCount() == 0 {
		//close(s.done)
	}
}

const (
	busName    = "service server messager bus"
	rpcName    = "service server messager rpc"
	brokerAddr = "nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222"
)

// NewServer creates a new service server.
func NewServer() *RegistryServer {
	sm := &RegistryServer{
		//done: make(chan struct{}),
	}

	m := ipc.NewMessager(&ipc.MessagerConf{
		BusConf: &ipc.BusConf{Name: busName, Type: ipc.InterProcBus, Broker: brokerAddr},
		RpcConf: &ipc.RPCConf{Name: busName, Type: ipc.InterProcRpc},
	})

	if m == nil {
		log.Errorln("create messager failed")
		return nil
	}

	log.Infoln("service manager created")

	return sm
}
