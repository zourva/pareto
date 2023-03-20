package mod

import (
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/ntop"
	"github.com/zourva/pareto/res"
	"sync"
)

// ServiceManager provides inter-service communication channels
// and manages all services implementing the Service interface.
type ServiceManager interface {
	ntop.Bus
	ntop.RPC

	// Startup starts the manager.
	Startup() error

	// Shutdown stops the manager and all services.
	Shutdown()

	// GetService returns the service associated with the
	// given name or nil if not found.
	GetService(name string) Service

	// AttachService registers a service with the given name and instance.
	// Instance will be substituted if already exists.
	AttachService(name string, svc Service)

	// DetachService detaches the service, identified by the name, from the manager.
	// Detached services will not be joined when service manager exits.
	// Does nothing when the service is not found.
	DetachService(name string)

	// Detached returns true if the service is detached from the manager already.
	// return false when the service is not found.
	Detached(name string) bool
}

type serviceDescriptor struct {
	service  Service
	detached bool
	join     chan int
}

// ServiceManagerImpl implements ServiceManager.
type ServiceManagerImpl struct {
	//low level bus & rpc service
	ntop.Bus
	ntop.RPC

	//composite a base service
	*BaseService

	//services repository
	services sync.Map

	//join for all
	done chan struct{}
}

// GetService returns the service associated with the
// given name or nil if not found.
func (s *ServiceManagerImpl) GetService(name string) Service {
	if sd, ok := s.services.Load(name); ok {
		return sd.(*serviceDescriptor).service
	}

	return nil
}

// Detached returns true if the service is detached from the manager already.
// return false when the service is not found.
func (s *ServiceManagerImpl) Detached(name string) bool {
	if sd, ok := s.services.Load(name); ok {
		return sd.(*serviceDescriptor).detached
	}

	return false
}

// DetachService detaches the service, identified by the name, from the manager.
// Detached services will not be joined when service manager exits.
// Does nothing when the service is not found.
func (s *ServiceManagerImpl) DetachService(name string) {
	if svc, ok := s.services.Load(name); ok {
		s.services.Store(name, &serviceDescriptor{
			service:  svc.(*serviceDescriptor).service,
			detached: true,
		})
	}
}

// AttachService registers a service with the given name and instance.
// Instance will be substituted if already exists.
func (s *ServiceManagerImpl) AttachService(name string, svc Service) {
	s.services.Store(name, &serviceDescriptor{
		service:  svc,
		detached: false,
	})
}

// Shutdown stops the manager and all services.
func (s *ServiceManagerImpl) Shutdown() {
	//notify all attached services
	_ = s.Notify(res.ServiceStop)

	//wait for all attached services to quit,
	//making a graceful shutdown
	if s.serviceCount() == 0 {
		close(s.done)
	} else {
		select {
		case <-s.done:
			log.Infoln("all attached services quit")
			break
		}
	}

	log.Infoln("service manager shutdown")
}

// Startup starts the manager.
func (s *ServiceManagerImpl) Startup() error {
	// init services
	s.services.Range(func(key, value interface{}) bool {
		name := key.(string)
		sd := value.(*serviceDescriptor)

		if !sd.service.OnInitialize() {
			log.Errorf("initialization failed for service %s", name)
			return false
		}

		log.Infoln("initialization done for service", name)
		return true
	})

	//enable service join when exiting
	if err := s.Listen(res.ServiceDown, s.onServiceDown); err != nil {
		return err
	}

	//transfer control to services themselves
	if err := s.Notify(res.ServiceStart); err != nil {
		return err
	}

	log.Infoln("service manager started")

	return nil
}

func (s *ServiceManagerImpl) serviceCount() int {
	var alive = 0
	s.services.Range(func(key, value interface{}) bool {
		alive++
		return true
	})

	return alive
}

func (s *ServiceManagerImpl) onServiceDown(serviceName string) {
	log.Infof("[%s] quit acknowledged", serviceName)

	service := s.GetService(serviceName)
	if service != nil {
		service.OnDestroying()

		s.services.Delete(serviceName)
	}

	if s.serviceCount() == 0 {
		close(s.done)
	}
}

// NewServiceManager creates a new service manager impl.
func NewServiceManager() ServiceManager {
	sm := &ServiceManagerImpl{}

	sm.BaseService = NewBaseService("service manager", sm)
	sm.Bus = ntop.NewBus()
	sm.RPC = ntop.NewRPC()
	sm.done = make(chan struct{})

	log.Infoln("service manager created")

	return sm
}
