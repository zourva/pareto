package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewServer(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	server := NewServer()
	assert.NotNil(t, server)

	err := server.Startup()
	assert.Nil(t, err)

	server.Shutdown()
}
