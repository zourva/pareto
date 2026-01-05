package meta

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

// Action defines the action to execute when a state is reached.
type Action func(args any)

// State is the state meta info.
type State[T string | int32] struct {
	Name   T      //name of this state
	Ticks  uint   //optional ticks to wait before trigger action, 0 means no wait
	Action Action //action of this state when ticks expire
	Args   interface{}

	tickCnt uint             //ticks already passed after started
	machine *StateMachine[T] //reference to owner
}

func (s *State[T]) trigger() {
	s.tickCnt++
	if 0 == s.Ticks || s.tickCnt%s.Ticks == 0 {
		if s.Action != nil {
			if s.tickCnt%5 == 0 {
				if s.machine.trace {
					log.Debugf("state machine [%s] trigger action %v", s.machine.name, s.Name)
				}
				//else {
				//	log.Tracef("state machine [%s] trigger action %v", s.machine.name, s.Name)
				//}
			}

			s.Action(s.Args)

			if /*!box.IsZero(s.machine.stopping) && */ s.Name == s.machine.stopping {
				s.machine.signalStopped()
			}
		}
	}
}

// StateMachine sums all states and related options to make a DFA.
type StateMachine[T string | int32] struct {
	name   string
	states map[T]*State[T] // not goroutine-safe, use in read-only mode after initialization

	starting T // name of starting state
	stopping T // name of stopping state
	saved    T // save state for later restore

	// active state
	current atomic.Pointer[T]

	ticker    *time.Ticker
	precision time.Duration
	trace     bool

	ctx    context.Context
	cancel context.CancelFunc
	loopWG sync.WaitGroup

	stopAck struct {
		wg   *sync.WaitGroup
		once sync.Once
	}
}

// NewStateMachine creates a new state machine
// with the given name and ticker duration.
func NewStateMachine[T string | int32](name string, precision time.Duration) *StateMachine[T] {
	sm := &StateMachine[T]{
		name:      name,
		states:    make(map[T]*State[T]),
		precision: precision,
		trace:     false,
	}

	sm.current.Store(new(T))
	sm.initRuntime()

	return sm
}

func (sm *StateMachine[T]) initRuntime() {
	sm.ctx, sm.cancel = context.WithCancel(context.Background())
	sm.stopAck.once = sync.Once{}
	sm.stopAck.wg = nil
}

func (sm *StateMachine[T]) signalStopped() {
	if sm.stopAck.wg == nil {
		return
	}
	sm.stopAck.once.Do(func() {
		sm.stopAck.wg.Done()
		log.Debugf("state machine [%s] stop acknowledged", sm.name)
	})
}

// GetState returns the current state.
func (sm *StateMachine[T]) GetState() T {
	return *sm.current.Load()
}

// EnableStateTrace enables or disables the tracing of internal flow.
// It's disabled by default.
func (sm *StateMachine[T]) EnableStateTrace(on bool) {
	sm.trace = on
}

// MoveToState moves the current state to the vien one
func (sm *StateMachine[T]) MoveToState(s T) bool {
	if sm.GetState() == s {
		log.Tracef("state machine [%s] is already in state %v", sm.name, sm.GetState())
		return true
	}

	state, exist := sm.states[s]
	if !exist {
		log.Errorf("state machine [%s] state %v not found", sm.name, s)
		return false
	}

	if sm.trace {
		log.Debugf("state machine [%s] move state from %v to %v", sm.name, sm.GetState(), s)
	}
	//else {
	//	log.Tracef("state machine [%s] move state from %v to %v", sm.name, sm.current, s)
	//}

	//sm.mutex.Lock()
	//defer sm.mutex.Unlock()
	//sm.current = s
	sm.current.Store(&state.Name)
	state.tickCnt = 0

	return true
}

// RegisterState registers a new state to the machine.
// The old is replaced if a state with the same name exists.
//
//	NOTE: This method is not goroutine-safe, call it when initialization only.
func (sm *StateMachine[T]) RegisterState(s *State[T]) bool {
	if s == nil {
		log.Errorf("state machine [%s] reg invalid state, ignored", sm.name)
		return false
	}

	s.tickCnt = 0
	s.machine = sm
	sm.states[s.Name] = s

	log.Debugf("state machine [%s] register state %v", sm.name, s.Name)

	return true
}

// RegisterStates registers a slice of states, which contains at least
// two elements, to the state machine.
//
// The starting state and stopping state are set to the first
// and last element of the slice separately.
//
// The State.Id field is set according to the slice suffix starting from 0.
//
//	NOTE: This method is not goroutine-safe, call it when initialization only.
func (sm *StateMachine[T]) RegisterStates(ss []*State[T]) bool {
	if len(ss) <= 1 {
		log.Errorf("state machine [%s] at least two states are needed", sm.name)
		return false
	}

	for _, s := range ss {
		if !sm.RegisterState(s) {
			return false
		}
	}

	sm.SetStartingState(ss[0].Name)
	sm.SetStoppingState(ss[len(ss)-1].Name)

	return true
}

