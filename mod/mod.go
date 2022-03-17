package mod

import (
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/ntop"
	"reflect"
)

type Service interface {
	// GetName returns the name of service
	GetName() string

	// GetStatus returns collected data from each service
	// with a format of protocol message
	GetStatus() interface{}

	//GetCapabilities
	GetCapabilities() interface{}

	// Detach detaches the service from its manager.
	// Services are attached by default when creating,
	// and can be attached again when after detached.
	Detach()

	// Detached returns true if the service is detached from its manager.
	Detached() bool

	//OnInitialize is called after the service is created
	OnInitialize() bool

	//OnDestroying is called before the service is destroyed
	OnDestroying()
}

// BaseService implements the Service interface and
// provides a bunch of methods for inheritance.
type BaseService struct {
	// name of the service
	name string

	// the manager that manages this service
	mgr ServiceManager
}

// NewBaseService creates a new base service with the given name.
// If mgr is not nil, the newly created service is attached automatically to the mgr.
func NewBaseService(name string, mgr ServiceManager) *BaseService {
	s := &BaseService{
		name: name,
		mgr:  mgr,
	}

	return s
}

func (s *BaseService) GetName() string {
	return s.name
}

func (s *BaseService) Detached() bool {
	return s.mgr.Detached(s.name)
}

func (s *BaseService) Detach() {
	s.mgr.DetachService(s.name)
}

func (s *BaseService) OnInitialize() bool {
	log.Infoln("initializing service", s.name)
	return true
}

func (s *BaseService) OnDestroying() {
	log.Infoln("destroying service", s.name)
}

func (s *BaseService) GetStatus() interface{} {
	return ""
}

func (s *BaseService) GetCapabilities() interface{} {
	return ""
}

// Listen binds a handler to a subscribed topic.
// Old handler will be replaced if already bounded.
func (s *BaseService) Listen(topic string, fn interface{}) error {
	log.Infof("%s subscribe to %s", s.name, topic)
	return s.mgr.Subscribe(topic, fn)
}

// Notify broadcasts a notice message to all subscribers and assumes no replies.
func (s *BaseService) Notify(topic string, args ...interface{}) error {
	log.Debugf("%s publish to %s", s.name, topic)
	s.mgr.Publish(topic, args...)

	return nil
}

// ExposeMethod expose a server-side method to the external world.
func (s *BaseService) ExposeMethod(id ntop.RpcMethod, fn interface{}) {
	name := id.SerializedName()

	log.Debugf("%s expose method %s", s.name, name)

	s.mgr.Expose(name, fn)
}

// CallMethod calls a remote method identified by id.
func (s *BaseService) CallMethod(id ntop.RpcMethod, args ...interface{}) (reflect.Value, error) {
	name := id.SerializedName()
	log.Debugf("%s invoke rpc %s", s.name, name)

	return s.mgr.Call(name, args...)
}
