package meta

import (
	log "github.com/sirupsen/logrus"
	"time"
)

// LoopRunHook defines loop lifecycle callbacks.
// for ref, see http://docs.libuv.org/en/v1.x/loop.html
type LoopRunHook struct {
	//optional, called, once, before the underlying loop starts
	BeforeRun func() error

	//mandatory, called periodically based on the configured ticks
	Working func() error

	//optional, called periodically based on the configured ticks or suppressed if set to nil
	Stalling func() error

	//optional, called, once, before the underlying loop quits
	BeforeQuit func() error
}

// LoopConfig loop configuration, all options have default values except CbWork
type LoopConfig struct {
	//Tick defines clock pulse base precision/resolution of a time-wheel loop in milliseconds.
	//Zero value means the default(1000 ms).
	Tick uint32

	//Work defines the tick count of work check interval.
	//Zero value means the default(1 tick).
	Work uint32

	//Idle defines the tick count of idle check interval.
	//Zero value means the default(1 tick).
	Idle uint32

	//#ticks after which the loop will be terminated,
	// set to 0 as disabling auto termination
	//Quit uint32

	//Sync, if set to false, hook functions provided by Run
	//will be executed in another go routine, i.e. asynchronously.
	//
	//It's false by default.
	Sync bool

	//BailOnError, if set to true, tells the loop manager to break the underlying loop
	//when error is returned by user hooks, otherwise, the loop continues to next iteration.
	//
	//It's false by default.
	BailOnError bool
}

// Loop interface exposed
type Loop interface {
	// Name returns name of the loop.
	Name() string

	// Conf configure the loop with the given configuration.
	// This methods should be called before Run and not be called
	// after the loop is running.
	Conf(conf LoopConfig) bool

	// Run starts the loop with the provided hooks.
	//
	// NOTE: Run is not re-entrant and must not be called within a callback.
	// When the loop is configured to run in async mode,
	// hook functions must be guaranteed to be goroutine-safe.
	Run(hooks LoopRunHook)

	// Alive returns true if loop is running.
	Alive() bool

	// Stopped returns true if loop is stopped.
	Stopped() bool

	// Stop stops the internal timer, close channels and clear all states.
	Stop()
}

type LoopConf = LoopConfig
type LoopHooks = LoopRunHook

// NewLoop creates a new loop object with the given name and conf.
func NewLoop(name string, conf LoopConfig) Loop {
	ticks := time.Millisecond * 1000

	if conf.Tick != 0 {
		ticks = time.Millisecond * time.Duration(conf.Tick)
	}

	return &TimeWheelLoop{
		name:  name,
		state: initialized,
		conf:  conf,
		tick:  time.NewTicker(ticks),
		quit:  make(chan struct{}),
		wait:  make(chan struct{}),
	}
}

const (
	initialized uint32 = iota
	configured
	running
	stopped
)

// TimeWheelLoop provides a simple breakable loop impl.
type TimeWheelLoop struct {
	name  string
	state uint32
	conf  LoopConfig
	tick  *time.Ticker
	quit  chan struct{}
	wait  chan struct{}
}

// Name returns name of the loop.
func (l *TimeWheelLoop) Name() string {
	return l.name
}

// Conf configure the loop with the given configuration.
// This methods should be called before Run and not be called
// after the loop is running.
func (l *TimeWheelLoop) Conf(conf LoopConfig) bool {
	l.conf = conf

	if conf.Tick != 0 {
		l.tick.Reset(time.Millisecond * time.Duration(conf.Tick))
	}

	l.state = configured

	return true
}

// Run starts the loop with the provided hooks.
//
// NOTE: Run is not re-entrant and must not be called within a callback.
// When the loop is configured to run in async mode,
// hook functions must be guaranteed to be goroutine-safe.
func (l *TimeWheelLoop) Run(hooks LoopRunHook) {
	if l.Alive() {
		log.Warnf("loop %s alreay started", l.name)
		return
	}

	if hooks.Working == nil {
		log.Errorf("loop %s has not provide main callback func yet", l.name)
		return
	}

	l.state = running

	log.Infof("%s loop started", l.name)

	if l.conf.Sync {
		l.loop(&hooks)
	} else {
		go l.loop(&hooks)
	}
}

// Alive returns true if loop is running.
func (l *TimeWheelLoop) Alive() bool {
	return l.state == running
}

func (l *TimeWheelLoop) Stopped() bool {
	return l.state == stopped
}

// Stop stops the internal timer, close channels and clear all states.
func (l *TimeWheelLoop) Stop() {
	if l.Stopped() {
		return
	}

	close(l.quit)
	l.tick.Stop()

	<-l.wait

	l.state = stopped
}

func (l *TimeWheelLoop) runHook(pos string, hook func() error) bool {
	if err := hook(); err != nil && l.conf.BailOnError {
		log.Errorf("%s loop hook %s call failed: %v", l.name, pos, err)
		return false
	}

	//log.Tracef("%s loop hook %s called", l.name, pos)

	return true
}

func (l *TimeWheelLoop) loop(hooks *LoopRunHook) {
	var workCount uint32 = 0
	var idleCount uint32 = 0

	if hooks.BeforeRun != nil {
		if !l.runHook("BeforeQuit", hooks.BeforeRun) {
			return
		}
	}

	defer close(l.wait)

	for {
		select {
		case <-l.quit:
			if hooks.BeforeQuit != nil {
				if !l.runHook("BeforeQuit", hooks.BeforeQuit) {
					return
				}
			}

			log.Infof("%s loop quit", l.name)
			return
		case <-l.tick.C:
			workCount++
			idleCount++

			if l.conf.Work == 0 || workCount%l.conf.Work == 0 {
				if !l.runHook("Working", hooks.Working) {
					return
				}
			}

			if hooks.Stalling != nil && (l.conf.Idle == 0 || idleCount%l.conf.Idle == 0) {
				if !l.runHook("Stalling", hooks.Stalling) {
					return
				}
			}

			//if l.conf.Quit != 0 && workCount%l.conf.Quit == 0 {
			//	l.stopTrigger()
			//	log.Tracef("%s loop Quit threshold exceeds", l.name)
			//}

			//log.Tracef("%s loop cycle check done", l.name)
		}
	}
}
