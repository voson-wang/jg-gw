package mq

import (
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"jg-gw/config"
	"log"
	"os"
	"strings"
	"time"
)

var client mqtt.Client

func Init(handler mqtt.OnConnectHandler) {
	address := os.Getenv("MQTT_ADDRESS")
	clientId := config.Host() + "/" + config.ProjectName()
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")

	if config.MQTTDebug() {
		mqtt.DEBUG = log.New(os.Stdout, "", 0)
		mqtt.ERROR = log.New(os.Stdout, "", 0)
	}

	opts := mqtt.NewClientOptions().
		SetClientID(clientId).
		SetUsername(username).
		SetPassword(password).
		SetOnConnectHandler(handler)

	for _, server := range strings.Split(address, ",") {
		opts.AddBroker(server)
	}

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
		select {
		// 无限期地等待令牌完成，即从代理发送发布和确认收据
		case <-token.Done():
			if token.Error() != nil {
				log.Println(token.Error())
			}
		case <-ticker.C:
			ticker.Stop()
			log.Println("发布超时")
		}
	}()
}
