package main

import (
	"context"
	"e.coding.net/ricnsmart/service/jg-modbus"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"jg-gw/config"
	"jg-gw/mq"
	"jg-gw/util"
	"net"
	"reflect"
	"sync"
	"time"
)

const (
	RegisterCmd       = uint8(0x8b) // 设备注册
	HeartCmd          = uint8(0x8d) // 心跳包
	FaultCmd          = uint8(0x2a) // 故障信息
	UploadLiveDataCmd = uint8(0x64) // 实时数据上传

	AlertSettingCmd = uint8(0xca)
)

var livedataCfg = []byte{0x80, 0x06, 0x00, 0x00, 0x00, 0x01, 0x40}    // 读取实时数据固定参数
var statusCfg = []byte{0x80, 0x06, 0x00, 0x00, 0x00, 0x01, 0x00}      // 读取开关状态固定参数
var writeStatusCfg = []byte{0x81, 0x06, 0x00, 0x00, 0x00}             // 写入开关状态固定参数
var readCfg = []byte{0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00}        // 读取数据固定参数
var writeCfg = []byte{0x01, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00} // 写入数据固定参数
var faultCfg = []byte{0x00, 0x03, 0x01, 0x00, 0x00}                   // 故障信息回传固定参数

var snAndRemoteAddrMap sync.Map

var snAndDataMap sync.Map

func findSN(addr any) (sn string) {
	snAndRemoteAddrMap.Range(func(key, value any) bool {
		if value == addr {
			sn = key.(string)
			return false
		}
		return true
	})
	return
}

// handleUploadPacket 处理设备上报的数据包
func handleUploadPacket(addr net.Addr, data []byte, answer func(payload []byte) error) {
	frame, err := jg.NewFrame(data)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	switch frame.Function {
	case RegisterCmd:
		// 设备每次上线都会发送一次注册信息

		values := make(map[string]any)

		connReg.Decode(frame.Data, values)

		snIF, ok := values[SNField.Name()]
		if !ok {
			log.Error().Msg("序列号获取失败")
			return
		}

		// 默认sn是string类型
		snAndRemoteAddrMap.Store(snIF, addr)

		sn, ok := snIF.(string)
		if !ok {
			log.Error().Str("期望", "string").Interface("实际", reflect.TypeOf(snIF)).Msg("序列号类型错误")
			return
		}

		log.Debug().Interface("data", values).Msg("注册包")

		log.Info().Str("sn", sn).Msg("设备上线")

		if err := mq.Publish(config.ProjectName()+"/"+sn+"/event", 1, false, &Event{
			Identifier: "ONLINE",
		}); err != nil {
			log.Error().Err(err).Msg("")
		}

	case HeartCmd:

		sn := findSN(addr)

		l := len(frame.Data)
		num := l / 6

		// 没有开关连接则退出
		if num == 0 {
			return
		}

		// 遥信－开关量    遥测－实时数据值
		for i := 0; i < num; i++ {
			frame.Data = frame.Data[6:]

			// 遥测获取开关实时数据值
			liveData, err := getLivedata(addr, frame)
			if err != nil {
				log.Debug().Err(err).Msg("get livedata failed")
				continue
			}

			// 遥信获取状态
			status, err := getStatus(addr, frame)
			if err != nil {
				log.Debug().Err(err).Msg("get status failed")
				continue
			}

			// 获取警报配置
			alarm, err := getAlarmSetting(addr, frame)
			if err != nil {
				log.Debug().Err(err).Msg("get alarmSetting failed")
				continue
			}

			for k, v := range alarm {
				liveData[k] = v
			}

			liveData["Switch"] = status
			lineNo := util.BytesToString(frame.Data[:6])
			lineModel, err := util.GetLineModel(lineNo)
			if err != nil {
				log.Error().Err(err).Msg("get lineModel failed")
				continue
			}
			if err := mq.Publish(config.ProjectName()+"/"+sn+"/"+lineModel+"/"+lineNo+"/property", 1, false, liveData); err != nil {
				log.Error().Err(err).Msg("")
			}
			log.Debug().Interface("liveData", liveData).Msg("设备实时数据")
		}

	case FaultCmd:
		//消除故障信息上传
		c := &jg.Frame{
			Address:  frame.Address,
			Function: 0x2A,
			Cfg:      faultCfg,
			Data:     frame.Data[5:17],
		}

		if err := answer(c.Bytes()); err != nil {
			log.Error().Err(err).Msg("消除故障信息失败")
		}
	default:
		log.Warn().Uint8("Function", frame.Function).Interface("data", fmt.Sprintf("0x%0 x", data)).Msg("未知的命令字")
	}
}

