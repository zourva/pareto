package node

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

// RaftRole defines role type of raft.
type RaftRole string

const (
	// RaftRoleLeader the raft leader role.
	RaftRoleLeader    RaftRole = "leader"

	// RaftRoleCandidate the raft candidate role.
	RaftRoleCandidate RaftRole = "candidate"

	// RaftRoleFollower the raft follower role.
	RaftRoleFollower  RaftRole = "follower"
)

// VoteReq defines request of raft AppendEntries Method,
// and is invoked by candidates to gather votes (§5.2).
type VoteReq struct {
	Term         uint64 `json:"term"`         //candidate’s term
	CandidateID  string `json:"candidateId"`  //candidate id requesting vote
	LastLogIndex uint64 `json:"lastLogIndex"` //index of candidate’s last log entry (§5.4)
	LastLogTerm  uint64 `json:"lastLogTerm"`  //term of candidate’s last log entry (§5.4)
}

// VoteRsp defines response of raft RequestVote Method.
type VoteRsp struct {
	Term        uint64 `json:"term"`        //currentTerm, for candidate to update itself
	VoteGranted bool   `json:"voteGranted"` //true means candidate received vote
}

// AppendReq defines request of raft AppendEntries Method,
// and is invoked by leader to replicate logs & to piggyback heartbeat.
type AppendReq struct {
	Term         uint64        //leader’s term
	LeaderID     string        //leader id so follower can redirect clients
	PrevLogIndex uint64        //index of log entry immediately preceding new ones
	PrevLogTerm  uint64        //term of prevLogIndex entry
	Entries      []interface{} //log entries to store (empty for heartbeat; may send more than one for efficiency)
	LeaderCommit uint64        // leader’s commitIndex
}

// AppendRsp defines response of raft AppendEntries Method.
type AppendRsp struct {
	Term    uint64 //currentTerm, for leader to update itself
	Success bool   //true if follower contained entry matching prevLogIndex and prevLogTerm
}

// State includes:
//  persistent state on all servers,
//  volatile state on all servers, and
//  volatile state on leaders (which will be reinitialized after election).
type State struct {
	//latest term server has seen (initialized to 0
	//on first boot, increases monotonically)
	CurrentTerm uint64 `json:"currentTerm"`

	//candidateId that received vote in current
	//term (or null if none)
	VotedFor string `json:"votedFor"`

	//log entries; each entry contains command
	//for state machine, and term when entry
	//was received by leader (first index is 1)
	Log []interface{} `json:"log"`

	//index of highest log entry known to be
	//committed (initialized to 0, increases monotonically)
	CommitIndex uint64 `json:"commitIndex"`

	//index of highest log entry applied to state
	//machine (initialized to 0, increases monotonically)
	LastApplied uint64 `json:"lastApplied"`

	//for each server, index of the next log entry
	//to send to that server (initialized to leader last log index + 1)
	NextIndex []uint64 `json:"nextIndex"`

	//for each server, index of highest log entry
	//	known to be replicated on server
	//(initialized to 0, increases monotonically)
	MatchIndex []uint64 `json:"matchIndex"`
}

// ReplicationService defines log-replication-related operations.
type ReplicationService interface {
	AppendEntries(req *AppendReq, rsp *AppendRsp) error
}

// ElectionService defines leader-election-related operations.
type ElectionService interface {
	RequestVote(req *VoteReq, rsp *VoteRsp) error
	IncrementTerm()
	VoteForMySelf()
	BroadcastVote()
	WinTheElection() bool
}

// RaftService defines services a raft-based cluster should expose
type RaftService interface {
	ReplicationService
	ElectionService
}

type consensus struct {
	node  *ClusterNode //ref to node owner
	mutex sync.Mutex
	role  RaftRole
	quit  chan bool

	interval  uint64 //heartbeat interval in milliseconds
	aliveTime uint64 //latest aliveness update time
	duration  uint64 //duration to wait before start an election

	roleServices map[RaftRole]func() bool

	// valid iff this node becomes the leader
	replicator *replicator
}

func newConsensus(node *ClusterNode) *consensus {
	var defaultDuration uint64 = 5
	c := &consensus{
		node:         node,
		role:         RaftRoleFollower,
		roleServices: make(map[RaftRole]func() bool),
		quit:         make(chan bool),
		interval:     defaultDuration,
		aliveTime:    0,
		duration:     2 * defaultDuration,
	}

	c.roleServices[RaftRoleFollower] = c.runFollowerLoop
	c.roleServices[RaftRoleCandidate] = c.runCandidateLoop
	c.roleServices[RaftRoleLeader] = c.runLeaderLoop

	return c
}

func (c *consensus) RequestVote(req *VoteReq, rsp *VoteRsp) error {
	panic("implement me")
}

func (c *consensus) AppendEntries(req *AppendReq, rsp *AppendRsp) error {
	panic("implement me")
}

// IncrementTerm increases the current term.
func (c *consensus) IncrementTerm() {

}

// VoteForMySelf votes a ticket for myself.
func (c *consensus) VoteForMySelf() {

}

// BroadcastVote broadcasts vote request to known peers.
func (c *consensus) BroadcastVote() {
	for _, peer := range c.node.peers {
		p := peer
		go func() {
			rsp, err := c.rpcCall(
				p.address,
				"ElectionService.RequestVote",
				&VoteReq{
					//c.curTerm,
					CandidateID: c.node.id,
					//c.lastLogIndex,
					//c.lastLogTerm,
				})
			if err != nil {
				log.Errorln("RequestVote failed:", err)
			}

			rsp = rsp.(*VoteRsp)

		}()
	}
}

