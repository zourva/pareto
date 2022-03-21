package ntop

import (
	"errors"
	"github.com/eclipse/paho.golang/paho"
	"github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"time"
)

// MQTTClientV3 provides an MQTT v3/v3.1.1 compatible client impl.
type MQTTClientV3 struct {
	mqtt.Client
	network  string        //network type string mqtt resides on
	endpoint string        //endpoint address of mqtt broker
	addr     string        //local address cached
	stat     *statistician //internal statistician
}

//NewMQTTClientV3 creates a client and establish a connection to the given broker.
func NewMQTTClientV3(network string, endpoint string) *MQTTClientV3 {
	c := &MQTTClientV3{
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

func (c *MQTTClientV3) create() bool {
	return true
}

// Connect connects to broker using v3.1.1 first
// and then v3 when the first try failed.
// NOTE: this method use paho.Connect v5 Connect message to keep compatibility.
func (c *MQTTClientV3) Connect(connMsg *paho.Connect) error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.network + "://" + c.endpoint)
	opts.SetClientID(connMsg.ClientID)
	opts.SetCleanSession(connMsg.CleanStart)
	opts.SetKeepAlive(time.Duration(connMsg.KeepAlive) * time.Second)
	if connMsg.WillMessage != nil {
		opts.SetWill(connMsg.WillMessage.Topic,
			string(connMsg.WillMessage.Payload),
			connMsg.WillMessage.QoS,
			connMsg.WillMessage.Retain)

	}

	c.Client = mqtt.NewClient(opts)

	token := c.Client.Connect()
	if !token.WaitTimeout(time.Second * 5) {
		log.Errorln("connect to broker failed: timeout")
		return errors.New("connect timeout")
	}

	if !c.IsConnected() {
		log.Errorln("connect to broker failed: timeout")
		return errors.New("connect timeout")
	}

	c.addr = ""

	log.Infoln("mqtt client: connected to broker", c.endpoint)

	return nil
}

// Disconnect will end the connection with the server,
// before waiting 3000 ms to wait for cleaning.
func (c *MQTTClientV3) Disconnect() error {
	if c.IsConnected() {
		c.Client.Disconnect(3000)
	}

	return nil
}
