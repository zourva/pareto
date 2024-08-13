package service

import (
	log "github.com/sirupsen/logrus"
	"sync"
)

// Monitor monitors status of services
// and that of some infrastructure, and
// report alerts if required.
type Monitor struct {
	registry *RegistryManager
}

// GetStatus returns status of the given service and nil
// if the service of the given name does not exist.
func (m *Monitor) GetStatus(name string) *Status {
	reg := m.registry.get(name)
	if reg == nil {
		return nil
	}

	return &Status{
		Name:   reg.name,
		Domain: reg.domain,
		State:  reg.state,
		Time:   reg.updateTime,
		Ready:  reg.ready,
	}
}

// GetStatusList returns full list of status of managed services.
func (m *Monitor) GetStatusList() StatusList {
	var list StatusList
	all := m.registry.all()
	for _, reg := range all {
		list.Services = append(list.Services, &Status{
			Name:   reg.name,
			Domain: reg.domain,
			State:  reg.state,
			Time:   reg.updateTime,
			Ready:  reg.ready,
		})

	}

	return list
}

// GetNotServicing returns names of services from the given names
// that are not ready yet.
func (m *Monitor) GetNotServicing(filtered []string) []string {
	var result []string
	for _, name := range filtered {
		s := m.registry.get(name)
		if s == nil || s.state != Servicing {
			result = append(result, name)
		}
	}

	return result
}

var monLock sync.Mutex
var monitor *Monitor

// NewMonitor creates an instance of Monitor.
// Monitor itself is not a service, however,
// it needs to inquiry service registry to
// get service status, so it creates a service registry
// internally to manage all services registered.
func NewMonitor(registry string) *Monitor {
	manager := NewRegistryManager(registry)
	if manager == nil {
		return nil
	}

	s := &Monitor{
		registry: manager,
	}

	return s
}

func GetMonitor() *Monitor {
	return monitor
}

// EnableMonitor enables service monitor by creating
// and attaching a service registry manager to the
// given service registry address.
func EnableMonitor(registry string) *Monitor {
	monLock.Lock()
	defer monLock.Unlock()

	if monitor == nil {
		monitor = NewMonitor(registry)
		if monitor == nil {
			log.Fatalln("enable monitor failed")
		}
	}

	Start(monitor.registry)

	return monitor
}

// DisableMonitor disables the service monitor
// if it is enabled already.
func DisableMonitor() {
	if monitor == nil {
		return
	}

	monLock.Lock()
	defer monLock.Unlock()

	Stop(monitor.registry)

	monitor = nil
}
