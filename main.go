package main

import (
	"e.coding.net/ricnsmart/service/jg-modbus"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"jg-gateway/config"
	"jg-gateway/mq"
	"jg-gateway/util"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 65010

var server *jg.Server

func main() {

	config.Init()

	util.InitLog()

	if config.Env() == "development" {
		util.PProf()
	}

	mq.SetOnConnectHandler(handleMQConn)

	mq.Init()

	server = jg.NewServer(fmt.Sprintf(":%v", port), handleUploadPacket)

	server.SetOnConnClose(afterConnClose)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("")
		}
	}()

	log.Info().Str("version", os.Getenv("version")).Interface("env", config.Env()).
		Bool("debug", config.Debug()).Str("host", config.Host()).Int("port", port).Msg(config.ProjectName() + " running")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit
}

type getHostRequest struct {
	RequestId string
}

// handleMQConn
// mqtt连接上后开始执行订阅
func handleMQConn(client mqtt.Client) {

	// 设备host查询
	if token := client.Subscribe("+/host", 1, func(client mqtt.Client, message mqtt.Message) {
		topic := message.Topic()

		arr := strings.Split(topic, "/")

		sn := arr[0]

		go getHost(sn, client, message.Payload())

		log.Debug().Str("sn", sn).Bytes("request", message.Payload()).Msg("invokeService")
	}); token.Wait() && token.Error() != nil {
		log.Error().Err(token.Error()).Msg("")
	}

	// 调用服务
	if token := client.Subscribe(config.ProjectName()+"/+/service", 2, func(client mqtt.Client, message mqtt.Message) {
		topic := message.Topic()

		arr := strings.Split(topic, "/")

		sn := arr[1]

		go invokeService(sn, message.Payload(), client)

		log.Debug().Str("sn", sn).Bytes("request", message.Payload()).Msg("invokeService")
	}); token.Wait() && token.Error() != nil {
		log.Error().Err(token.Error()).Msg("")
	}

	// 设置属性
	// 相同型号的应用共享请求，使用序列号来找出所在应用
	if token := client.Subscribe(config.ProjectName()+"/+/+/property/set", 2, func(client mqtt.Client, message mqtt.Message) {
		topic := message.Topic()

		arr := strings.Split(topic, "/")

		sn := arr[1]

		lineNo := arr[2]

		go setProperty(sn, lineNo, message.Payload(), client)

		log.Debug().Str("sn", sn).Bytes("request", message.Payload()).Msg("setProperty")
	}); token.Wait() && token.Error() != nil {
		log.Error().Err(token.Error()).Msg("")
	}
}

type Event struct {
	Identifier string                 `json:"identifier"`
	Params     map[string]interface{} `json:"params"`
}

func afterConnClose(addr net.Addr) {
	sn := findSN(addr)
	if sn != "" {
		snAndRemoteAddrMap.Delete(sn)
		snAndDataMap.Delete(sn)
		log.Warn().Str("sn", sn).Msg("设备离线")
		if err := mq.Publish(config.ProjectName()+"/"+sn+"/event", 1, false, &Event{
			Identifier: "OFFLINE",
		}); err != nil {
			log.Error().Err(err).Msg("")
		}
	}
}
