package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	logger "ricn-smart/jg-gw/log"
	"ricn-smart/jg-gw/modbus"
	"ricn-smart/jg-gw/mq"
	"ricn-smart/jg-gw/util"
	"syscall"
)

const port = 65010

var (
	GitCommitID string
	ProjectName string
	debug       = os.Getenv("DEBUG") == "true"
)

func init() {
	if ProjectName == "" {
		// 如果ProjectName不存在，则尝试读取go.mod中的module作为项目名
		ProjectName = util.GetProjectNameFromModule()
	}

	logger.Init(debug, fmt.Sprintf("log/%v.log", ProjectName))

}

func main() {
	ip, err := util.GetLocalIP()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	clientID := fmt.Sprintf("%v.%v", ProjectName, ip)

	// 因为会有多个应用实例运行在不同的主机上，因此不能使用可能重复的GitCommitID作为客户端ID
	opts := mq.Init(clientID)
	opts.SetOnConnectHandler(handleMQConn)
	mq.Connect(opts)

	server := modbus.NewServer(fmt.Sprintf(":%v", port))

	server.SetServe(handler)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("")
		}
	}()

	log.Info().Str("commit", GitCommitID).
		Bool("debug", debug).Int("port", port).Str("clientID", clientID).Msg(ProjectName + " started")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit
}
