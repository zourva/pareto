package node

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"github.com/zourva/pareto/box/env"
	"github.com/zourva/pareto/box/meta"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/stats"
	"time"
)

// AgentSideCallback the callback of hooks.
type AgentSideCallback func()

// AgentSideHooks defines callbacks agent exposed
type AgentSideHooks struct {
	//called after initialized
	OnInitialized AgentSideCallback

	//called after successfully authed
	OnAuthenticated AgentSideCallback

	//called when preparing or re-preparing finished
	OnMaintained AgentSideCallback

	//called when stream message received
	OnServerNasMsg func(*StreamMessage)

	//called after stopped
	OnStopped AgentSideCallback
}

// AgentOption func-closure pattern
type AgentOption func(agent *Agent)

// agentOptions used by Agent
type agentOptions struct {
	endpoint  string //server endpoint
	clientID  string //conn id assigned by RegistryServer
	interval  uint32 //status report interval, in milliseconds
	threshold uint32 //threshold to rebuild underlying connection
	callbacks AgentSideHooks
}

func defaultAgentOptions() agentOptions {
	return agentOptions{
		endpoint:  connectEndpoint,
		clientID:  emptyString,
		interval:  defInterval,
		threshold: 3,
		callbacks: AgentSideHooks{},
	}
}

// WithStatusReportInterval sets status report interval to the given value, in milliseconds.
func WithStatusReportInterval(interval uint32) AgentOption {
	return func(agent *Agent) {
		agent.options.interval = box.ClampU32(minInterval, maxInterval, interval)
	}
}

// WithClientID sets agent identity to id.
func WithClientID(id string) AgentOption {
	return func(agent *Agent) {
		agent.options.clientID = id
	}
}

// WithThreshold provides a retry threshold,
// which will result to rebuild underlying connection
// if the number of internal failures exceeds it.
func WithThreshold(t uint32) AgentOption {
	return func(agent *Agent) {
		agent.options.threshold = t
	}
}

// WithCallbacks provides agent side hooks.
func WithCallbacks(cbs AgentSideHooks) AgentOption {
	return func(agent *Agent) {
		agent.options.callbacks = cbs
	}
}

// Agent models node of the terminal side.
type Agent struct {
	*meta.StateMachine[string]
	//*service.MetaService
	options   agentOptions
	configMgr AgentConfManager
	protoMgr  *AgentProto

	// grpc underlying connection
	clientConn *grpc.ClientConn

	clientID string //conn id assigned by RegistryServer

	// statistics
	msgCount int64
	failures int64
}

