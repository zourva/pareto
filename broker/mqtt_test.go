package broker

import (
	"context"
	"fmt"
	"github.com/eclipse/paho.golang/paho"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var srv *MQTTServer

func BenchmarkPublish(b *testing.B) {
	client := NewMQTTClient("tcp", "127.0.0.1:1884")
	client.SetDebugLogger(&logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.TraceLevel,
	})

	err := client.Connect(&paho.Connect{
		ClientID:  "test-client",
		KeepAlive: 30,
		WillMessage: &paho.WillMessage{
			Retain:  true,
			QoS:     2,
			Topic:   "/mqtt/vehicle/die",
			Payload: nil,
		},
		CleanStart: true,
	})
	if err != nil {
		fmt.Println("err when connect:", err)
		return
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = client.Publish(context.Background(), &paho.Publish{
			QoS:     0,
			Retain:  false,
			Topic:   "/test",
			Payload: []byte("test"),
		})
	}
}

//func TestMain(m *testing.M) {
//	srv = NewMQTTServer("s1-broker", "127.0.0.1:1884")
//
//	_ = srv.Startup()
//	defer srv.Shutdown()
//
//	time.Sleep(time.Second)
//
//	m.Run()
//
//	fmt.Println("hold...")
//}

func TestMQTTClient_Connect(t *testing.T) {
	//client
	client := NewMQTTClientV3("tcp", "127.0.0.1:1884")
	assert.NotNil(t, client)

	//client.SetDebugLogger(&logrus.Logger{
	//	Out:       os.Stderr,
	//	Formatter: new(logrus.TextFormatter),
	//	Hooks:     make(logrus.LevelHooks),
	//	Level:     logrus.TraceLevel,
	//})

	err := client.Connect(&paho.Connect{
		ClientID:  "test-client",
		KeepAlive: 30,
		WillMessage: &paho.WillMessage{
			Retain:  true,
			QoS:     2,
			Topic:   "/mqtt/vehicle/die",
			Payload: nil,
		},
		CleanStart: true,
	})
	assert.Nil(t, err)

	TopicVehicleStatus := "/mqtt/vehicle/status"
	token := client.Subscribe(TopicVehicleStatus, 1, func(client mqtt.Client, message mqtt.Message) {
		fmt.Println("recv msg:", message)
	})

	assert.True(t, token.WaitTimeout(time.Second*5))

	//sa, err := client.Subscribe(context.Background(), &paho.Subscribe{
	//	Subscriptions: map[string]paho.SubscribeOptions{
	//		TopicVehicleStatus: {QoS: 0},
	//	},
	//})
	//if err != nil {
	//	fmt.Println("sa:", sa)
	//}
	//assert.Nil(t, err)
}
