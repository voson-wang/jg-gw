package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	. "gopkg.in/check.v1"
	"ricn-smart/jg-gw/mq"
	"testing"
)

func TestMQ(t *testing.T) {
	TestingT(t)
}

type MQTestSuite struct{}

var _ = Suite(&MQTestSuite{})

const (
	sn            = "182112180128"
	childDeviceNo = "072107630289"
)

var (
	setActionRequest = &setPropertyRequest{
		RequestId:     "1",
		Identifiers:   []string{"OverCurrentTripSetting"},
		ChildDeviceNo: childDeviceNo,
		Params: map[string]any{
			"OverCurrentTripSetting": 70,
		},
	}

	getActionRequest = &getPropertyRequest{
		RequestId:     "2",
		Identifiers:   []string{"OverCurrentTripSetting"},
		ChildDeviceNo: childDeviceNo,
	}

	setControlRequest = &setPropertyRequest{
		RequestId:     "3",
		Identifiers:   []string{"Switch"},
		ChildDeviceNo: childDeviceNo,
		Params: map[string]any{
			"Switch": 1,
		},
	}
)

func (s *MQTestSuite) TestSetProperty(c *C) {
	quit := make(chan byte)
	opts := mq.Init(fmt.Sprintf("%v.%v", ProjectName, "TestSetProperty"))
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		if token := client.Subscribe(setActionRequest.RequestId, mq.AtMostOnce, func(client mqtt.Client, message mqtt.Message) {
			defer func() {
				quit <- 0
			}()
			var result CommonResponse
			if err := json.Unmarshal(message.Payload(), &result); err != nil {
				c.Fatal(err)
			}
			c.Assert(result.Success, Equals, true)
		}); token.Wait() && token.Error() != nil {
			log.Error().Err(token.Error()).Msg("")
		}

		if token := client.Subscribe(setControlRequest.RequestId, mq.AtMostOnce, func(client mqtt.Client, message mqtt.Message) {
			defer func() {
				quit <- 0
			}()
			var result CommonResponse
			if err := json.Unmarshal(message.Payload(), &result); err != nil {
				c.Fatal(err)
			}
			c.Assert(result.Success, Equals, true)
		}); token.Wait() && token.Error() != nil {
			log.Error().Err(token.Error()).Msg("")
		}
	})

	mq.Connect(opts)

	topic := ProjectName + "/" + sn + "/property/set"

	mq.Publish(topic, mq.AtMostOnce, false, setActionRequest)

	mq.Publish(topic, mq.AtMostOnce, false, setControlRequest)

	<-quit
	<-quit
}

func (s *MQTestSuite) TestGetProperty(c *C) {
	quit := make(chan byte)
	opts := mq.Init(fmt.Sprintf("%v.%v", ProjectName, "TestGetProperty"))
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		if token := client.Subscribe(getActionRequest.RequestId, mq.AtMostOnce, func(client mqtt.Client, message mqtt.Message) {
			defer func() {
				quit <- 0
			}()
			var result CommonResponse
			if err := json.Unmarshal(message.Payload(), &result); err != nil {
				c.Fatal(err)
			}
			c.Assert(result.Success, Equals, true)
			c.Assert(result.Data.(map[string]any)["OverCurrentTripSetting"], Equals, float64(70))
		}); token.Wait() && token.Error() != nil {
			log.Error().Err(token.Error()).Msg("")
		}
	})

	mq.Connect(opts)

	topic := ProjectName + "/" + sn + "/property/get"

	mq.Publish(topic, mq.AtMostOnce, false, getActionRequest)

	<-quit
}
