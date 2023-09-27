package res

// public bus topics
const (
	// NodeJoin is published by s1 server when a new node joined the s1 network
	// and can be subscribed by any service who's interested in it.
	NodeJoin = "/s1-server/node/join"

	// NodeLeave is published by s1 server when a node left the s1 network
	// and can be subscribed by any service who's interested in it.
	NodeLeave = "/s1-server/node/leave"

	// ServiceStart is published by service manager, and is
	// expected to be subscribed by each service,
	// telling each service to start running.
	ServiceStart = "/service-manager/service/start"

	// ServiceStop is published by service manager, and is
	// expected to be subscribed by each service,
	// telling each service to stop running.
	ServiceStop = "/service-manager/service/stop"

	// ServiceDown is required to be published by each service after
	// stopped bue before quit, and is subscribed by the service
	// manager to do the cleaning.
	ServiceDown = "/service-manager/service/down"
)

const (
	ServiceStatusReportInterval = 5
)
