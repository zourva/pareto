package ntop

import (
	"context"
	"errors"
	"github.com/eclipse/paho.golang/paho"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"net"
	"time"
)

type MQTTClient struct {
	*paho.Client
	network  string        //network type string mqtt resides on
	endpoint string        //endpoint address of mqtt broker
	addr     string        //local address cached
	stat     *statistician //internal statistician
}

//NewMQTTClient creates a client and establish a connection to the given broker.
func NewMQTTClient(network string, endpoint string) *MQTTClient {
	c := &MQTTClient{
		network:  network,
		endpoint: endpoint,
		stat:     newStatistician(),
	}

	if !c.create() {
		log.Errorln("MQTT client: init failed")
		return nil
	}

	return c
}

func (c *MQTTClient) create() bool {
	conn, err := net.Dial(c.network, c.endpoint)
	if err != nil {
		log.Errorf("MQTT client: failed to connect to %s: %s", c.endpoint, err)
		return false
	}

	mqc := paho.NewClient(paho.ClientConfig{
		Conn: conn,
		OnServerDisconnect: func(why *paho.Disconnect) {
			log.Warnln("MQTT client: receive DISCONNECT from MQTT broker:", *why)
		},
		OnClientError: func(err error) {
			log.Infoln("MQTT client: receive DISCONNECT from MQTT broker")
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
		log.Errorf("MQTT client: connect to ni server %s failed: %v, %v\n", c.endpoint, err)

		if ca != nil {
			log.Errorf("MQTT client: reason: %d - %s\n", ca.ReasonCode, ca.Properties.ReasonString)
		}

		return err
	}

	if ca != nil && ca.ReasonCode != 0 {
		if ca.Properties != nil {
			log.Errorf("MQTT client: failed to connect to %s : %d - %s\n",
				c.endpoint, ca.ReasonCode, ca.Properties.ReasonString)
			return errors.New("MQTT connect failure: " + box.Itoa(int(ca.ReasonCode)) + ca.Properties.ReasonString)
		} else {
			log.Errorf("MQTT client: failed to connect to %s : %d\n",
				c.endpoint, ca.ReasonCode)
			return errors.New("MQTT connect failure: " + box.Itoa(int(ca.ReasonCode)))
		}
	}

	c.addr = c.Client.Conn.LocalAddr().String()

	log.Infoln("MQTT client: connected to broker", c.endpoint)

	return nil
}

// Disconnect sends DISCONNECT msg to broker and
// always close the network connection
func (c *MQTTClient) Disconnect() error {
	return c.Client.Disconnect(&paho.Disconnect{ReasonCode: 0})
}

type counters struct {
	bytesSent uint64 //number bytes sent
	bytesRecv uint64 //number bytes received
	requests  uint64 //number requests sent
	replies   uint64 //number replies received
}

func (c *counters) clear() {
	c.bytesSent = 0
	c.bytesRecv = 0
	c.requests = 0
	c.replies = 0
}

type kpis struct {
	rps    uint64
	ulRate uint64
	dlRate uint64
	delay  int64
}

func (k *kpis) clear() {
	k.rps = 0
	k.ulRate = 0
	k.dlRate = 0
	k.delay = 0
}

type statSession struct {
	startTime int64
	stopTime  int64
	counters  counters
}

func (s *statSession) hackStart(szSent, nrReq uint64) {
	s.startTime = time.Now().Unix()
	s.counters.requests += nrReq
	s.counters.bytesSent += szSent
}

func (s *statSession) hackStop(szRecv, nrRep uint64) {
	s.stopTime = time.Now().Unix()
	s.counters.bytesRecv += szRecv
	s.counters.replies += nrRep
}

type statistician struct {
	sampleTime int64
	counters   counters
	kpis       kpis
}

func newStatistician() *statistician {
	s := &statistician{
		sampleTime: time.Now().Unix(),
	}

	return s
}

func (s *statistician) session() *statSession {
	return &statSession{}
}

func (s *statistician) updateCounters(ssn *statSession) {
	s.kpis.delay = box.MaxI64(s.kpis.delay, ssn.stopTime-ssn.startTime)
	s.counters.bytesSent += ssn.counters.bytesSent
	s.counters.bytesRecv += ssn.counters.bytesRecv
	s.counters.requests += ssn.counters.requests
	s.counters.replies += ssn.counters.replies
}

func (s *statistician) sample(reset bool) *kpis {
	now := time.Now().Unix()
	duration := now - s.sampleTime
	log.Traceln("now, start, duration", now, s.sampleTime, duration)
	if duration == 0 {
		log.Traceln("sample time too short, skip this round")
		return &s.kpis
	}

	s.kpis.rps = s.counters.requests / uint64(duration)
	s.kpis.dlRate = s.counters.bytesRecv / uint64(duration)
	s.kpis.ulRate = s.counters.bytesSent / uint64(duration)

	if reset {
		s.kpis.clear()
		s.counters.clear()
	}

	s.sampleTime = now

	return &s.kpis
}
