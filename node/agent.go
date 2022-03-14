package node

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"github.com/zourva/pareto/env"
	"github.com/zourva/pareto/meta"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/stats"
	"time"
)

type Callback func()

// LifecycleHooks defines callbacks agent exposed
type LifecycleHooks struct {
	//called after initialized
	OnInitialized Callback

	//called after successfully authed
	OnAuthenticated Callback

	//called when preparing or re-preparing finished
	OnMaintained Callback

	//called when stream message received
	OnNasMessage func(*StreamMessage)

	//called after stopped
	OnStopped Callback
}

// option func-closure pattern
type AgentOption func(agent *Agent)

// agentOptions used by Agent
type agentOptions struct {
	endpoint  string //server endpoint
	clientId  string //conn id assigned by Server
	interval  uint32 //status report interval, in milliseconds
	threshold uint32 //threshold to rebuild underlying connection
	callbacks LifecycleHooks
}

func defaultAgentOptions() agentOptions {
	return agentOptions{
		endpoint:  connectEndpoint,
		clientId:  emptyString,
		interval:  defInterval,
		threshold: 3,
		callbacks: LifecycleHooks{},
	}
}

func WithStatusReportInterval(interval uint32) AgentOption {
	return func(agent *Agent) {
		agent.options.interval = box.ClampU32(minInterval, maxInterval, interval)
	}
}

func WithClientId(id string) AgentOption {
	return func(agent *Agent) {
		agent.options.clientId = id
	}
}

func WithThreshold(t uint32) AgentOption {
	return func(agent *Agent) {
		agent.options.threshold = t
	}
}

func WithCallbacks(cbs LifecycleHooks) AgentOption {
	return func(agent *Agent) {
		agent.options.callbacks = cbs
	}
}

// Agent models node of the terminal side.
type Agent struct {
	*meta.StateMachine
	options   agentOptions
	configMgr AgentConfManager
	protoMgr  *AgentProto

	// grpc underlying connection
	clientConn *grpc.ClientConn

	clientId string //conn id assigned by Server

	// statistics
	msgCount int64
	failures int64
}

func NewAgent(endpoint string, opts ...AgentOption) *Agent {
	if !box.ValidateEndpoint(endpoint) {
		return nil
	}

	c := &Agent{
		StateMachine: meta.NewStateMachine(agentStateMachine, time.Second),
		options:      defaultAgentOptions(),
		configMgr:    NewAgentConfManager(env.GetExecFilePath() + "/../etc/conf.db"),
		protoMgr:     nil,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.RegisterStates([]*meta.State{
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

func (a *Agent) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}

func (a *Agent) HandleRPC(ctx context.Context, rpcStats stats.RPCStats) {
	//
}

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

	a.Startup()

	log.Infoln("node agent started")

	return nil
}

func (a *Agent) Stop() {
	a.Shutdown()

	_ = a.clientConn.Close()

	log.Infof("node agent stopped")
}

//onStopping signs out from server and quit agent.
func (a *Agent) onStopping(args interface{}) {
	a.protoMgr.doSignOut()
	log.Infoln("agent stopping done")
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
	a.clientId = a.configMgr.GetConf().Identity

	if !a.protoMgr.doSignIn() {
		log.Errorln("agent SignIn failed")
		return
	}

	log.Infoln("agent authentication done")
	a.MoveToState(maintaining)
}

func (a *Agent) onStreamMsg(msg *StreamMessage) {
	log.Debugln("onStreamMsg callback:", *msg)
	if a.options.callbacks.OnNasMessage != nil {
		a.options.callbacks.OnNasMessage(msg)
	}
}
