package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"ricn-smart/jg-gw/modbus"
	"ricn-smart/jg-gw/mq"
	"time"
)

const (
	timeout = 60 * time.Second // 据观察，京硅设备心跳间隔在60s以内
	size    = 500              // 设定读取数据的最大长度，必须大于设备发送的数据长度
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

			mq.Publish(ProjectName+"/"+sn+"/event", mq.ExactlyOnce, false, map[string]any{"Identifier": "ONLINE"})

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
				if err := conn.Write(modbus.NewTelemetering(id), timeout); err != nil {
					log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
					return
				}

				telemeterAckFrame, err := conn.Read(size, timeout)
				if err != nil {
					log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
					return
				}

				data := make(map[string]any)

				if err := telemeterAckFrame.NewTelemeteringAck(data); err != nil {
					log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
					return
				}

				// 遥测读取电压等数据
				if err := conn.Write(modbus.NewTeleindication(id), timeout); err != nil {
					log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
					return
				}

				teleindicationAckFrame, err := conn.Read(size, timeout)
				if err != nil {
					log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
					return
				}

				if err := teleindicationAckFrame.NewTeleindicationAck(data); err != nil {
					log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
					return
				}

				log.Debug().Interface("data", data).Str("node", id.String()).Msg("开关和模拟量")

				mq.Publish(ProjectName+"/"+sn+"/"+id.String()+"/property", mq.AtMostOnce, false, data)
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
		case modbus.TeleFun:
			// 设备接收到其他途径（比如：485）的下发命令，会把其他途径下发的命令也发送给主站
			log.Debug().Msg("设备收到其他途径的遥信")

		default:
			log.Debug().Str("Function", fmt.Sprintf("0x%X", f.Function)).Str("Ctrl", fmt.Sprintf("0x%X", f.Ctrl)).Msg("未处理的命令码")
		}
	}

}
