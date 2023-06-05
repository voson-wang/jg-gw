package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	logger "ricn-smart/ricn-jg-gw/log"
	"ricn-smart/ricn-jg-gw/modbus"
	"ricn-smart/ricn-jg-gw/mq"
	"ricn-smart/ricn-jg-gw/util"
	"syscall"
)

const port = 65010

var (
	GitCommitID string
	ProjectName string
	ip          string
	debug       = os.Getenv("DEBUG") == "true"
)

func init() {
	var err error
	ip, err = util.GetLocalIP()
	if err != nil {
		panic(err)
	}

	if ProjectName == "" {
		// 如果ProjectName不存在，则尝试读取go.mod中的module作为项目名
		ProjectName = util.GetProjectNameFromModule()
	}

	logger.Init(debug, fmt.Sprintf("log/%v.log", ProjectName))

	opts := mq.Init(fmt.Sprintf("%v.%v", ProjectName, ip))

	mq.Connect(opts)
}

func main() {

	server := modbus.NewServer(fmt.Sprintf(":%v", port))

	server.SetServe(handler)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("")
		}
	}()

	log.Info().Str("commit", GitCommitID).Str("ip", ip).
		Bool("debug", debug).Int("port", port).Msg(ProjectName + " started")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit
}
