package ipc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMessager(t *testing.T) {
	// precondition: nats-server process is not started yet.
	broker := "nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222"
	messager, err := NewMessager(&MessagerConf{
		BusConf: &BusConf{Name: "test-bus", Type: InterProcBus, Broker: broker},
		RpcConf: &RPCConf{Name: "test-rpc", Type: InterProcRpc, Broker: broker},
	})

	assert.NotNil(t, err)
	assert.Nil(t, messager)
	t.Log(err)
}

func TestNewMessager_NatsServer(t *testing.T) {
	// precondition: nats-server process is running
	broker := "nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222"
	messager, err := NewMessager(&MessagerConf{
		BusConf: &BusConf{Name: "test-bus", Type: InterProcBus, Broker: broker},
		RpcConf: &RPCConf{Name: "test-rpc", Type: InterProcRpc, Broker: broker},
	})

	assert.Nil(t, err)
	assert.NotNil(t, messager)
}
