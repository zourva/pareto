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
	//clock pulse interval in milliseconds, 0 means using the default(1000 ms)
	Tick uint32

	//#ticks of work check interval,  set to 0 as using the default(1)
	Work uint32

	//#ticks of idle check interval, set to 0 to use the default(1)
	Idle uint32

	//#ticks after which the loop will be terminated,
	// set to 0 as disabling auto termination
	//Quit uint32

	//if false, fn will be executed in another go routine, false by default
	Sync bool

	//if true, quit the underlying loop when error returned by user hooks, false by default
	BailOnError bool
}

// Loop interface exposed
type Loop interface {
	Name() string
	Conf(conf LoopConfig) bool
	Run(hooks LoopRunHook)
	Alive() bool
	Stop()
}

// NewLoop creates a new loop object with the given name and conf.
func NewLoop(name string, conf LoopConfig) Loop {
	ticks := time.Millisecond * 1000

	if conf.Tick != 0 {
		ticks = time.Millisecond * time.Duration(conf.Tick)
	}

	return &BreakableLoop{
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

//BreakableLoop provides a simple breakable loop impl.
type BreakableLoop struct {
	name  string
	state uint32
	conf  LoopConfig
	tick  *time.Ticker
	quit  chan struct{}
	wait  chan struct{}
}

// Name returns name of the loop.
func (l *BreakableLoop) Name() string {
	return l.name
}

// Conf configure the loop with the given configuration.
// This methods should be called before Run and not be called
// after the loop is running.
func (l *BreakableLoop) Conf(conf LoopConfig) bool {
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
func (l *BreakableLoop) Run(hooks LoopRunHook) {
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
func (l *BreakableLoop) Alive() bool {
	return l.state == running
}

// Stop stops the internal timer, close channels and clear all states.
func (l *BreakableLoop) Stop() {
	close(l.quit)
	l.tick.Stop()

	<-l.wait

	l.state = stopped
}

func (l *BreakableLoop) runHook(pos string, hook func() error) bool {
	if err := hook(); err != nil && l.conf.BailOnError {
		log.Errorf("%s loop %s hook call failed: %v", l.name, pos, err)
		return false
	}

	log.Debugf("%s loop %s hook called", pos, l.name)

	return true
}

func (l *BreakableLoop) loop(hooks *LoopRunHook) {
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

			log.Tracef("%s loop cycle check done", l.name)
		}
	}
}
