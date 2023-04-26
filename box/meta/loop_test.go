package meta

import (
	log "github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestNewLoop(t *testing.T) {
	loop := NewLoop("system monitor", LoopConfig{
		Tick:        1000,  //tick interval, 1000 ms
		Work:        1,     //ticks triggering callbacks
		Sync:        false, //execute callback in a separate goroutine
		BailOnError: false, //no, bail only when asked to quit
	})

	loop.Run(LoopRunHook{Working: func() error {
		log.Infoln("checked")
		return nil
	}})

	time.Sleep(10 * time.Second)

	loop.Stop()
}
