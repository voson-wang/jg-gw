package mq

import (
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"jg-gw/config"
	"log"
	"os"
)

// SetOnConnectHandler 要先于Init调用
func SetOnConnectHandler(handler mqtt.OnConnectHandler) {
	onConn = handler
}

var onConn mqtt.OnConnectHandler
var client mqtt.Client

func Init() {
	server := os.Getenv("MQTT_ADDRESS")
	clientId := config.Host() + "/" + config.ProjectName()
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")

	if config.MQTTDebug() {
		mqtt.DEBUG = log.New(os.Stdout, "", 0)
		mqtt.ERROR = log.New(os.Stdout, "", 0)
	}

	opts := mqtt.NewClientOptions().
		AddBroker(server).
		SetClientID(clientId).
		SetUsername(username).
		SetPassword(password).
		SetOnConnectHandler(onConn)

	client = mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
}

func Publish(topic string, qos byte, retained bool, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if token := client.Publish(topic, qos, retained, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
