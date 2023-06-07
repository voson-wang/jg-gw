package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"ricn-smart/ricn-jg-gw/modbus"
	"time"
)

const (
	timeout = 60 * time.Second
	size    = 300 // 设定读取数据的最大长度，必须大于设备发送的数据长度
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

			// 扩展规约 6.1 原样回复给集中器
			if err := conn.Write(f, timeout); err != nil {
				log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
				return
			}

			sn := heartBeat.ID.String()

			log.Debug().Str("sn", sn).Str("node", modbus.NodesString(heartBeat.NodeIDs)).Msg("心跳包")

			for _, id := range heartBeat.NodeIDs {
				// 遥信读取开关状态
				telemeterFrame := modbus.NewTelemetering(id)

				if err := conn.Write(telemeterFrame, timeout); err != nil {
					log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
					return
				}

				telemeterAckFrame, err := conn.Read(500, timeout)
				if err != nil {
					log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
					return
				}

				telemeterAck, err := telemeterAckFrame.NewTelemeteringAck()
				if err != nil {
					log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
					return
				}

				log.Debug().Uint8(telemeterAck.Switches[0].Name, telemeterAck.Switches[0].Value).
					Uint8(telemeterAck.Switches[1].Name, telemeterAck.Switches[1].Value).Str("node", id.String()).Msg("遥信")
			}

		case modbus.PowerDownFun:
			log.Debug().Msg("掉电")
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
		case modbus.TelemeteringFun:
			// 设备接收到485的下发命令，会把485下发的命令也发送给主站
			log.Debug().Msg("遥信")

		default:

			log.Debug().Str("Function", fmt.Sprintf("0x%X", f.Function)).Msg("未处理的命令码")

		}
	}

}
