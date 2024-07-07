package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/zourva/pareto/endec/jsonrpc2"
	"testing"
	"time"
)

const (
	subject = "/test"
)

type broker struct {
	*MetaService
}

func (b *broker) Bind(channels map[string]jsonrpc2.Dispatcher) error {
	for name, handler := range channels {
		if err := b.ExposeMethod(name, handler); err != nil {
			return err
		}
	}

	return nil
}

func (b *broker) Call(channel string, data []byte, to time.Duration) ([]byte, error) {
	return b.Messager().CallV2(channel, data, to)
}

func newBroker() *broker {
	svc := NewMetaService(&Descriptor{
		Name:     "xxx",
		Registry: "nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222",
	})

	return &broker{MetaService: svc}
}

type TestReq struct {
	Name string `json:"name"`
}

type TestRsp struct {
	Value string `json:"value"`
}

func TestCallerCallee(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	mm := EnableMonitor("nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222")
	assert.NotNil(t, mm)

	//
	bearer := newBroker()
	Start(bearer)

	rr := jsonrpc2.NewRouter(bearer)

	// /test/method1 ok
	ch := rr.AddChannel(subject, map[string]jsonrpc2.Handler{
		"method1": func(request *jsonrpc2.RPCRequest) *jsonrpc2.RPCResponse {
			return jsonrpc2.NewResponse(request, &TestRsp{Value: "method1 on test no group"})
		},
	})

	// /test/method2-5[group1] passed first then terminated
	ch.Group("group1", map[string]jsonrpc2.Handler{
		"method2": func(request *jsonrpc2.RPCRequest) *jsonrpc2.RPCResponse {
			return jsonrpc2.NewResponse(request, &TestRsp{Value: "method2 on test group1"})
		},
	}).AddInterceptors(func(request *jsonrpc2.RPCRequest) *jsonrpc2.RPCResponse {
		log.Infoln("passed interceptor")
		return nil
	})

	ch.Group("group1", map[string]jsonrpc2.Handler{
		"method3": func(request *jsonrpc2.RPCRequest) *jsonrpc2.RPCResponse {
			return jsonrpc2.NewResponse(request, &TestRsp{Value: "method3 on test group1"})
		},
		"method4": func(request *jsonrpc2.RPCRequest) *jsonrpc2.RPCResponse {
			return jsonrpc2.NewResponse(request, &TestRsp{Value: "method4 on test group1"})
		},
		"method5": func(request *jsonrpc2.RPCRequest) *jsonrpc2.RPCResponse {
			return jsonrpc2.NewResponse(request, &TestRsp{Value: "method5 on test group1"})
		},
	}).AddInterceptors(func(request *jsonrpc2.RPCRequest) *jsonrpc2.RPCResponse {
		log.Infoln("terminated interceptor at group1")
		return jsonrpc2.NewErrorResponse(jsonrpc2.ErrApplicationError, "terminated by interceptor")
	})

	// /test1/method1-2 done
	rr.AddChannel("/test1", map[string]jsonrpc2.Handler{
		"method1": func(request *jsonrpc2.RPCRequest) *jsonrpc2.RPCResponse {
			return jsonrpc2.NewResponse(request, &TestRsp{Value: "method1 on /test1"})
		},
		"method2": func(request *jsonrpc2.RPCRequest) *jsonrpc2.RPCResponse {
			return jsonrpc2.NewResponse(request, &TestRsp{Value: "method2 on /test1"})
		},
	}).AddInterceptors(func(request *jsonrpc2.RPCRequest) *jsonrpc2.RPCResponse {
		return nil
	}).AddPostHandlers(func(request *jsonrpc2.RPCRequest, response *jsonrpc2.RPCResponse) {
		log.Infoln("it's okay, done")
	})

	//rr.EnableTrace(true, 1024)

	server := jsonrpc2.NewServer(rr)
	err := server.Serve()
	if err != nil {
		t.Fatal(err)
	}

	client := jsonrpc2.NewClient(bearer)
	rsp, err := client.Invoke(subject, "method1", 30*time.Second, &TestReq{Name: "hello"})
	t.Log(rsp, err)

	////t.Log(rsp)
	//var tr TestRsp
	//rsp.GetObject(&tr)
	//t.Log(rsp)
	//assert.Equal(t, tr.Value, "world")

	rsp, err = client.Invoke(subject, "method2", time.Second, &TestReq{Name: "hello"})
	t.Log(rsp, err)

	rsp, err = client.Invoke(subject, "method3", time.Second, &TestReq{Name: "hello"})
	t.Log(rsp, err)

	rsp, err = client.Invoke(subject, "method4", time.Second, &TestReq{Name: "hello"})
	t.Log(rsp, err)

	rsp, err = client.Invoke(subject, "method5", time.Second, &TestReq{Name: "hello"})
	t.Log(rsp, err)

	rsp, err = client.Invoke("/test1", "method1", time.Second, &TestReq{Name: "hello"})
	t.Log(rsp, err)

	rsp, err = client.Invoke("/test1", "method2", time.Second, &TestReq{Name: "hello"})
	t.Log(rsp, err)

	Stop(bearer)

	DisableMonitor()
}
