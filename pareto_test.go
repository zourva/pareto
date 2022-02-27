package pareto

import (
	"testing"
)

func TestDefaultSetup(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "working dir"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Setup()
			Teardown()
		})
	}
}

