package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box/meta"
	"github.com/zourva/pareto/ipc"
	"time"
)

type Registrar interface {
	EnableStatusExport()
	DisableStatusExport()
	QueryStatus(name string) *Status
	QueryStatusList(namesWhitelist []string) *StatusList
	StatusList() *StatusList
	Register() bool
	Unregister()
}

// Registrar acts as a registry delegator,
// helping service instances
// interacting with service server.
// Registrar implements s1 interface.
type registrar struct {
	//ref to the service this registrar serves.
	service Service

	exporter meta.Loop

	//status of followed services, used for recovery
	list *StatusList
}

// EnableStatusExport exports status of the service periodically.
func (r *registrar) EnableStatusExport() {
	if r.exporter != nil {
		log.Warnln("already enabled status export")
		return
	}

	r.exporter = meta.NewLoop("registrar", meta.LoopConfig{
		Tick: r.service.StatusConf().Interval * 1000,
	})

	// always report start
	_ = r.report()

	// start the loop
	r.exporter.Run(meta.LoopRunHook{
		Working: func() error {
			return r.report()
		},
	})

}

// DisableStatusExport disables export status of the service.
func (r *registrar) DisableStatusExport() {
	// always report stop
	_ = r.report()

	r.exporter.Stop()
}

// QueryStatus returns status of the given service and nil
// if the service of the given name does not exist.
func (r *registrar) QueryStatus(name string) *Status {
	client := r.service.RpcClient()
	rsp, err := client.Invoke(EndpointServiceInfo, QueryStatus, 2*time.Second, &QueryStatusReq{Name: name})
	if err != nil {
		log.Errorf("query status of service %s failed, %v", name, err)
		return nil
	}

	var qsr QueryStatusRsp
	err = rsp.GetObject(&qsr)
	if err != nil {
		log.Errorf("query status of service %s failed, %v", name, err)
		return nil
	}

	return qsr.Status
}

// QueryStatusList returns list of status of the filtered services.
// If a service is in namesWhitelist while not registered,
// its Status will not be included in the returned list.
// All list is returned if namesWhitelist is nil or its length is 0.
func (r *registrar) QueryStatusList(namesWhitelist []string) *StatusList {
	client := r.service.RpcClient()

	req := &QueryStatusListReq{}
	req.Observed = append(req.Observed, namesWhitelist...)

	rsp, err := client.Invoke(EndpointServiceInfo, QueryStatusList, StatusQueryTimeout*time.Second, req)
	if err != nil {
		log.Warnf("query status list of services %s failed, %v", namesWhitelist, err)
		return nil
	}

	var list QueryStatusListRsp
	err = rsp.GetObject(&list)
	if err != nil {
		log.Errorf("query status of services %s failed, %v", namesWhitelist, err)
		return nil
	}

	return list.List
}

// StatusList returns cached status list copy of recently queried.
func (r *registrar) StatusList() *StatusList {
	return r.list
}

// Register registers the delegated service to the registry server,
// and starts a separate long-running loop to export service status
// periodically when the registration succeeded.
//
// Returns false when any error occurs, and true otherwise.
func (r *registrar) Register() bool {
	//_, err := msgpack.Marshal(
	//	map[string]interface{}{
	//		"name": r.service.Name(),
	//	})
	//if err != nil {
	//	return false
	//}

	if Registry != r.service.Name() {
		// for all non-registry services,
		// try getting all services states for later recovery
		r.list = r.QueryStatusList(nil)
	}

	r.EnableStatusExport()

	return true
}

// Unregister unregisters the service from the registry server.
func (r *registrar) Unregister() {
	//r.messager.CallV2("/ew1/service/deregister", []byte(s.Name()), time.Second)
}

func (r *registrar) report() error {
	err := r.service.Notify(EndpointServiceStatus, r.service.MarshalStatus())
	if err != nil {
		log.Warnf("export status for service %s failed: %v", r.service.Name(), err)
		return err
	}
	return nil
}

// NewRegistrar creates a registrar for a service
// as the delegator of service manager.
func NewRegistrar(s Service) Registrar {
	r := &registrar{
		service: s,
	}

	log.Infof("registrar is created for service %s", s.Name())
	return r
}

// NewRegisterer creates a new registerer with the
// given messager as its communication channel.
// Deprecated use NewRegistrar instead.
func NewRegisterer(m *ipc.Messager) Registrar {
	r := &registrar{}
	log.Infof("registrar is created for messager %p", m)
	return r
}
