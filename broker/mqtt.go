package broker

import (
	mqtt "github.com/mochi-co/mqtt/server"
	"github.com/mochi-co/mqtt/server/events"
	"github.com/mochi-co/mqtt/server/listeners"
	"github.com/mochi-co/mqtt/server/listeners/auth"
	log "github.com/sirupsen/logrus"
	"time"
)

// MQTTServer provides an MQTT v3/v3.1.1 compatible broker impl.
type MQTTServer struct {
	*mqtt.Server
}

// NewMQTTServer creates a single node MQTT broker with the given name and
// listen address endpoint.
func NewMQTTServer(name string, endpoint string) *MQTTServer {
	// create the new MQTT Server
	server := &MQTTServer{
		Server: mqtt.New(),
	}

	// create a TCP listener on a standard port
	tcp := listeners.NewTCP(name, endpoint)

	// add the listener to the server with default options (nil)
	if err := server.AddListener(tcp, &listeners.Config{
		Auth: new(auth.Allow),
	}); err != nil {
		log.Errorln("mqtt broker add listener failed:", err)
		return nil
	}

	// Add OnConnect Event Hook
	server.Events.OnConnect = func(cl events.Client, pk events.Packet) {
		log.Infof("<<mqtt broker OnConnect client connected %s: %+v\n", cl.ID, pk)
	}

	// Add OnDisconnect Event Hook
	server.Events.OnDisconnect = func(cl events.Client, err error) {
		log.Infof("<<mqtt broker OnDisconnect client disconnected %s: %v\n", cl.ID, err)
	}

	log.Infoln("mqtt broker created")

	return server
}

// Startup starts the broker.
func (s *MQTTServer) Startup() error {
	// start the broker.
	// NOTE: Serve is non-blocking.
	if err := s.Serve(); err != nil {
		log.Errorln("serve mqtt broker failed:", err)
		return err
	}

	//wait for server ready
	time.Sleep(time.Millisecond * 500)

	log.Infoln("mqtt broker started")

	return nil
}

// Shutdown stops the broker.
func (s *MQTTServer) Shutdown() error {
	if err := s.Close(); err != nil {
		log.Errorln("stop mqtt broker failed:", err)
		return err
	}

	log.Infoln("mqtt broker stopped")
	return nil
}