func getLivedata(addr net.Addr, f *jg.Frame) (map[string]interface{}, error) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Err(fmt.Errorf("%v", err)).Msg("GetLiveDataFailed")
		}
	}()

	c := &jg.Frame{
		Function: UploadLiveDataCmd,
		Address:  f.Data[:6],
		Cfg:      livedataCfg,
		Data:     []byte{0x20},
	}
	cmd := c.Bytes()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	out, err := server.DownloadCommand(ctx, addr, cmd)
	if err != nil {
		return nil, err
	}
	livedata, err := jg.NewFrame(out)
	if err != nil {
		return nil, err
	}

	// 忽略接收到的心跳包
	l := len(livedata.Data)
	if l < livedataReg.Len() || livedata.Function != UploadLiveDataCmd {
		return nil, fmt.Errorf("未正确获取实时数据")
	}
	livedataMap := make(map[string]interface{})

	livedataReg.Decode(livedata.Data, livedataMap)
	return livedataMap, nil
}

func getStatus(addr net.Addr, f *jg.Frame) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error().Err(fmt.Errorf("%v", err)).Msg("GetStatusFailed")
		}
	}()

	cfg := statusCfg
	c := &jg.Frame{
		Function: 0x64,
		Address:  f.Data[:6],
		Cfg:      cfg,
		Data:     []byte{0x20},
	}
	cmd := c.Bytes()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	out, err := server.DownloadCommand(ctx, addr, cmd)
	if err != nil {
		return nil, err
	}
	switchData, err := jg.NewFrame(out)
	if err != nil {
		return nil, err
	}

	if switchData.Function != UploadLiveDataCmd {
		return nil, fmt.Errorf("未正确获取开关状态")
	}

	switchDataMap := SwitchPacket.Decode(switchData.Data[8:])
	return switchDataMap[Status.Name()], nil
}

func getAlarmSetting(addr net.Addr, f *jg.Frame) (map[string]interface{}, error) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Err(fmt.Errorf("%v", err)).Msg("GetAlarmSettingFailed")
		}
	}()

	cfg := readCfg
	cfg[0] = byte(AlarmSettingPacket.Len())
	cmd, err := AlarmSettingPacket.NewReadRegisters(AlertSettingCmd, cfg, f.Data[:6])

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	out, err := server.DownloadCommand(ctx, addr, cmd)
	if err != nil {
		return nil, err
	}
	settingData, err := jg.NewFrame(out)
	if err != nil {
		return nil, err
	}

	if settingData.Function != AlertSettingCmd {
		return nil, fmt.Errorf("未正确获取警报设定值")
	}

	data := settingData.Data[8:]
	alarmData := make([]byte, 0)
	for i := 0; i < len(AlarmSettingPacket); i++ {
		alarmData = append(alarmData, data[i*6+4])
		alarmData = append(alarmData, data[i*6+5])
	}
	alarmSettingMap := AlarmSettingPacket.Decode(alarmData)
	return alarmSettingMap, nil
}

type CommonResponse struct {
	RequestId string      `json:"request_id"`
	Success   bool        `json:"success"` // 调用结果是否成功
	Message   string      `json:"message"` // 消息;
	Data      interface{} `json:"data"`    // 实际数据
}

type setPropertyRequest struct {
	RequestId   string
	Identifiers []string // 如设备指向的域名、端口必须同时写入，因此这里必须是一个数组
	Params      map[string]any
}

