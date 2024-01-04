package jsonrpc2

import (
	"github.com/stretchr/testify/assert"
	"github.com/zourva/pareto/service"
	"testing"
	"time"
)

const (
	subject = "/test"
)

type broker struct {
	*service.MetaService
}

func (b *broker) Bind(fn func([]byte) ([]byte, error)) error {
	return b.ExposeMethod(subject, fn)
}

func (b *broker) Call(channel string, data []byte, to time.Duration) ([]byte, error) {
	return b.Messager().CallV2(channel, data, to)
}

func newBroker() *broker {
	svc := service.NewMetaService(&service.Descriptor{
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
	bearer := newBroker()
	service.Start(bearer)

	server := NewServer(bearer)
	server.RegisterHandler("method1", func(request *RPCRequest) *RPCResponse {
		return NewResponse(request, &TestRsp{Value: "world"})
	})

	err := server.Serve()
	if err != nil {
		t.Fatal(err)
	}

	client := NewClient(bearer)
	rsp, err := client.Invoke(subject, "method1", time.Second, &TestReq{Name: "hello"})

	//t.Log(rsp)
	var tr TestRsp
	rsp.GetObject(&tr)
	assert.Equal(t, tr.Value, "world")

	service.Stop(bearer)
}
