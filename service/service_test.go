package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGenericMetaService(t *testing.T) {
	assert.Nil(t, NewGenericMetaService("", ""))
	assert.Nil(t, NewGenericMetaService("test", ""))
	s := NewGenericMetaService("test", "nats://dag0HTXl4RGg7dXdaJwbC8@localhost:4222")
	assert.NotNil(t, s)
	assert.Equal(t, s.Name(), "test")
}
