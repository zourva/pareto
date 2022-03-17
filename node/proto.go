package node

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"time"
)

// agent protocol stack entity
type AgentProto struct {
	upper  *Agent
	lower  S1ServiceClient
	stream S1Service_StreamTransferClient
}

func NewAgentProto(upper *Agent, lower S1ServiceClient) *AgentProto {
	return &AgentProto{
		upper: upper,
		lower: lower,
	}
}

func (a *AgentProto) getContextPiggybacked() context.Context {
	md := metadata.New(map[string]string{clientKeyId: a.upper.clientId})
	return metadata.NewOutgoingContext(context.Background(), md)
}

// sign up & update provision info
func (a *AgentProto) doSignUp() bool {
	id := box.CpuId()
	timestamp := box.TimeNowUs()
	key := getAgentAuthKey(AES, id, getCipherKey(AES), timestamp)
	if key == nil {
		log.Errorf("make SignUp authentication token failed")
		return false
	}

	// sign up
	rsp, err := a.lower.SignUp(context.Background(), &SignUpReq{
		Alg:       AES,
		Hid:       id,
		Key:       box.Base64(key),
		Timestamp: timestamp,
	})

	if err != nil {
		log.Errorln("agent SignUp failed:", err)
		return false
	}

	log.Debugln("receive SignUp reply:", *rsp)

	// todo check server identity

	// update provision info
	if !a.upper.configMgr.UpdateProvision(rsp.Id, rsp.Expire) {
		log.Errorln("update agent provision info failed")
		return false
	}

	// no update manually, reload from config manager
	//a.upper.clientId = rsp.Id

	return true
}

func (a *AgentProto) doSignIn() bool {
	rsp, err := a.lower.SignIn(a.getContextPiggybacked(), &SignInReq{})
	if err != nil {
		log.Errorln("agent SignIn failed:", err)
		return false
	}

	log.Debugln("receive SignIn reply:", *rsp)

	return true
}

func (a *AgentProto) doConfigOta() bool {
	return false
}

func (a *AgentProto) doVersionOta() bool {
	return false
}

func (a *AgentProto) doSignOut() bool {
	rsp, err := a.lower.SignOut(a.getContextPiggybacked(), &SignOutReq{})
	if err != nil {
		log.Errorln("agent SignOut failed:", err)
		return false
	}

	log.Debugln("receive SignOut reply:", *rsp)

	return true
}

// enableWatcher enables agent protocol stack to accept server-initiated commands.
func (a *AgentProto) enableWatcher() bool {
	stream, err := a.lower.StreamTransfer(a.getContextPiggybacked())
	if err != nil {
		log.Errorln("create stream transfer channel failed:", err)
		return false
	}

	a.stream = stream

	req := &StreamMessage{
		Proc: Procedure_Initiate,
		Code: ErrorCode_Success,
	}

	if err := a.stream.Send(req); err != nil {
		log.Errorln("stream initiate failed:", err)
		return false
	}

	// recv and callback
	go func() {
		for {
			msg := a.recvStream()
			if msg == nil {
				return
			}

			a.upper.onStreamMsg(msg)
		}
	}()

	log.Infoln("agent watcher registered")

	return true
}

func (a *AgentProto) recvStream() *StreamMessage {
	msg, err := a.stream.Recv()
	if err == io.EOF {
		log.Infoln("agent stream channel quit")
		return nil
	}

	if s, ok := status.FromError(err); ok {
		if s.Code() == codes.Canceled {
			log.Infoln("recv from stream canceled")
			return nil
		}
	}

	if err != nil {
		log.Warnln("recv from stream channel failed:", err)
		log.Infoln("try recreating stream channel by restarting")
		a.upper.MoveToState(restarting)
		return nil
	}

	log.Debugln("recv msg from stream channel:", *msg)

	return msg
}

func (a *AgentProto) sendStream(msg *StreamMessage) error {
	if err := a.stream.Send(msg); err != nil {
		_ = a.stream.CloseSend()
		log.Errorln("send to stream channel failed:", err)
		return err
	}

	return nil
}

type streamProcHandler func(S1Service_StreamTransferServer, *StreamMessage) (*StreamMessage, error)

type ServerProto struct {
	upper    *Server
	handlers map[Procedure]streamProcHandler
}

func NewServerProto(upper *Server) *ServerProto {
	s := &ServerProto{
		upper:    upper,
		handlers: make(map[Procedure]streamProcHandler),
	}

	s.handlers[Procedure_Initiate] = s.onReady

	return s
}

func (s *ServerProto) getClientId(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		return md.Get(clientKeyId)[0]
	}

	return emptyString
}

func (s *ServerProto) validate(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Errorf("metadata is not provided")
		return emptyString, status.Errorf(codes.Code(ErrorCode_IntegrityError), "metadata is not provided")
	}

	id := md.Get(clientKeyId)[0]
	if len(id) == 0 {
		log.Errorf("metadata is invalid")
		return emptyString, status.Errorf(codes.Code(ErrorCode_IntegrityError), "metadata is invalid")
	}

	return id, nil
}

