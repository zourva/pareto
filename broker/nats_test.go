package broker

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zourva/pareto/service"
	"testing"
	"time"
)

func TestNewEmbeddedNats(t *testing.T) {
	server, err := NewEmbeddedNats(
		WithPort(4223),
		WithMonitorPort(8223),
		WithLoggerFile("stdout"),
		WithAuthorizationToken("dag0HTXl4RGg7dXdaJwbC8"))
	require.Nil(t, err)

	err = server.Startup()
	require.Nil(t, err)

	svc := service.New(&service.Descriptor{Name: "test", Registry: "nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4223"})

	_ = svc.Messager().ExposeV2("test.rr", func(data []byte) ([]byte, error) {
		return []byte("world"), nil
	})

	_ = svc.Messager().Subscribe("test.ps", func(bytes []byte) {
		t.Log("ok:", string(bytes))
	})

	rsp, _ := svc.Messager().CallV2("test.rr", []byte("hello"), 5*time.Second)
	assert.EqualValues(t, rsp, "world")

	_ = svc.Messager().Publish("test.ps", []byte("follow me"))

	err = server.Shutdown()
	assert.Nil(t, err)
}
