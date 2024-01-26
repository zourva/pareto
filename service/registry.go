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
	//liveness state
	state State
	//readiness
	ready bool
	//timestamp in ms when up
	onlineTime uint64
	//timestamp in ms when down
	offlineTime uint64
	//timestamp in ms of last heartbeat
	updateTime uint64

	interval  uint64 //duration in milliseconds
	threshold uint64 //number of failures allowed
}

func (r *registry) timeout() bool {
	duration := box.TimeNowMs() - r.updateTime
	return duration > r.interval*r.threshold
}

func (r *registry) offline() {
	r.state = Offline
	r.ready = false
	r.updateTime = box.TimeNowMs()
	r.offlineTime = r.updateTime
	log.Infof("force service %s offline", r.name)
}

func (r *registry) update(s *Status) {
	// update check conditions
	if s.CheckInterval != 0 {
		r.interval = uint64(s.CheckInterval)
	}

	if s.AllowFailures != 0 {
		r.threshold = uint64(s.AllowFailures)
	}

	r.state = s.State
	r.updateTime = box.TimeNowMs()
}

func (r *registry) toStatus() *Status {
	return &Status{
		Name:  r.name,
		State: r.state,
		Ready: r.ready,
		Time:  r.updateTime,
	}
}

// RegistryManager manages all services as service clients.
type RegistryManager struct {
	*MetaService
	services sync.Map      //registry repository
	timer    *time.Timer   //timeout check timer
	duration time.Duration //timeout check timer duration, 5s by default

	//watchers map[string][]*Watcher
	//mutex    sync.RWMutex
}

// Startup starts the server.
func (s *RegistryManager) Startup() bool {
	_ = s.Listen(EndpointServiceStatus, s.handleStatus)

	//_ = s.ExposeMethod(EndpointServiceNotice, s.handleWatch)

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

// registry spawns an instance of registry from status.
func (s *RegistryManager) registry(status *Status) *registry {
	t := box.TimeNowMs()

	r := &registry{
		name:       status.Name,
		state:      status.State,
		ready:      status.Ready,
		updateTime: t,
		onlineTime: t,
		interval:   uint64(status.CheckInterval) * 1000,
		threshold:  uint64(status.AllowFailures),
	}

	box.SetIfEq(&r.interval, 0, StatusReportInterval*1000)
	box.SetIfEq(&r.threshold, 0, StatusLostThreshold)

	return r
}

// Up saves a service with the given name and set
// state to online.
//
//	This method is goroutine-safe.
func (s *RegistryManager) register(status *Status) {
	s.services.Store(status.Name, s.registry(status))

	log.Infof("service %s registered, state = %s", status.Name, status.State.String())
}

func (s *RegistryManager) update(reg *registry, status *Status) {
	// notify watchers if state changed
	if reg.state != status.State ||
		reg.ready != status.Ready {
		s.notifyWatched(reg, status)
	}

	// overwrite states
	reg.update(status)

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
	log.Infof("service %s state changed(%s -> %s)",
		reg.name, reg.state, status.State)

	data, _ := json.Marshal(status)
	_ = s.Notify(EndpointServiceNotice, data)

	//s.mutex.RLock()
	//defer s.mutex.RUnlock()
	//
	//watchers, ok := s.watchers[reg.name]
	//if !ok {
	//	log.Tracef("no watcher registered for service %s", reg.name)
	//	return
	//}
	//
	//// multi-cast
	//data, _ := json.Marshal(status)
	//for _, w := range watchers {
	//	if len(w.spec.Channel) == 0 {
	//		continue
	//	}
	//
	//	_ = s.Notify(w.spec.Channel, data)
	//}
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

//func (s *RegistryManager) handleWatch(data []byte) ([]byte, error) {
//	spec := &WatchSpec{}
//	if err := json.Unmarshal(data, spec); err != nil {
//		log.Errorln("registry manager: json unmarshal watch spec failed:", err)
//		return nil, nil
//	}
//
//	s.mutex.RLock()
//	defer s.mutex.RUnlock()
//
//	watchers, ok := s.watchers[spec.Watched]
//	if !ok {
//		s.watchers[spec.Watched] = append(s.watchers[spec.Watched], &Watcher{
//			spec: *spec,
//		})
//		return []byte("ok"), nil
//	}
//
//	var watcher *Watcher
//	for _, watcher = range watchers {
//		if watcher.spec.Watcher == spec.Watcher {
//			break
//		}
//	}
//
//	//update
//	if watcher != nil {
//		log.Debugf("service %s watcher %s updated", spec.Watched, spec.Watcher)
//		watcher.spec = *spec
//	} else {
//		watchers = append(watchers, &Watcher{
//			spec: *spec,
//		})
//	}
//
//	return []byte("ok"), nil
//}

// checkTimeout iterates over each service
// and checks if its state is deprecated.
func (s *RegistryManager) checkTimeout() {
	s.services.Range(func(key, value any) bool {
		service := value.(*registry)
		if service.state != Offline && service.timeout() {
			//save old state
			old := *service

			//force offline
			service.offline()

			//notify
			s.notifyWatched(&old, service.toStatus())
		}

		return true
	})

	s.timer.Reset(s.duration)
}

type RegistryOption func(*RegistryManager)

func WithTimeoutCheckDuration(d time.Duration) RegistryOption {
	return func(m *RegistryManager) {
		m.duration = d
	}
}

// NewRegistryManager creates a service registry, which itself is also a service,
// and nil is returned if the meta service creation failed.
func NewRegistryManager(registry string, opts ...RegistryOption) *RegistryManager {
	regMgr := NewMetaService(&Descriptor{
		Name:     "registry-manager",
		Registry: registry,
	})
	if regMgr == nil {
		log.Errorln("create registry manager failed")
		return nil
	}

	s := &RegistryManager{
		MetaService: regMgr,
		duration:    5 * time.Second, // default
		//watchers:    make(map[string][]*Watcher),
	}

	for _, fn := range opts {
		fn(s)
	}

	log.Infoln("registry manager created")

	return s
}
