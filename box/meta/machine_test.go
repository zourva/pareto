package meta

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func onStarting(args interface{}) {
	//
}

func onStopping(args interface{}) {
	//
}

func TestNewStateMachine(t *testing.T) {
	sm := NewStateMachine("test machine", time.Millisecond*100)
	assert.NotNil(t, sm)

	sm.RegisterStates([]*State{
		{Name: "starting", Action: onStarting},
		{Name: "stopping", Action: onStopping},
	})

	sm.SetStartingState("starting")
	sm.SetStoppingState("stopping")

	sm.Startup()

	time.Sleep(time.Second * 10)

	sm.Shutdown()
}
