package node

import (
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	server := NewServer("127.0.0.1:21985")
	go server.Start()

	time.Sleep(3 * time.Second)

	got := NewAgent("127.0.0.1:21985")
	got.Start()
	time.Sleep(30 * time.Second)
	got.Stop()

	//select {}
	time.Sleep(10 * time.Second)
	server.Stop(true)
}
