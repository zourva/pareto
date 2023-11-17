package service

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"sync"
	"time"
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

// service registry info
type registry struct {
	//service name
	name string
	//server-side-perceived state
	state State
	//timestamp in ms when up
	onlineTime uint64
	//timestamp in ms when down
	offlineTime uint64
	//timestamp in ms of last heartbeat
	updateTime uint64

	checkTimeout bool
	interval     uint64 //duration in milliseconds
	threshold    uint64
}

func (r *registry) timeout() bool {
	if !r.checkTimeout {
		return false
	}

	duration := box.TimeNowMs() - r.updateTime
	return duration > r.interval*r.threshold
}

func (r *registry) offline() {
	r.state = Offline
	log.Infof("force service %s offline", r.name)
}

// RegistryManager manages all services as service clients.
type RegistryManager struct {
	*MetaService
	services sync.Map //registry repository
	timer    *time.Timer
	duration time.Duration
}

// Startup starts the server.
func (s *RegistryManager) Startup() bool {
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

	_ = s.Listen(EndpointServiceStatus, s.handleStatus)

	s.timer = time.AfterFunc(s.duration, s.checkTimeout)

	log.Infoln("registry manager started")

	return true
}

// Shutdown stops the server.
// It notifies all the registered and alive service clients before quit.
func (s *RegistryManager) Shutdown() {
	//notify all registered services
	//s.messager.Publish(res.ServiceStop)
	//
	////wait for all attached services to quit,
	////making a graceful shutdown
	//if s.Size() == 0 {
	//	close(s.done)
	//} else {
	//	select {
	//	case <-s.done:
	//		log.Infoln("all attached services quit")
	//		break
	//	}
	//}
	s.timer.Stop()

	log.Infoln("registry manager shutdown")
}

// Registered returns true if the service is
// registered to the center and false otherwise.
//
//	This method is goroutine-safe.
func (s *RegistryManager) Registered(name string) bool {
	if _, ok := s.services.Load(name); ok {
		return true
	}

	return false
}

// Count returns number of services registered.
func (s *RegistryManager) Count() int {
	var counter = 0
	s.services.Range(func(key, value any) bool {
		counter++
		return true
	})

	return counter
}

// GetService returns the service associated with the
// given name or nil if not found.
//
//	This method is goroutine-safe.
func (s *RegistryManager) get(name string) *registry {
	if sd, ok := s.services.Load(name); ok {
		return sd.(*registry)
	}

	return nil
}

// Up saves a service with the given name and set
// state to online.
//
//	This method is goroutine-safe.
func (s *RegistryManager) register(status *Status) {
	t := box.TimeNowMs()
	s.services.Store(status.Name, &registry{
		name:       status.Name,
		state:      status.State,
		onlineTime: t,
		updateTime: t,
	})

	log.Infof("service %s registered, state = %s", status.Name, StateString(status.State))
}

func (s *RegistryManager) update(reg *registry, status *Status) {
	// update status conf
	if status.Health != nil {
		if status.Health.Threshold != 0 &&
			status.Health.Interval != 0 {
			reg.checkTimeout = true
			reg.interval = uint64(status.Health.Interval) * 1000
			reg.threshold = uint64(status.Health.Threshold)
		}
	}

	// notify watchers if state changed
	if reg.state != status.State {
		s.notifyWatched(reg, status)
	}

	// overwrite states
	reg.state = status.State
	reg.updateTime = box.TimeNowMs()

	// de-register if stopped normally
	if reg.state == Stopped {
		s.unregister(reg.name)
	}
}

// Down de-registers a service and sets state to offline.
// Does nothing when the service is not found.
//
//	This method is goroutine-safe.
func (s *RegistryManager) unregister(name string) {
	s.services.Delete(name)
	log.Infof("service %s unregistered", name)
}

func (s *RegistryManager) notifyWatched(reg *registry, status *Status) {
	// TODO:
	log.Infof("service %s state changed(%s -> %s)",
		reg.name, StateString(reg.state), StateString(status.State))
}

func (s *RegistryManager) handleStatus(data []byte) {
	status := &Status{}
	if err := json.Unmarshal(data, status); err != nil {
		log.Errorln("registry manager: json unmarshal failed:", err)
		return
	}

	if ss, ok := s.services.Load(status.Name); ok {
		reg := ss.(*registry)
		s.update(reg, status)
	} else {
		s.register(status)
	}
}

func (s *RegistryManager) checkTimeout() {
	s.services.Range(func(key, value any) bool {
		service := value.(*registry)
		if service.timeout() {
			//force to offline and notify watched
			service.offline()
			s.notifyWatched(service, &Status{
				Name:  service.name,
				State: Offline,
				Time:  box.TimeNowMs(),
			})
		}

		return true
	})

	s.timer.Reset(s.duration)
}

// NewRegistryManager creates a new service server.
func NewRegistryManager(broker string) *RegistryManager {
	s := &RegistryManager{
		MetaService: NewGenericMetaService("registry", broker),
		duration:    5 * time.Second,
	}

	log.Infoln("registry manager created")

	return s
}
