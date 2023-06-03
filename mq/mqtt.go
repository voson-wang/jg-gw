package mq

import (
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"time"
)

const (
	AtMostOnce byte = iota
	AtLeastOnce
	ExactlyOnce
)

var client mqtt.Client

func Init(clientId string) *mqtt.ClientOptions {
	address := os.Getenv("MQTT_ADDRESS")
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")

	if os.Getenv("MQTT_DEBUG") == "true" {
		mqtt.DEBUG = log.New(os.Stdout, "", 0)
		mqtt.ERROR = log.New(os.Stdout, "", 0)
	}

	return mqtt.NewClientOptions().
		SetClientID(clientId).
		SetUsername(username).
		SetPassword(password).
		SetResumeSubs(true).AddBroker(address)
}

func Connect(opts *mqtt.ClientOptions) {
	client = mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
}

func Publish(topic string, qos byte, retained bool, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}
	token := client.Publish(topic, qos, retained, payload)
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		select {
		// 无限期地等待令牌完成，即从代理发送发布和确认收据
		case <-token.Done():
			if token.Error() != nil {
				log.Println(token.Error())
			}
		case <-ticker.C:
			log.Println("发布超时")
		}
	}()
}
