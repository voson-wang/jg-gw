package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"jg-gw/config"
	"jg-gw/mq"
	"log"
	"os"
	"testing"
)

const sn = "182112180128"
const lineNo = "072107630284"

func Test_setProperty(t *testing.T) {

	_ = os.Setenv("PROJECT_NAME", "Test_setProperty")

	config.Init()

	topic := "jg-gw/" + sn + "/" + lineNo + "/property/set"

	tests := []struct {
		name string
		args setPropertyRequest
	}{
		{
			name: "identifier缺失",
			args: setPropertyRequest{
				RequestId:   config.Host() + "s0",
				LineNo:      lineNo,
				Identifiers: []string{},
				Params:      map[string]any{"OverLoadValue": 300},
			},
		},
		{
			name: "参数缺失",
			args: setPropertyRequest{
				RequestId:   config.Host() + "s1",
				LineNo:      lineNo,
				Identifiers: []string{"OverLoadValue"},
				Params:      map[string]any{},
			},
		},
		{
			name: "参数类型错误",
			args: setPropertyRequest{
				RequestId:   config.Host() + "s2",
				LineNo:      lineNo,
				Identifiers: []string{"OverLoadValue"},
				Params:      map[string]any{"OverLoadValue": 200.1},
			},
		},
		{
			name: "正常写入1个普通寄存器",
			args: setPropertyRequest{
				RequestId:   config.Host() + "s3",
				LineNo:      lineNo,
				Identifiers: []string{"OverVoltageValue"},
				Params:      map[string]any{"OverVoltageValue": 260},
			},
		},
		{
			name: "Status",
			args: setPropertyRequest{
				RequestId:   config.Host() + "s6",
				LineNo:      lineNo,
				Identifiers: []string{"Status"},
				Params:      map[string]any{"Status": 1},
			},
		},
	}

	onConn := func(client mqtt.Client) {
		client.Subscribe(config.Host()+"s0", 2, func(client mqtt.Client, message mqtt.Message) {
			got := string(message.Payload())
			fmt.Println(got)

		})

		client.Subscribe(config.Host()+"s1", 2, func(client mqtt.Client, message mqtt.Message) {
			got := string(message.Payload())
			fmt.Println(got)

		})

		client.Subscribe(config.Host()+"s2", 2, func(client mqtt.Client, message mqtt.Message) {
			got := string(message.Payload())
			fmt.Println(got)

		})

		client.Subscribe(config.Host()+"s3", 2, func(client mqtt.Client, message mqtt.Message) {
			got := string(message.Payload())
			fmt.Println(got)

		})

		client.Subscribe(config.Host()+"s4", 2, func(client mqtt.Client, message mqtt.Message) {
			got := string(message.Payload())
			fmt.Println(got)

		})

		client.Subscribe(config.Host()+"s5", 2, func(client mqtt.Client, message mqtt.Message) {
			got := string(message.Payload())
			fmt.Println(got)

		})

		client.Subscribe(config.Host()+"s6", 2, func(client mqtt.Client, message mqtt.Message) {
			got := string(message.Payload())
			fmt.Println(got)

		})
	}

	mq.Init(onConn)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mq.Publish(topic, 2, false, tt.args)
		})
	}

	quit := make(chan int)
	<-quit
}

func Test_invokeService(t *testing.T) {

	_ = os.Setenv("PROJECT_NAME", "Test_invokeService")

	config.Init()

	topic := "jg-gw/" + sn + "/service"

	tests := []struct {
		name string
		args invokeServiceRequest
	}{
		{
			name: "标识符缺失",
			args: invokeServiceRequest{
				RequestId:  config.Host() + "v0",
				Identifier: "",
			},
		},
		{
			name: "主动断开连接",
			args: invokeServiceRequest{
				RequestId:  config.Host() + "v1",
				Identifier: "Disconnect",
			},
		},
		{
			name: "绑定开关",
			args: invokeServiceRequest{
				RequestId:  config.Host() + "v3",
				Identifier: "BindSwitch",
				Params:     map[string]any{"SN": sn, "LineNo": []string{lineNo}},
			},
		},
	}

	onConn := func(client mqtt.Client) {
		if token := client.Subscribe(config.Host()+"v0", 2, func(client mqtt.Client, message mqtt.Message) {
			got := string(message.Payload())
			fmt.Println(got)

		}); token.Wait() && token.Error() != nil {
			log.Fatal(token.Error())
		}

		if token := client.Subscribe(config.Host()+"v1", 2, func(client mqtt.Client, message mqtt.Message) {
			got := string(message.Payload())
			fmt.Println(got)

		}); token.Wait() && token.Error() != nil {
			log.Fatal(token.Error())
		}

		if token := client.Subscribe(config.Host()+"v2", 2, func(client mqtt.Client, message mqtt.Message) {
			got := string(message.Payload())
			fmt.Println(got)

		}); token.Wait() && token.Error() != nil {
			log.Fatal(token.Error())
		}

		if token := client.Subscribe(config.Host()+"v3", 2, func(client mqtt.Client, message mqtt.Message) {
			got := string(message.Payload())
			fmt.Println(got)

		}); token.Wait() && token.Error() != nil {
			log.Fatal(token.Error())
		}

	}

	mq.Init(onConn)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mq.Publish(topic, 2, false, tt.args)
		})
	}

	quit := make(chan int)
	<-quit
}
