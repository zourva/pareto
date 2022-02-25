package meta

import (
	log "github.com/sirupsen/logrus"
	"time"
)

// for ref. http://docs.libuv.org/en/v1.x/loop.html
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

// loop configuration, all options have default values except CbWork
type LoopConfig struct {
	//clock pulse interval in milliseconds, 0 means using the default(1000 ms)
	Tick uint32

	//#ticks of work check interval, 0 means using the default(0)
	Work uint32

	//#ticks of idle check interval, 0 means using the default(0)
	Idle uint32

	//if false, fn will be executed in another go routine, false by default
	Sync bool

	//if true, quit the underlying loop when error returned by user hooks, false by default
	BailOnError bool
}

type Loop interface {
	Name() string
	Conf(conf LoopConfig) bool
	Run(hooks LoopRunHook)
	Alive() bool
	Stop()
}

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

//simple breakable loop impl.
type BreakableLoop struct {
	name  string
	state uint32
	conf  LoopConfig
	tick  *time.Ticker
	quit  chan struct{}
	wait  chan struct{}
}

func (l *BreakableLoop) Name() string {
	return l.name
}

func (l *BreakableLoop) Conf(conf LoopConfig) bool {
	l.conf = conf

	if conf.Tick != 0 {
		l.tick.Reset(time.Millisecond * time.Duration(conf.Tick))
	}

	l.state = configured

	return true
}

// not reentrant, must not be called from a callback.
// NOTE: when the loop is configured to run in async mode,
//user must make sure hook functions are goroutine-safe
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

func (l *BreakableLoop) Alive() bool {
	return l.state == running
}

func (l *BreakableLoop) Stop() {
	close(l.quit)
	l.tick.Stop()

	<-l.wait

	l.state = stopped
}

func (l *BreakableLoop) loop(hooks *LoopRunHook) {
	var workCount uint32 = 0
	var idleCount uint32 = 0

	if hooks.BeforeRun != nil {
		if err := hooks.BeforeRun(); err != nil && l.conf.BailOnError {
			log.Errorf("%s loop BeforeQuit hook call failed: %v", l.name, err)
			return
		}

		log.Infof("%s loop BeforeRun hook called", l.name)
	}

	defer close(l.wait)

	for {
		select {
		case <-l.quit:
			if hooks.BeforeQuit != nil {
				if err := hooks.BeforeQuit(); err != nil && l.conf.BailOnError {
					log.Errorf("%s loop BeforeQuit hook call failed: %v", l.name, err)
					return
				}
				log.Debugf("%s loop BeforeQuit hook called", l.name)
			}

			log.Infof("%s loop quit", l.name)
			return
		case <-l.tick.C:
			workCount++
			idleCount++

			if l.conf.Work == 0 || workCount%l.conf.Work == 0 {
				if err := hooks.Working(); err != nil && l.conf.BailOnError {
					log.Errorf("%s loop Working hook call failed: %v", l.name, err)
					return
				}
				log.Tracef("%s loop Working hook called", l.name)
			}

			if (l.conf.Idle == 0 || idleCount%l.conf.Idle == 0) && hooks.Stalling != nil {
				if err := hooks.Stalling(); err != nil && l.conf.BailOnError {
					log.Errorf("%s loop Stalling hook call failed: %v", l.name, err)
					return
				}
				log.Tracef("%s loop Stalling hook called", l.name)
			}

			log.Tracef("%s loop cycle check done", l.name)
		}
	}
}
