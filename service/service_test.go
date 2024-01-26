package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestNewMetaService(t *testing.T) {
	assert.Nil(t, NewMetaService(&Descriptor{"", ""}))
	assert.Nil(t, NewMetaService(&Descriptor{"test", ""}))
	s := NewMetaService(&Descriptor{"test", "nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222"})
	assert.NotNil(t, s)
	assert.Equal(t, s.Name(), "test")
}

func TestWatch(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	mm := EnableMonitor("nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222")
	require.NotNil(t, mm)

	time.Sleep(5 * time.Second)

	w := NewMetaService(&Descriptor{"watcher", "nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222"})
	require.NotNil(t, w)

	w1 := NewMetaService(&Descriptor{"watched1", "nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222"})
	require.NotNil(t, w1)

	w2 := NewMetaService(&Descriptor{"watched2", "nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222"})
	require.NotNil(t, w2)

	err := w.Watch(func(status *Status) {
		t.Logf("service %s state changed: %v", status.Name, status)
	}, "watched1", "watched2")
	require.Nil(t, err)

	Start(w1)
	Start(w2)
	Start(w)

	time.Sleep(5 * time.Second)

	w1.SetReady(false)
	w2.SetReady(false)

	time.Sleep(3 * time.Second)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	// hold
	<-exit

	Stop(w1)
	Stop(w2)
	Stop(w)
	DisableMonitor()
}
