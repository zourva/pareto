package service

type Watcher struct {
	//watch specification
	spec WatchSpec
}

var defaultWatchedStates = []State{
	Servicing, Stopped,
}

type WatchSpec struct {
	//NotifyType PS or RR

	//Watched is the name of service watched.
	//Mandatory.
	Watched string

	//Watcher is the name of watching service.
	//Optional, will be overwritten by service client.
	Watcher string

	//Channel name to send notify when states of watched service are detected.
	//If not provided or empty, notice is ignored.
	//Channel string

	//TargetStates defines a subset states to observe,
	//if not provided, Servicing and Stopped are set.
	TargetStates []State
}
