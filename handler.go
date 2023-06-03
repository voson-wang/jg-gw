package main

import (
	"github.com/rs/zerolog/log"
	"ricnsmart/jg-gw/modbus"
	"time"
)

const (
	timeout = 10 * time.Second
	size    = 500 // 设定读取数据的最大长度，必须大于设备发送的数据长度
)

func handler(conn *modbus.Conn) {

	registerFrame, err := conn.Read(size, timeout)
	if err != nil {
		log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
		return
	}

	log.Debug().Hex("Function", []byte{registerFrame.Function}).Msg("register")
}
