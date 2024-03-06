package service

import "fmt"

const (
	Registry = "registry"
)

const (
	StatusReportInterval = 5 //seconds
	StatusLostThreshold  = 3 //3 intervals to wait before treat service as offline
	StatusCheckInterval  = 5 //seconds
	StatusQueryTimeout   = 2 //seconds
	ReviveWaitThreshold  = 3 //another 3 intervals to wait before purge offline services
)

const (
	//EndpointServiceInfo is bound as an RR endpoint accepting
	//service registration and querying request.
	EndpointServiceInfo = "/registry-center/service/info"

	//EndpointServiceStatus is bound as a PS endpoint accepting only
	//service status periodically reported.
	EndpointServiceStatus = "/registry-center/service/status"

	//EndpointServiceNotice is bound as a PS endpoint publishing only
	//service status changing events.
	EndpointServiceNotice = "/registry-center/service/notice"
)

const (
	Register        = "Register"
	ReportStatus    = "ReportStatus"
	QueryStatus     = "QueryStatus"
	QueryStatusList = "QueryStatusList"
)

type RegisterReq struct {
	Name          string `json:"name"`
	State         State  `json:"state"`
	Ready         bool   `json:"ready"`
	CheckInterval uint32 `json:"checkInterval,omitempty"`
	AllowFailures uint32 `json:"allowFailures,omitempty"`
}

type RegisterRsp struct {
}

type ReportStatusReq struct {
	Status *Status `json:"status"`
}

type ReportStatusRsp struct {
}

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
