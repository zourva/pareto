package service

import "fmt"

const (
	Registry = "registry"
)

const (
	StatusReportInterval = 5 //seconds
	StatusLostThreshold  = 3
	StatusCheckInterval  = 5 //seconds
	StatusQueryTimeout   = 2 //seconds
)

const (
	// EndpointServiceStart is published by service manager, and is
	// expected to be subscribed by each service,
	// telling each service to start running.
	//EndpointServiceStart = "/registry-center/service/start"

	// EndpointServiceDown is required to be published by service manager
	// after stopped but before quit, and is subscribed by the service
	// manager to do the cleaning.
	// EndpointServiceDown = "/registry-center/service/down"

	// EndpointServiceStop is published by service manager, and is
	// expected to be subscribed by each service,
	// telling each service to stop running.
	//EndpointServiceStop = "/registry-center/service/stop"

	EndpointServiceNotice = "/registry-center/service/notice"
	EndpointServiceStatus = "/registry-center/service/status"
	EndpointServiceInfo   = "/registry-center/service/info"
)

const (
	QueryStatus     = "QueryStatus"
	QueryStatusList = "QueryStatusList"
)

type QueryStatusReq struct {
	Name string `json:"name"`
}

type QueryStatusRsp struct {
	Status *Status `json:"status"`
}

type QueryStatusListReq struct {
	// names of services observed
	Observed []string `json:"observed"`
}

type QueryStatusListRsp struct {
	List *StatusList `json:"list"`
}

// State defines service liveliness state.
type State int

const (
	Offline   State = iota // stopping -> offline
	Starting               // offline -> starting
	Servicing              // starting -> servicing
	Stopping               // starting/servicing -> stopping
	Stopped                // stopping -> stopped
)

func (s State) String() string {
	switch s {
	case Servicing:
		return "servicing"
	case Starting:
		return "starting"
	case Stopping:
		return "stopping"
	case Offline:
		return "offline"
	case Stopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// Endpoint defines the identity of a bus endpoint or rpc channel.
type Endpoint struct {
	Service string
	Object  string
	Method  string
}

// SerializedName returns the path-like name of the method, i.e.:
//
//	service.object.method
//
// e.g.:
//
//	webserver/cookie/get
func (r *Endpoint) SerializedName() string {
	return fmt.Sprintf("%s/%s/%s", r.Service, r.Object, r.Method)
}
