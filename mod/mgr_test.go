package mod

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewServiceManager(t *testing.T) {
	tests := []struct {
		name string
		want ServiceManager
	}{
		{name: "default-service-manager"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.SetLevel(log.DebugLevel)

			got := NewServiceManager()

			assert.NotNil(t, got)

			assert.Nil(t, got.Startup())

			got.Shutdown()
		})
	}
}