/*
写报警参数定值命令
字节	1字节	1字节	1字节	1字节	6字节	1字节	1字节	 2字节	 2字节	N字节	1字节	1字节
68H	长度L	长度L	68H		03H		终端地址	CBH		参数个数1  0006H	 0000	数据		CS		16H

数据
2字节	1字节	2字节	 1字节	1字节	K字节
0000	0000	信息地址   002d	数据长度	值
*/

func setProperty(sn, lineNo string, payload []byte, client mqtt.Client) {

	// 判断sn是否在连接过当前app
	addr, ok := snAndRemoteAddrMap.Load(sn)
	if !ok {
		// 设备未在当前应用上线，忽略请求
		return
	}

	var request setPropertyRequest

	if err := json.Unmarshal(payload, &request); err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	log.Info().Str("sn", sn).Interface("request", request).Msg("setProperty")

	var e error

	defer func() {
		resp := CommonResponse{RequestId: request.RequestId, Success: true, Message: "OK"}
		if e != nil {
			log.Debug().Err(e).Msg("")
			resp.Success = false
			resp.Message = e.Error()
		}

		buf, err := json.Marshal(resp)
		if err != nil {
			log.Error().Err(err).Msg("")
			return
		}

		if token := client.Publish(request.RequestId, 2, false, buf); token.Wait() && token.Error() != nil {
			log.Error().Err(token.Error()).Msg("")
		}
	}()

	targetRwRegisters := FindRegisters(request.Identifiers)

	if len(targetRwRegisters) == 0 {
		e = errors.New("匹配不到对应的寄存器")
		return
	}

	if len(targetRwRegisters) > 1 {
		e = errors.New("请勿传入多个属性！")
		return
	}

	//获取终端地址
	address, err := util.SetByteSN(lineNo)

	targetRwRegister := targetRwRegisters[0]
	data := make([]byte, targetRwRegisters[0].Len())

	if targetRwRegister.Name() == Status.Name() {
		//写入遥控固定参数和参数信息地址
		err := targetRwRegister.Encode(request.Params, data)
		cfg := writeStatusCfg
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, targetRwRegister.Addr())
		cfg = append(cfg, b...)

		c := &jg.Frame{
			Function: 0x2d,
			Address:  address,
			Cfg:      cfg,
			Data:     data,
		}
		cmd := c.Bytes()

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		out, err := server.DownloadCommand(ctx, addr, cmd)
		if err != nil {
			e = err
			return
		}
		switchData, err := jg.NewFrame(out)
		if err != nil {
			e = err
			return
		}
		res := switchData.Data[len(switchData.Data)-1]
		if res == 0x00 {
			return
		}
		e = fmt.Errorf("执行失败！错误代码：%v", res)
		return
	}

	err = targetRwRegisters[0].Encode(request.Params, data)
	if err != nil {
		e = err
		return
	}

	cfg := writeCfg
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, targetRwRegisters[0].Addr())
	cfg = append(cfg, b...)
	cfg = append(cfg, 0x2d)
	cfg = append(cfg, byte(targetRwRegisters[0].Len()))
	c := &jg.Frame{
		Function: 0xcb,
		Address:  address,
		Cfg:      cfg,
		Data:     data,
	}
	cmd := c.Bytes()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	out, err := server.DownloadCommand(ctx, addr, cmd)
	if err != nil {
		e = err
		return
	}
	settingData, err := jg.NewFrame(out)
	if err != nil {
		e = err
		return
	}
	res := settingData.Data[len(settingData.Data)-1]
	if res == 0x00 {
		go func() {
			if IsAlarmSetting(targetRwRegisters) {
				lineModel, err := util.GetLineModel(lineNo)
				if err != nil {
					log.Error().Err(err).Msg("get lineModel failed")
					return
				}
				// 推送设备属性
				if err := mq.Publish(config.ProjectName()+"/"+sn+"/"+lineModel+"/"+lineNo+"/property", 1, false, request.Params); err != nil {
					log.Error().Err(err).Msg("")
				}
			}
		}()
		return
	}
	if settingData.Function == UploadLiveDataCmd {
		e = fmt.Errorf("错误的设备响应，请稍后重试,data:% x", out)
		return
	}
	e = fmt.Errorf("执行失败！错误代码：%v", res)
}

