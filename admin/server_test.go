package admin

import (
	"github.com/zourva/pareto/mod"
	"testing"
)

func TestNewServer(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		server := NewServer("0.0.0.0:8888", mod.NewServiceManager())
		if server == nil {
			t.FailNow()
		}

		server.Run()
		//time.Sleep(3 * time.Second)
		select {}
	})
}
