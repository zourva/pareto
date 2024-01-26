package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewServer(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	server := NewRegistryManager("nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222")
	assert.NotNil(t, server)

	assert.True(t, server.Startup())

	server.Shutdown()
}

func TestMonitor(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	mm := EnableMonitor("nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222")
	assert.NotNil(t, mm)

	DisableMonitor()
}

func TestMonitorConcurrent(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	mm := EnableMonitor("nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222")
	assert.NotNil(t, mm)

	mm2 := EnableMonitor("nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222")
	assert.NotNil(t, mm2)

	assert.Equal(t, mm, mm2)

	DisableMonitor()
	DisableMonitor()
	DisableMonitor()
}

func TestMonitorNil(t *testing.T) {
	DisableMonitor()
}
