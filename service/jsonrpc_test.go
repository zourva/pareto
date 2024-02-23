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

func (b *broker) Bind(channels map[string]jsonrpc2.ChannelHandler) error {
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
	rr.AddChannel("test1")
	rr.AddChannel("test2")
	rr.AddChannel("test3")
	assert.Equal(t, len(rr.AllChannels()), 3)

	server := jsonrpc2.NewServer(rr)
	server.RegisterHandler(subject, "method1", func(request *jsonrpc2.RPCRequest) *jsonrpc2.RPCResponse {
		return jsonrpc2.NewResponse(request, &TestRsp{Value: "world"})
	})

	assert.Equal(t, len(rr.AllChannels()), 4)

	err := server.Serve()
	if err != nil {
		t.Fatal(err)
	}

	client := jsonrpc2.NewClient(bearer)
	rsp, err := client.Invoke(subject, "method1", time.Second, &TestReq{Name: "hello"})

	//t.Log(rsp)
	var tr TestRsp
	rsp.GetObject(&tr)
	assert.Equal(t, tr.Value, "world")

	Stop(bearer)

	DisableMonitor()
}
