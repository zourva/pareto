package res

// public bus topics
const (
	// NodeJoin is published by s1 server when a new node joined the s1 network
	// and can be subscribed by any service who's interested in it.
	NodeJoin = "/s1-server/node/join"

	// NodeLeave is published by s1 server when a node left the s1 network
	// and can be subscribed by any service who's interested in it.
	NodeLeave = "/s1-server/node/leave"
)
