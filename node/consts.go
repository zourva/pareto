package node

const (
	authenticating = "initial"
	maintaining    = "preparing"
	servicing      = "servicing"
	restarting     = "restarting"
	stoppingState  = "stopping"
	//stoppedState   = "stopped"
)

const (
	listenEndpoint  = ":21985"
	connectEndpoint = "127.0.0.1:21985"
	tcpNetwork      = "tcp"
	emptyString     = ""

	agentStateMachine = "Agent"
)

// to suppress golint
type contextKey string

const (
	sessionKeyID contextKey = "sessionKey"
	clientKeyID  contextKey = "clientID"
)

const (
	minInterval = 10        //10 milliseconds
	maxInterval = 10 * 1000 //10 seconds
	defInterval = 1000      //1 second
)

const (
	// DES algorithm
	DES uint32 = 1

	// AES algorithm
	AES uint32 = 2

	// GM4 Guo Mi 4 algorithm
	GM4 uint32 = 3
)
