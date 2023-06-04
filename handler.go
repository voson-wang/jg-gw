package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"reflect"
	"ricn-smart/ricn-jg-gw/modbus"
	"time"
)

const (
	timeout = 10 * time.Second
	size    = 500 // 设定读取数据的最大长度，必须大于设备发送的数据长度
)

const (
	RegisterCmd       = uint8(0x8b) // 设备注册
	HeartCmd          = uint8(0x8d) // 心跳包
	FaultCmd          = uint8(0x2a) // 故障信息
	UploadLiveDataCmd = uint8(0x64) // 实时数据上传

	ReadParamsCmd = uint8(0xca) // 远程参数读写
)

func handler(conn *modbus.Conn) {
	registerFrame, err := conn.Read(size, timeout)
	if err != nil {
		log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
		return
	}

	if registerFrame.Function == RegisterCmd {
		values := make(map[string]any)

		connReg.Decode(registerFrame.Data, values)

		snIF, ok := values[SNField.Name()]
		if !ok {
			log.Error().Msg("序列号获取失败")
			return
		}

		sn, ok := snIF.(string)
		if !ok {
			log.Error().Str("期望", "string").Interface("实际", reflect.TypeOf(snIF)).Msg("序列号类型错误")
			return
		}

		log.Info().Str("sn", sn).Str("remote", conn.Addr().String()).Msg("设备上线")

		if err := conn.Write(registerFrame, timeout); err != nil {
			log.Error().Err(err).Msg("")
			return
		}

		heartBeatFrame, err := conn.Read(size, timeout)
		if err != nil {
			log.Error().Err(err).Str("remote", conn.Addr().String()).Msg("")
			return
		}

		if heartBeatFrame.Function == HeartCmd {

			if err := conn.Write(heartBeatFrame, timeout); err != nil {
				log.Error().Err(err).Msg("")
				return
			}

			l := len(heartBeatFrame.Data)
			num := l / 6

			// 没有开关连接则退出
			if num == 0 {
				return
			}

			// 遥信－开关量    遥测－实时数据值
			for i := 0; i < num; i++ {

			}
		}

		log.Warn().Str("msg", fmt.Sprintf("% x", heartBeatFrame.Bytes())).Hex("Function", []byte{registerFrame.Function}).Str("remote", conn.Addr().String()).Msg("收到了注册包，但是没有心跳包")

		return
	}

	log.Warn().Hex("Function", []byte{registerFrame.Function}).Str("msg", fmt.Sprintf("% x", registerFrame.Bytes())).Str("remote", conn.Addr().String()).Msg("有新的链接，但是没有注册包")
}
