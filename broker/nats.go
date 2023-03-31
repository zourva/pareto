package broker

import (
	"errors"
	"github.com/nats-io/nats-server/v2/server"
	log "github.com/sirupsen/logrus"
	"time"
)

// EmbeddedNats embeds the nats MQ server
// based on https://docs.nats.io/running-a-nats-service/clients#embedding-nats
// and https://dev.to/karanpratapsingh/embedding-nats-in-go-19o
type EmbeddedNats struct {
	ns *server.Server
}

// Startup starts the broker.
func (s *EmbeddedNats) Startup() error {
	//ns.Start is blocking, so run in a go routine.
	go s.ns.Start()

	//wait for ready
	for i := 1; i <= 3; i++ {
		if s.ns.ReadyForConnections(5 * time.Second) {
			break
		}

		log.Warnf("nats server is not ready yet, %d retry", i)

		if i == 3 {
			//failed, shutdown and quit
			log.Errorln("nats server startup timeout")
			s.ns.Shutdown()
			return errors.New("startup timeout")
		}
	}

	log.Infoln("nats broker started")

	return nil
}

// Shutdown stops the broker.
func (s *EmbeddedNats) Shutdown() error {
	s.ns.Shutdown()

	log.Infoln("nats broker stopped")
	return nil
}

func NewEmbeddedNats() *EmbeddedNats {
	opts := &server.Options{
		//Host: "0.0.0.0",
		//Port: 4222,
		//Authorization: "token",
	}
	ns, err := server.NewServer(opts)
	if err != nil {
		log.Errorln("create nats server failed:", err)
		return nil
	}

	srv := &EmbeddedNats{
		ns: ns,
	}

	log.Infoln("nats broker created")

	return srv
}