// WinTheElection checks if this node(as a candidate) wins the election. 
func (c *consensus) WinTheElection() bool {
	//TODO
	return true
}

func (c *consensus) rpcCall(ep string, name string, req interface{}) (interface{}, error) {
	conn, err := rpc.DialHTTP("tcp", ep)
	if err != nil {
		log.Errorf("rpc client connect to %s failed: %v", ep, err)
		return nil, err
	}

	defer conn.Close()

	var rsp interface{}
	err = conn.Call(name, req, rsp)
	if err != nil {
		log.Errorf("rpc client call %s failed: %v", name, err)
		return nil, err
	}

	log.Debugf("rpc client call %s with rsp: %v", name, rsp)

	return rsp, nil
}

func (c *consensus) become(role RaftRole) {
	log.Debugf("cluster peer changes role from %v to %v", c.role, role)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.role = role
}

func (c *consensus) getRole() RaftRole {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.role
}

func (c *consensus) aliveTimeout() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return box.TimeNowMs()-c.aliveTime > c.duration
}

func (c *consensus) stop() {
	// close to quit all loops selects on channel
	close(c.quit)
}

func (c *consensus) run() {
	for {
		//select {
		//case <-c.quit:
		//	log.Infoln("consensus manager quit")
		//	return
		//default:
		//}

		role := c.getRole()
		log.Debugf("current role:%s", role)

		if !c.roleServices[role]() {
			log.Infoln("consensus manager quit")
			break
		}
	}
}

//To begin an election, a follower increments its current
//term and transitions to candidate state.
//	It then votes for itself and issues RequestVote RPCs in parallel to each of
//	the other servers in the cluster.
//
//A candidate continues in this state until one of three things happens:
//	(a) it wins the election,
//	(b) another server establishes itself as leader, or
//	(c) a period of time goes by with no winner.
func (c *consensus) runCandidateLoop() bool {
	duration := time.Duration(150+rand.Intn(150)) * time.Millisecond
	t := time.NewTimer(duration)

	for {
		//check if it wins the election
		if c.WinTheElection() {
			c.become(RaftRoleLeader)
			t.Stop()
		}

		select {
		case <-c.quit:
			log.Infoln("quit candidate loop")
			return false
		case <-t.C:
			if c.getRole() != RaftRoleCandidate {
				log.Infoln("change of role detected, skip current candidate loop")
				return true
			}

			c.IncrementTerm()
			c.VoteForMySelf()
			c.BroadcastVote()

			t.Reset(duration)
		}
	}
}

func (c *consensus) runLeaderLoop() bool {
	return false
}

func (c *consensus) runFollowerLoop() bool {
	t := time.NewTimer(time.Duration(c.interval) * time.Millisecond)

	for {
		select {
		case <-c.quit:
			log.Infoln("quit follower loop")
			return false
		case <-t.C:
			if c.getRole() != RaftRoleFollower {
				log.Infoln("change of role detected, skip current follower loop")
				return true
			}

			if c.aliveTimeout() {
				log.Infoln("heartbeat timeout, start an election")
				c.become(RaftRoleCandidate)
			}

			t.Reset(time.Duration(c.interval) * time.Millisecond)
		}
	}
}

type replicator struct {
}

type rpcServer struct {
	network  string      //network type to listen on
	endpoint string      //endpoint address to listen
	service  RaftService //rpc services to register
}

func newRPCServer(ep string, service RaftService) *rpcServer {
	return &rpcServer{
		network:  "tcp",
		endpoint: ep,
		service:  service,
	}
}

func (s *rpcServer) run() {
	err := rpc.Register(s.service)
	if err != nil {
		log.Fatalln("rpc server register failed:", err)
	}

	rpc.HandleHTTP()

	lis, err := net.Listen(s.network, s.endpoint)
	if err != nil {
		log.Fatalln("rpc server listen failed:", err)
	}

	if err := http.Serve(lis, nil); err != nil {
		log.Fatalln("rpc server serve failed:", err)
	}
}

// Peer defines the maintained meta
// info of other nodes by a cluster node.
type Peer struct {
	synced  bool   //true if get in touched
	address string //rpc endpoint of a peer
}

// ClusterNode as a cluster node
type ClusterNode struct {
	//basic rpc service
	rpcServer *rpcServer

	//consensus manager of this node
	consensus *consensus

	//peers, discovered from initial seed node,
	//seen by this node, including this node itself
	peers map[string]*Peer

	//identity of this node
	id string
}

func newClusterNode() *ClusterNode {
	n := &ClusterNode{}
	n.consensus = newConsensus(n)
	n.rpcServer = newRPCServer(":33785", n.consensus)

	return n
}

func (p *ClusterNode) running() bool {
	// TODO
	return false
}

// start consensus manager
func (p *ClusterNode) run() error {
	if p.running() {
		return fmt.Errorf("server is running")
	}

	go p.rpcServer.run()

	go p.consensus.run()

	return nil
}

//
//func (p *ClusterNode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	switch r.URL.Path {
//	case "/appendEntries":
//		p.appendEntriesHandler(w, r)
//
//	case "/requestVote":
//		p.requestVoteHandler(w, r)
//
//	case "/clientApply":
//		p.clientApplyHandler(w, r)
//
//	default:
//		w.WriteHeader(http.StatusNotFound)
//	}
//}
//
//func (p *ClusterNode) appendEntriesHandler(w http.ResponseWriter, r *http.Request) {
//
//}
//
//func (p *ClusterNode) requestVoteHandler(w http.ResponseWriter, r *http.Request) {
//
//}
//
//func (p *ClusterNode) clientApplyHandler(w http.ResponseWriter, r *http.Request) {
//
//}