// SetStartingState sets a state, identified by the given name,
// as the starting state. False is returned if no state found
// for the given name, or true is returned.
//
//	NOTE: This method is not goroutine-safe, call it when initialization only.
func (sm *StateMachine[T]) SetStartingState(state T) bool {
	_, ok := sm.states[state]
	if !ok {
		log.Errorf("given starting state %v is not registered", state)
		return false
	}

	sm.starting = state
	return true
}

// SetStoppingState sets a state, identified by the given name,
// as the stopping state. False is returned if no state found
// for the given name, or true is returned.
//
//	NOTE: This method is not goroutine-safe, call it when initialization only.
func (sm *StateMachine[T]) SetStoppingState(state T) bool {
	_, ok := sm.states[state]
	if !ok {
		log.Errorf("given stopping state %v is not registered", state)
		return false
	}

	sm.stopping = state
	return true
}

// Startup starts running of the machine.
func (sm *StateMachine[T]) Startup() bool {
	if len(sm.states) == 0 {
		log.Errorf("state machine [%s] has no states registered, cannot startup", sm.name)
		return false
	}

	if sm.ticker != nil {
		log.Warnf("state machine [%s] already running", sm.name)
		return false
	}

	//if box.IsZero(sm.starting) {
	//	log.Errorf("state machine [%s] has no starting state", sm.name)
	//	return false
	//}
	//
	//if box.IsZero(sm.GetState()) {
	//	sm.MoveToState(sm.starting)
	//}

	if _, ok := sm.states[sm.starting]; !ok {
		log.Errorf("state machine [%s] has no valid starting state", sm.name)
		return false
	}

	sm.initRuntime()
	if _, ok := sm.states[sm.stopping]; ok {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		sm.stopAck.wg = wg
	}
	if !sm.MoveToState(sm.starting) {
		sm.cancel()
		sm.ctx = nil
		return false
	}

	sm.ticker = time.NewTicker(sm.precision)

	sm.loopWG.Add(1)
	go sm.loop()
	log.Infof("state machine [%s] started", sm.name)

	return true
}

// Shutdown stops the internal loop and wait
// until the stopping state action returns.
func (sm *StateMachine[T]) Shutdown() {
	if sm.ticker == nil {
		log.Infof("state machine [%s] already stopped", sm.name)
		return
	}

	log.Infof("state machine [%s] is exiting", sm.name)

	if _, ok := sm.states[sm.stopping]; ok {
		sm.MoveToState(sm.stopping)
		if sm.stopAck.wg != nil {
			sm.stopAck.wg.Wait()
		}
	}

	if sm.cancel != nil {
		sm.cancel()
	}
	sm.loopWG.Wait()

	sm.ticker.Stop()
	sm.ticker = nil
	sm.cancel = nil
	sm.ctx = nil
	sm.stopAck.wg = nil

	log.Infof("state machine [%s] exited", sm.name)
}

// Pause pauses the internal timer tick loop.
//
//	NOTE: Not goroutine-safe.
func (sm *StateMachine[T]) Pause() {
	if sm.ticker == nil {
		return
	}
	sm.ticker.Stop()
}

// Resume resumes the internal timer tick loop.
//
//	NOTE: Not goroutine-safe.
func (sm *StateMachine[T]) Resume() {
	if sm.ticker == nil {
		return
	}
	sm.ticker.Reset(sm.precision)
}

// SaveState saves the current state for restore.
//
//	NOTE: Not goroutine-safe.
func (sm *StateMachine[T]) SaveState() {
	sm.saved = sm.GetState()
}

// RestoreState moves to the latest saved state.
//
//	NOTE: Not goroutine-safe.
func (sm *StateMachine[T]) RestoreState() {
	sm.MoveToState(sm.saved)
}

// triggers execution of the action defined in current state.
func (sm *StateMachine[T]) trigger() {
	//sm.mutex.RLock()
	//defer sm.mutex.RUnlock()

	stateName := sm.GetState()
	state, ok := sm.states[stateName]
	if !ok {
		log.Errorf("state machine [%s] state %v not registered", sm.name, stateName)
		return
	}
	state.trigger()
}

func (sm *StateMachine[T]) loop() {
	defer sm.loopWG.Done()

	// Trigger immediately once.
	sm.trigger()

	for {
		select {
		case <-sm.ctx.Done():
			log.Infof("state machine [%s] loop quit", sm.name)
			return
		case <-sm.ticker.C:
			sm.trigger()
		}
	}
}
