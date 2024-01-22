package service

type StatusConf struct {
	//Interval, in seconds, to refresh and publish service status,
	//optional with a minimum value of 1 second.
	//If not provided, or set to 0, status publishing is disabled.
	Interval uint32 `json:"interval"`

	//Threshold number of intervals before out-of-sync
	Threshold uint32 `json:"threshold"`

	//Endpoint used to export service status periodically, optional.
	//If not provided, the default format is used: {service name}/status.
	//Changed: use centered topic to aggregate service status.
	//endpoint string
}

// Status defines heartbeat info published by a service.
type Status struct {
	Name  string `json:"name"`  //name of the service
	State State  `json:"state"` //liveness state of the service
	Time  uint64 `json:"time"`  //report timestamp in milliseconds
	Ready bool   `json:"ready"` //readiness state of the service

	Metrics any `json:"metrics,omitempty"` //detail metrics, optional

	//Conf *StatusConf `json:"health,omitempty"`
	//if provided, overwrites the default timeout check interval(5000ms)
	CheckInterval uint32 `json:"checkInterval,omitempty"`

	//if provided, overwrites the default failure count(3 times)
	//allowed before treating a service as offline
	AllowFailures uint32 `json:"allowFailures,omitempty"`
}

// StatusList defines all services status info.
type StatusList struct {
	Services []*Status `json:"services"`
}

func getDefaultStatusConf() *StatusConf {
	return &StatusConf{
		Interval:  StatusReportInterval,
		Threshold: StatusLostThreshold,
		//endpoint:  EndpointServiceStatus,
	}
}

//func getDefaultStatusEndpoint(name string) string {
//	return fmt.Sprintf("%s/status", name)
//}