// SignUp assigns new id to the client if not exists yet.
func (s *ServerProto) SignUp(ctx context.Context, req *SignUpReq) (*SignUpRsp, error) {
	log.Traceln("SignUp request received:", *req)

	// todo check client identity
	//

	//two cases exists:
	//one for duplicate sign up in current link
	//another for link re-establish during window
	key := s.upper.ssnMgr.getSessionKey(ctx)
	session := s.upper.ssnMgr.getSessionByKey(key)
	if session == nil {
		log.Errorln("illegal state: session not found for", key)
		return nil, status.Error(codes.Code(ErrorCode_IllegalState), "illegal state: connection lost")
	}

	// for new client
	id := box.UUID()
	expire := uint64(time.Now().Unix() + 7*24*3600)

	// duplicate SignUp
	if len(session.clientId) > 0 {
		id = session.clientId
		node := s.upper.confMgr.GetNode(id)
		if node != nil {
			expire = node.ExpireTime
		}

		log.Infof("client %s duplicate SignUp", session.clientId)
	}

	// session creation phase II
	s.upper.ssnMgr.bind(id, session)

	serverKey := getServerAuthKey(AES, id, req.Hid, getCipherKey(AES), req.Timestamp, expire)

	_ = s.upper.confMgr.SaveNode(&Node{
		Identity:   id,
		Endpoint:   key.RemoteAddr.String(),
		ExpireTime: expire,
		SignUpTime: box.TimeNowUs(),
		UpdateTime: box.TimeNowUs(),
	})

	return &SignUpRsp{
		Alg:    AES,
		Id:     id,
		Key:    box.Base64(serverKey),
		Expire: expire,
	}, nil
}

func (s *ServerProto) SignIn(ctx context.Context, req *SignInReq) (*SignInRsp, error) {
	log.Traceln("SignIn request received:", *req)

	//two cases exists:
	//one for duplicate sign in in current link
	//another for link re-establish during window
	key := s.upper.ssnMgr.getSessionKey(ctx)
	session := s.upper.ssnMgr.getSessionByKey(key)
	if session == nil {
		log.Errorln("illegal state: session not found for", key)
		return nil, status.Error(codes.Code(ErrorCode_IllegalState), "illegal state: connection lost")
	}

	// duplicate SignIn or there's proceeding SignUp
	if len(session.clientId) > 0 {
		log.Infof("client %s duplicate SignIn, ignore", session.clientId)
		return &SignInRsp{}, nil
	}

	// load, update and write back node info
	if id, err := s.validate(ctx); err != nil {
		log.Errorf("metadata is not provided")
		return nil, err
	} else {
		session.clientId = id
	}

	node := s.upper.confMgr.GetNode(session.clientId)
	if node == nil {
		log.Errorf("node info is not found")
		return nil, status.Errorf(codes.Code(ErrorCode_NotFound), "node is not found")
	}

	node.Status = 1
	node.Endpoint = key.RemoteAddr.String()
	node.SignInTime = box.TimeNowUs()
	node.UpdateTime = node.SignInTime

	_ = s.upper.confMgr.SaveNode(node)

	if s.upper.options.hooks.OnNodeJoin != nil {
		s.upper.options.hooks.OnNodeJoin(node)
	}

	return &SignInRsp{}, nil
}

// SignOut if session exists, otherwise do nothing and assume success
func (s *ServerProto) SignOut(ctx context.Context, req *SignOutReq) (*SignOutRsp, error) {
	log.Traceln("SignOut request received", *req)

	session := s.getSession(ctx)
	if session != nil {
		node := s.upper.confMgr.GetNode(session.clientId)
		if node != nil {
			node.Status = 0
			node.UpdateTime = node.SignInTime
			_ = s.upper.confMgr.SaveNode(node)
		}
	}

	return &SignOutRsp{}, nil
}

func (s *ServerProto) Report(ctx context.Context, req *ReportStatusReq) (*empty.Empty, error) {
	panic("implement me")
}

func (s *ServerProto) Config(ctx context.Context, req *GetConfigReq) (*GetConfigRsp, error) {
	panic("implement me")
}

func (s *ServerProto) StreamTransfer(server S1Service_StreamTransferServer) error {
	for {
		msg, err := server.Recv()
		if err == io.EOF {
			log.Infoln("stream server finished successfully")
			return nil
		}

		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.Canceled {
				log.Infoln("stream server canceled")
				return nil
			}
		}

		if err != nil {
			log.Errorln("stream server recv error:", err)
			return err
		}

		log.Traceln("recv stream msg:", msg)

		cb, exist := s.handlers[msg.Proc]
		if !exist {
			log.Errorln("unknown stream procedure", msg.Proc)
			return status.Errorf(codes.Code(ErrorCode_NotFound), "unknown stream procedure")
		}

		reply, err := cb(server, msg)
		if err != nil {
			log.Errorln("failed to proc stream procedure", msg.Proc)
			return status.Errorf(codes.Code(ErrorCode_InternalError), err.Error())
		}

		// send reply if any
		if reply != nil {
			if err = server.Send(reply); err != nil {
				log.Errorln("send stream reply failed:", msg, err)
				return err
			}

			log.Traceln("send stream reply done:", reply)
		}
	}
}

func (s *ServerProto) getSession(ctx context.Context) *session {
	key := s.upper.ssnMgr.getSessionKey(ctx)
	session := s.upper.ssnMgr.getSessionByKey(key)
	return session
}

func (s *ServerProto) onReady(server S1Service_StreamTransferServer, msg *StreamMessage) (*StreamMessage, error) {
	if msg.Proc == Procedure_Initiate {
		log.Infoln("server installed watcher for client", s.getClientId(server.Context()))

		if err := s.upper.ssnMgr.updateStream(&server); err != nil {
			log.Errorln("updateStream failed:", err)
			return nil, status.Errorf(codes.Code(ErrorCode_InternalError), "update stream failed")
		}

		return nil, nil
	}

	return nil, status.Errorf(codes.Code(ErrorCode_UnSupported),
		"unsupported procedure %s", msg.Proc.String())
}