type invokeServiceRequest struct {
	RequestId  string
	Identifier string
	Params     map[string]any
}

/*
写入开关分合闸控制 Status
1字节	1字节	1字节	1字节	1字节	6字节	1字节	1字节	2字节	2字节	2字节	1字节	1字节	1字节
68H		长度L	长度L	68H		03H		终端地址	2DH		81H		0006H	0000	信息地址	数据		CS		16H

数据
0：执行“分”
1：执行“合”
*/

func invokeService(sn string, payload []byte, client mqtt.Client) {

	// 判断sn是否在连接过当前app
	addr, ok := snAndRemoteAddrMap.Load(sn)
	if !ok {
		// 设备未在当前应用上线，忽略请求
		return
	}

	var request invokeServiceRequest

	if err := json.Unmarshal(payload, &request); err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	log.Info().Str("sn", sn).Interface("request", request).Msg("invokeService")

	var e error
	var data interface{}

	defer func() {
		var resp = CommonResponse{RequestId: request.RequestId, Success: true, Message: "OK", Data: data}

		if e != nil {
			resp.Success = false
			resp.Message = e.Error()
		}

		result, err := json.Marshal(resp)
		if err != nil {
			log.Error().Err(err).Msg("")
			return
		}

		if token := client.Publish(request.RequestId, 2, false, result); token.Wait() && token.Error() != nil {
			log.Error().Err(token.Error()).Msg("")
		}
	}()

	if request.Identifier == "" {
		e = errors.New("identifier不能为空")
		return
	}

	// 主动断开连接
	if request.Identifier == "Disconnect" {
		e = server.CloseConn(addr)
		return
	}

	service := findService(request.Identifier)

	if service == nil {
		e = errors.New("找不到服务")
		return
	}

	cmd, err := service.Encode(request.Params)
	if err != nil {
		e = err
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	out, err := server.DownloadCommand(ctx, addr, cmd)
	if err != nil {
		e = err
		return
	}
	result, err := jg.NewFrame(out)
	if err != nil {
		e = err
		return
	}
	res := result.Data[len(result.Data)-1]
	if res == 0x00 {
		return
	}
	if result.Function == UploadLiveDataCmd {
		e = fmt.Errorf("错误的设备响应，请稍后重试,data:% x", out)
		return
	}
	e = fmt.Errorf("执行失败！错误代码：%v", res)
	return
}

// FindRegisters 查找寄存器
// 目前对外开放的寄存器都可读写
func FindRegisters(identifiers []string) (targetRWRegisters jg.ReadAndWritableRegisters) {
	for _, identifier := range identifiers {
		for _, rw := range RWPacket {
			name := rw.Name()
			if name == identifier {
				targetRWRegisters = append(targetRWRegisters, rw)
			}
		}
	}
	return
}

func findService(identifier string) *Service {
	for _, s := range services {
		if s.Identifier == identifier {
			return s
		}
	}

	return nil
}

func getHost(sn string, client mqtt.Client, payload []byte) {
	_, ok := snAndRemoteAddrMap.Load(sn)
	if ok {
		var request getHostRequest
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Error().Err(err).Msg("")
			return
		}

		log.Info().Str("sn", sn).Interface("request", request).Msg("getHost")

		buf, _ := json.Marshal(&CommonResponse{
			RequestId: request.RequestId,
			Success:   true,
			Message:   "OK",
			Data:      config.Host(),
		})

		if token := client.Publish(request.RequestId, 1, false, buf); token.Wait() && token.Error() != nil {
			log.Error().Err(token.Error()).Msg("")
		}
	}
}

// IsAlarmSetting 是否是报警设定值
func IsAlarmSetting(rs jg.ReadAndWritableRegisters) bool {
	for _, r := range rs {
		for _, a := range AlarmSettingPacket {
			if r.Name() == a.Name() {
				return true
			}
		}
	}
	return false
}
