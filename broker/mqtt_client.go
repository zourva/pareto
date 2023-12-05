package broker

import (
	"context"
	"errors"
	"github.com/eclipse/paho.golang/paho"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/conv"
	"net"
)

// MQTTClient provides an MQTT v5 compatible client impl.
type MQTTClient struct {
	*paho.Client
	network  string        //network type string mqtt resides on
	endpoint string        //endpoint address of mqtt broker
	addr     string        //local address cached
	stat     *statistician //internal statistician
}

// NewMQTTClient creates a client and establish a connection to the given broker.
func NewMQTTClient(network string, endpoint string) *MQTTClient {
	c := &MQTTClient{
		network:  network,
		endpoint: endpoint,
		stat:     newStatistician(),
	}

	if !c.create() {
		log.Errorln("mqtt client: init failed")
		return nil
	}

	return c
}

func (c *MQTTClient) create() bool {
	conn, err := net.Dial(c.network, c.endpoint)
	if err != nil {
		log.Errorf("mqtt client: failed to connect to %s: %s", c.endpoint, err)
		return false
	}

	mqc := paho.NewClient(paho.ClientConfig{
		Conn: conn,
		OnServerDisconnect: func(why *paho.Disconnect) {
			log.Warnln("mqtt client: receive DISCONNECT from mqtt broker:", *why)
		},
		OnClientError: func(err error) {
			log.Infoln("mqtt client: receive DISCONNECT from mqtt broker")
		},
	})

	mqc.Router = paho.NewStandardRouter()

	c.Client = mqc

	return true
}

// Connect send CONNECT msg to the broker
// on the established underlying network connection.
func (c *MQTTClient) Connect(connMsg *paho.Connect) error {
	ca, err := c.Client.Connect(context.Background(), connMsg)
	if err != nil {
		log.Errorf("mqtt client: connect to broker %s failed: %v", c.endpoint, err)

		if ca != nil {
			log.Errorf("mqtt client: reason: %d - %s", ca.ReasonCode, ca.Properties.ReasonString)
		}

		return err
	}

	if ca != nil && ca.ReasonCode != 0 {
		if ca.Properties != nil {
			log.Errorf("mqtt client: failed to connect to %s : %d - %s",
				c.endpoint, ca.ReasonCode, ca.Properties.ReasonString)
			return errors.New("mqtt connect failure: " + conv.Itoa(int(ca.ReasonCode)) + ca.Properties.ReasonString)
		}

		log.Errorf("mqtt client: failed to connect to %s : %d",
			c.endpoint, ca.ReasonCode)
		return errors.New("mqtt connect failure: " + conv.Itoa(int(ca.ReasonCode)))
	}

	c.addr = c.Client.Conn.LocalAddr().String()

	log.Infoln("mqtt client: connected to broker", c.endpoint)

	return nil
}

// Disconnect sends DISCONNECT msg to broker and
// always close the network connection
func (c *MQTTClient) Disconnect() error {
	return c.Client.Disconnect(&paho.Disconnect{ReasonCode: 0})
}
