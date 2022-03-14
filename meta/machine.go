package meta

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type Action func(args interface{})

type State struct {
	Name   string //name of this state
	Ticks  uint   //optional ticks to wait before trigger action, 0 means no wait
	Action Action //action of this state when ticks expire
	Args   interface{}

	tickCnt uint          //ticks already passed after started
	machine *StateMachine //reference to owner
}

type StateMachine struct {
	name   string
	states map[string]*State

	starting  string // name of starting state
	stopping  string // name of stopping state
	current   string // current state
	saved     string // save current state for later on restore
	ticker    *time.Ticker
	precision time.Duration
	quit      chan struct{}
	stopped   chan struct{}
	trace     bool
}

func NewStateMachine(name string, precision time.Duration) *StateMachine {
	sm := &StateMachine{
		name:      name,
		states:    make(map[string]*State),
		precision: precision,
		ticker:    time.NewTicker(precision),
		quit:      make(chan struct{}),
		stopped:   make(chan struct{}),
		trace:     false,
	}

	return sm
}

func (s *State) trigger() {
	s.tickCnt++
	if 0 == s.Ticks || s.tickCnt%s.Ticks == 0 {
		if s.Action != nil {
			if s.tickCnt%5 == 0 {
				if s.machine.trace {
					log.Debugf("state machine %s trigger action %s", s.machine.name, s.Name)
				} else {
					log.Tracef("state machine %s trigger action %s", s.machine.name, s.Name)
				}
			}

			s.Action(s.Args)

			if s.machine.stopping != "" && s.Name == s.machine.stopping {
				close(s.machine.stopped)
				//s.machine.stopped = nil
				log.Debugf("state machine %s stop acknowledged", s.machine.name)
			}
		}
	}
}

func (sm *StateMachine) GetState() string {
	return sm.current
}

func (sm *StateMachine) EnableStateTrace(on bool) {
	sm.trace = on
}

func (sm *StateMachine) MoveToState(s string) bool {
	if _, exist := sm.states[s]; !exist {
		log.Errorf("state machine %s state %s not found", sm.name, s)
		return false
	}

	if sm.trace {
		log.Debugf("state machine %s move state from %s to %s", sm.name, sm.current, s)
	} else {
		log.Tracef("state machine %s move state from %s to %s", sm.name, sm.current, s)
	}

	sm.current = s
	return true
}

func (sm *StateMachine) RegisterState(s *State) bool {
	if s == nil || len(s.Name) == 0 {
		log.Errorf("state machine %s reg invalid state, ignored", sm.name)
		return false
	}

	s.tickCnt = 0
	s.machine = sm
	sm.states[s.Name] = s

	log.Debugf("state machine %s save state %s", sm.name, s.Name)

	return true
}

func (sm *StateMachine) RegisterStates(ss []*State) bool {
	for _, s := range ss {
		if !sm.RegisterState(s) {
			return false
		}
	}

	return true
}

func (sm *StateMachine) SetStartingState(state string) bool {
	_, ok := sm.states[state]
	if !ok {
		log.Errorf("given starting state %s is not registered", state)
		return false
	}

	sm.starting = state
	return true
}

func (sm *StateMachine) SetStoppingState(state string) bool {
	_, ok := sm.states[state]
	if !ok {
		log.Errorf("given stopping state %s is not registered", state)
		return false
	}

	sm.stopping = state
	return true
}

func (sm *StateMachine) Startup() bool {
	if len(sm.states) == 0 {
		log.Errorln("state machine %s has no states registered, cannot startup")
		return false
	}

	if sm.starting == "" {
		log.Errorln("state machine %s has no starting state")
		return false
	}

	if sm.current == "" {
		sm.MoveToState(sm.starting)
	}

	go sm.loop()
	log.Infof("state machine %s started", sm.name)

	return true
}

// Shutdown stops the internal loop and wait until the stopping state action returns.
func (sm *StateMachine) Shutdown() {
	log.Infof("state machine %s is exiting", sm.name)

	if len(sm.stopping) != 0 {
		sm.MoveToState(sm.stopping)
		<-sm.stopped
		//log.Infof("state machine %s loop quit", sm.name)
	}

	close(sm.quit)

	sm.ticker.Stop()

	log.Infof("state machine %s exited", sm.name)
}

func (sm *StateMachine) Pause() {
	sm.ticker.Stop()
}

func (sm *StateMachine) Resume() {
	sm.ticker.Reset(sm.precision)
}

func (sm *StateMachine) SaveState() {
	sm.saved = sm.current
}

func (sm *StateMachine) RestoreState() {
	sm.MoveToState(sm.saved)
}

func (sm *StateMachine) trigger() {
	state := sm.states[sm.current]
	state.trigger()
}

func (sm *StateMachine) loop() {
	for
	{
		select {
		case <-sm.quit:
			log.Infof("state machine %s loop quit", sm.name)
			return
		case <-sm.ticker.C:
			sm.trigger()
		}

	}
}
