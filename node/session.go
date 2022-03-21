package node

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/stats"
	"sync"
	"time"
)

type sessionKey = stats.ConnTagInfo

// session defines the client connection sessions,
// managed by server during the lifecycle of a client connection.
// Generally, a session needs to finish two phases to be complete:
//  1. creation phase when a new underlying connection detected
//  2. binding & updating phase when a service layer procedure is done
type session struct {
	// the primary key of this session
	key *sessionKey

	// the channel used by this session
	stream *S1Service_StreamTransferServer

	// the client id associated with this session
	clientID string

	// timestamp, in Unix seconds, of session updating
	updateTime int64
}

// bind binds a client id to this session
func (s *session) bind(cli string) {
	s.clientID = cli
	s.updateTime = time.Now().Unix()
}

// sessionManager manages client sessions.
type sessionManager struct {
	server     *Server                  //ref to owner
	mutex      sync.RWMutex             //mutex control for sessions
	sessions   map[*sessionKey]*session //client connection key -> client connection info
	indices    map[string]*session      //client id -> client connection info
	accessTime int64
}

func newSessionManager(owner *Server) *sessionManager {
	sm := &sessionManager{
		server:   owner,
		sessions: make(map[*sessionKey]*session),
		indices:  make(map[string]*session),
	}

	return sm
}

// bind binds a client id to an existing client session (phase II).
func (sm *sessionManager) bind(client string, l *session) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	l.bind(client)
	sm.indices[client] = l

	log.Infof("complete session %p with id: %s", l.key, client)
}

// save creates a new session, using the provided
// key as the key of the new session (phase I).
func (sm *sessionManager) save(k *sessionKey) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.sessions[k] = &session{
		key:        k,
		stream:     nil,
		clientID:   "",
		updateTime: time.Now().Unix(),
	}

	log.Infoln("create session with key:", *k)
}

// getSessionByKey returns the session identified by the key.
func (sm *sessionManager) getSessionByKey(k *sessionKey) *session {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	return sm.sessions[k]
}

// getSessionByKey returns the session associated
// with the given client id.
func (sm *sessionManager) getSessionByID(id string) *session {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	return sm.indices[id]
}

// delete deletes the session identified by the key.
func (sm *sessionManager) delete(k *sessionKey) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	l := sm.sessions[k]

	//cli := GetPlatoonManager().GetSessionByClientId(l.clientID)
	//if cli != nil {
	//	cli.Pause()
	//}

	delete(sm.sessions, k)
	delete(sm.indices, l.clientID)
}

func (sm *sessionManager) getSessionKey(ctx context.Context) *sessionKey {
	key, ok := ctx.Value(sessionKeyID).(*sessionKey)
	if !ok {
		return nil
	}

	return key
}

// updateStream updates the stream server part of a session with a new stream.
// The session is identified by
// by the context stream of
func (sm *sessionManager) updateStream(stream *S1Service_StreamTransferServer) error {
	srv := *stream
	key := sm.getSessionKey(srv.Context())
	if key == nil {
		return errors.New("session key is nil")
	}

	session := sm.getSessionByKey(key)
	if session == nil {
		return fmt.Errorf("session not found for key %p", key)

	}

	session.stream = stream

	log.Infoln("update stream server to", session.stream)

	return nil
}

func (sm *sessionManager) size() int {
	return len(sm.sessions)
}
