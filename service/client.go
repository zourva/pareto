package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/zourva/pareto/ipc"
)

// Registerer acts as a registry delegator,
// helping service instances
// interacting with service server.
// Registerer implements s1 interface.
type Registerer struct {
	//ref to the messager of the service
	messager *ipc.Messager
}

// Register registers the service to the registry server.
// Returns false when any error occurs, and true otherwise.
func (r *Registerer) Register(s Service) bool {
	_, err := msgpack.Marshal(
		map[string]interface{}{
			"name": s.Name(),
		})
	if err != nil {
		return false
	}

	//r.messager.CallV2("/ew1/service/register", b, time.Second)

	return true
}

// Deregister de-registers the service from the registry server.
func (r *Registerer) Deregister(s Service) {
	//r.messager.CallV2("/ew1/service/deregister", []byte(s.Name()), time.Second)
}

// NewRegisterer creates a new registerer with the
// given messager as its communication channel.
func NewRegisterer(m *ipc.Messager) *Registerer {
	r := &Registerer{messager: m}
	log.Infoln("a new registerer is created")
	return r
}
