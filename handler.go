package main

import (
	"github.com/rs/zerolog/log"
	"ricn-smart/ricn-jg-gw/modbus"
	"time"
)

const (
	timeout = 120 * time.Second
	size    = 500 // 设定读取数据的最大长度，必须大于设备发送的数据长度
)

func handler(conn *modbus.Conn) {
	f1, err := conn.Read(size, timeout)
	if err != nil {
		log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
		return
	}

	switch f1.Function {
	case modbus.RegisterFun:

		login, err := f1.NewLogin()
		if err != nil {
			log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
			return
		}

		sn := login.ID.String()

		log.Info().Str("sn", sn).Msg("设备上线")

	case modbus.PowerDownFun:

	case modbus.FaultFun:
		fault, err := f1.NewFault()
		if err != nil {
			log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
			return
		}
		faultAckFrame := f1.NewFaultAck(fault)
		// 回复确认
		if err := conn.Write(faultAckFrame, timeout); err != nil {
			log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
			return
		}
	}

}
