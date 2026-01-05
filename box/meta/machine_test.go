package meta

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newTestMachine(precision time.Duration) *StateMachine[string] {
	return NewStateMachine[string]("test machine", precision)
}

func noop(args any) {}

func TestStateMachineStartupRequiresRegisteredStates(t *testing.T) {
	sm := newTestMachine(time.Millisecond * 5)
	assert.False(t, sm.Startup())
}

func TestStateMachineStartupFailsWithoutStartingState(t *testing.T) {
	sm := newTestMachine(time.Millisecond * 5)
	assert.True(t, sm.RegisterState(&State[string]{Name: "running", Action: noop}))
	assert.True(t, sm.RegisterState(&State[string]{Name: "stopping", Action: noop}))
	assert.True(t, sm.SetStoppingState("stopping"))
	assert.False(t, sm.Startup())
}

func TestStateMachineMoveToStateResetsTickCounter(t *testing.T) {
	sm := newTestMachine(time.Millisecond * 5)
	delayed := &State[string]{Name: "delayed", Ticks: 3, Action: noop}
	other := &State[string]{Name: "other", Action: noop}
	assert.True(t, sm.RegisterState(delayed))
	assert.True(t, sm.RegisterState(other))
	assert.True(t, sm.MoveToState("delayed"))
	delayed.trigger()
	assert.Equal(t, uint(1), delayed.tickCnt)
	assert.True(t, sm.MoveToState("other"))
	assert.True(t, sm.MoveToState("delayed"))
	assert.Equal(t, uint(0), delayed.tickCnt)
}

func TestStateMachineShutdownIsIdempotentAndRestartable(t *testing.T) {
	sm := newTestMachine(time.Millisecond * 5)
	started := make(chan struct{}, 1)
	running := &State[string]{
		Name: "running",
		Action: func(args interface{}) {
			select {
			case started <- struct{}{}:
			default:
			}
		},
	}
	stopping := &State[string]{Name: "stopping", Action: noop}
	assert.True(t, sm.RegisterStates([]*State[string]{running, stopping}))
	assert.True(t, sm.Startup())
	select {
	case <-started:
	case <-time.After(time.Millisecond * 50):
		t.Fatal("running action never triggered")
	}
	assert.NotPanics(t, func() { sm.Shutdown() })
	assert.NotPanics(t, func() { sm.Shutdown() })
	assert.True(t, sm.Startup())
	time.Sleep(time.Millisecond * 20)
	sm.Shutdown()
}

func TestStateMachineExecutesActionsAndStops(t *testing.T) {
	sm := newTestMachine(time.Millisecond * 5)
	startHits := make(chan struct{}, 1)
	stopHits := make(chan struct{}, 1)
	starting := &State[string]{
		Name: "starting",
		Action: func(args interface{}) {
			select {
			case startHits <- struct{}{}:
			default:
			}
		},
	}
	stopping := &State[string]{
		Name: "stopping",
		Action: func(args interface{}) {
			select {
			case stopHits <- struct{}{}:
			default:
			}
		},
	}
	assert.True(t, sm.RegisterStates([]*State[string]{starting, stopping}))
	assert.True(t, sm.Startup())
	select {
	case <-startHits:
	case <-time.After(time.Millisecond * 50):
		t.Fatal("starting action never triggered")
	}
	sm.Shutdown()
	select {
	case <-stopHits:
	case <-time.After(time.Millisecond * 50):
		t.Fatal("stopping action never triggered")
	}
}