// NewAgent creates an agent with the given endpoint address of the server and options.
func NewAgent(endpoint string, opts ...AgentOption) *Agent {
	if !box.ValidateEndpoint(endpoint) {
		return nil
	}

	c := &Agent{
		StateMachine: meta.NewStateMachine[string](agentStateMachine, time.Second),
		//MetaService: service.NewMetaService(&service.Config{
		//	Name: agentStateMachine,
		//	Messager: ipc.NewMessager(&ipc.MessagerConf{
		//		BusConf: &ipc.BusConf{Name: "agent-bus", Type: ipc.InterProcBus, Broker: endpoint},
		//		RpcConf: &ipc.RPCConf{Name: "agent-rpc", Type: ipc.InterProcRpc, Broker: endpoint},
		//	}),
		//	Registerer: nil,
		//}),
		options:   defaultAgentOptions(),
		configMgr: NewAgentConfManager(env.GetExecFilePath() + "/../etc/conf.db"),
		protoMgr:  nil,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.RegisterStates([]*meta.State[string]{
		{Name: authenticating, Action: c.onAuthenticating},
		{Name: maintaining, Action: c.onMaintaining},
		{Name: servicing, Action: c.onServicing},
		{Name: restarting, Action: c.onRestarting},
		{Name: stoppingState, Action: c.onStopping},
		//{Name: stoppedState, Action: c.onStopped},
	})

	c.SetStartingState(authenticating)
	c.SetStoppingState(stoppingState)

	log.Infoln("new node agent with endpoint", endpoint)

	return c
}

// TagRPC not used yet
func (a *Agent) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC not used yet
func (a *Agent) HandleRPC(ctx context.Context, rpcStats stats.RPCStats) {
	//
}

// TagConn not used yet
func (a *Agent) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn the hook, which handles state
// migration according to the underlying connection state change.
func (a *Agent) HandleConn(ctx context.Context, connStats stats.ConnStats) {
	switch connStats.(type) {
	case *stats.ConnBegin:
		log.Infoln("agent connection begin...")
		if a.GetState() == servicing {
			log.Infoln("agent connection begin caused by repairing")
			a.MoveToState(restarting)
		}
	case *stats.ConnEnd:
		log.Infoln("agent connection end...")
	default:
		log.Infoln("illegal conn callback type")
	}
}

// Start starts the agent by connecting, creating a service client,
// establishing the client side protocol stack and moving to starting state.
// Returns err if failed.
func (a *Agent) Start() error {
	client, err := grpc.Dial(
		a.options.endpoint,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithStatsHandler(a),
		grpc.WithKeepaliveParams(
			keepalive.ClientParameters{
				Time:                2 * time.Hour,
				Timeout:             20 * time.Second,
				PermitWithoutStream: true,
			},
		),
	)

	if err != nil {
		log.Errorln("agent connection failed", err)
		return err
	}

	service := NewS1ServiceClient(client)
	if service == nil {
		log.Errorln("create agent service client failed", err)
		return err
	}

	a.clientConn = client
	a.protoMgr = NewAgentProto(a, service)

	if a.options.callbacks.OnInitialized != nil {
		a.options.callbacks.OnInitialized()
	}

	a.Startup()

	log.Infoln("node agent started")

	return nil
}

// Stop stops the agent.
func (a *Agent) Stop() {
	a.Shutdown()

	_ = a.clientConn.Close()

	log.Infof("node agent stopped")
}

// onStopping signs out from server and quit agent.
func (a *Agent) onStopping(args interface{}) {
	a.protoMgr.doSignOut()
	log.Infoln("agent stopping done")

	if a.options.callbacks.OnStopped != nil {
		a.options.callbacks.OnStopped()
	}
}

func (a *Agent) onRestarting(args interface{}) {
	// pause and wait
	// TODO: pause and wait

	// restart
	log.Infoln("agent restart preparation done, try re-authenticating")

	a.MoveToState(authenticating)
}

func (a *Agent) onServicing(args interface{}) {
	log.Traceln("agent servicing...")
}

// onMaintaining handles actions happened during maintaining phase,
// i.e, parameter reconfiguration/version upgrade
func (a *Agent) onMaintaining(args interface{}) {
	if !a.protoMgr.doConfigOta() {
		log.Infoln("agent configuration failed, skip")
	}

	if !a.protoMgr.doVersionOta() {
		log.Infoln("agent configuration failed, skip")
	}

	a.protoMgr.enableWatcher()

	log.Infoln("agent maintaining done")

	if a.options.callbacks.OnMaintained != nil {
		a.options.callbacks.OnMaintained()
	}

	a.MoveToState(servicing)
}

func (a *Agent) onAuthenticating(args interface{}) {
	if a.configMgr.NeedProvision() {
		if !a.protoMgr.doSignUp() {
			log.Errorln("agent SignUp failed")
			return
		}
	}

	// always load from config
	a.clientID = a.configMgr.GetConf().Identity

	if !a.protoMgr.doSignIn() {
		log.Errorln("agent SignIn failed")
		return
	}

	log.Infoln("agent authentication done")

	if a.options.callbacks.OnAuthenticated != nil {
		a.options.callbacks.OnAuthenticated()
	}

	a.MoveToState(maintaining)
}

func (a *Agent) onStreamMsg(msg *StreamMessage) {
	log.Debugln("onStreamMsg callback:", *msg)
	if a.options.callbacks.OnServerNasMsg != nil {
		a.options.callbacks.OnServerNasMsg(msg)
	}
}
