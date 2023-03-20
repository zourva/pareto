package node

import (
	"github.com/zourva/pareto/mod"
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	type args struct {
		endpoint string
		opts     []AgentOption
	}
	tests := []struct {
		name string
		args args
		//want *Agent
	}{
		{
			name: "default",
			args: args{
				endpoint: "127.0.0.1:21985",
				opts:     []AgentOption{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(tt.args.endpoint, mod.NewServiceManager())
			go server.Start()

			time.Sleep(3 * time.Second)

			got := NewAgent(tt.args.endpoint, mod.NewServiceManager(), tt.args.opts...)
			got.Start()
			time.Sleep(30 * time.Second)
			got.Stop()

			//select {}
			time.Sleep(10 * time.Second)
			server.Stop(true)
		})
	}
}
