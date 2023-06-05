package main

import (
	"github.com/rs/zerolog/log"
	"ricn-smart/ricn-jg-gw/modbus"
	"strings"
	"time"
)

const (
	timeout = 120 * time.Second
	size    = 500 // 设定读取数据的最大长度，必须大于设备发送的数据长度
)

func handler(conn *modbus.Conn) {

	for {
		f, err := conn.Read(size, timeout)
		if err != nil {
			log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
			return
		}

		switch f.Function {
		case modbus.RegisterFun:
			login, err := f.NewLogin()
			if err != nil {
				log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
				return
			}

			sn := login.ID.String()

			log.Info().Str("sn", sn).Msg("上线")

		case modbus.HeartBeatFun:
			heartBeat, err := f.NewHeartBeat()
			if err != nil {
				log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
				return
			}

			sn := heartBeat.ID.String()

			var node []string
			for _, id := range heartBeat.NodeID {
				node = append(node, id.String())
			}

			log.Debug().Str("sn", sn).Str("node", strings.Join(node, ",")).Msg("心跳包")

		case modbus.PowerDownFun:

		case modbus.FaultFun:
			fault, err := f.NewFault()
			if err != nil {
				log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
				return
			}
			log.Debug().Time("time", fault.TelemeteringTimeMark.Time()).Msg("故障")
			faultAckFrame := f.NewFaultAck(fault)
			// 回复确认
			if err := conn.Write(faultAckFrame, timeout); err != nil {
				log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
				return
			}
		}
	}

}
