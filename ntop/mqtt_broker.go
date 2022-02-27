package ntop

import (
	mqtt "github.com/mochi-co/mqtt/server"
	"github.com/mochi-co/mqtt/server/events"
	"github.com/mochi-co/mqtt/server/listeners"
	"github.com/mochi-co/mqtt/server/listeners/auth"
	log "github.com/sirupsen/logrus"
	"time"
)

type MQTTServer struct {
	*mqtt.Server
}

func NewMQTTServer(name string, endpoint string) *MQTTServer {
	// create the new MQTT Server
	server := &MQTTServer{
		Server: mqtt.New(),
	}

	// create a TCP listener on a standard port
	tcp := listeners.NewTCP(name, endpoint)

	// add the listener to the server with default options (nil)
	err := server.AddListener(tcp, &listeners.Config{
		Auth: new(auth.Allow),
	})
	if err != nil {
		log.Fatalln("MQTT broker add listener failed:", err)
	}

	// Add OnConnect Event Hook
	server.Events.OnConnect = func(cl events.Client, pk events.Packet) {
		log.Infof("<<MQTT broker OnConnect client connected %s: %+v\n", cl.ID, pk)
	}

	// Add OnDisconnect Event Hook
	server.Events.OnDisconnect = func(cl events.Client, err error) {
		log.Infof("<<MQTT broker OnDisconnect client disconnected %s: %v\n", cl.ID, err)
	}

	return server
}

func (s *MQTTServer) Start() {
	// start the broker in non-blocking mode
	go func() {
		err := s.Serve()
		if err != nil {
			log.Fatalln("MQTT broker serve failed:", err)
		}
	}()

	time.Sleep(time.Millisecond * 500)

	log.Infoln("MQTT broker started")
}

func (s *MQTTServer) Stop() {
	_ = s.Close()
	log.Infoln("MQTT broker stopped")
}
