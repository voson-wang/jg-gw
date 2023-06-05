// Package modbus
// 协议文件（2个）：如下
// KS用电终端通讯规约 2020-7-7
// KSV3扩展规约 2021-8-4
package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

type (
	TimeMark [7]byte

	TeleindicationData struct {
		TeleindicationDit   [2]byte // 遥信点号
		TeleindicationValue [2]byte // 遥测值
	}

	// Fault
	// 故障
	// 4.6.2 数据头+故障数据
	Fault struct {
		TelemeteringNum      byte                 // 遥信个数
		TelemeteringType     byte                 // 遥信类型
		TelemeteringDit      [2]byte              // 遥信点号
		TelemeteringValue    byte                 // 遥信值
		TelemeteringTimeMark TimeMark             // 故障时标
		TeleindicationNum    [2]byte              // 遥测个数
		TeleindicationType   byte                 // 遥测类型
		TeleindicationData   []TeleindicationData // 遥测数据N
	}

	FaultAck struct {
		TelemeteringNum      byte     // 遥信个数
		TelemeteringType     byte     // 遥信类型
		TelemeteringDit      [2]byte  // 遥信点号
		TelemeteringValue    byte     // 遥信值
		TelemeteringTimeMark TimeMark // 故障时标
	}

	ID [6]byte // 集中器ID（网关ID）

	Login struct {
		ID ID
	}
)

// 终端、集中器控制
const (
	DeviceCtrl80 Ctrl = 0x80 // 注册、掉电、心跳

	DeviceCtrl83 Ctrl = 0x83 // 故障
)

// 主站控制
const (
	ServerCtrl Ctrl = 0x03 // 故障回复确认
)

// 命令码
const (
	FaultFun     Function = 0x2A
	RegisterFun  Function = 0x8B
	PowerDownFun Function = 0x8C
	HeartBeatFun Function = 0x8D
)

var (
	FaultHeader    = [5]byte{0x00, 0x03, 0x00, 0x00, 0x00}
	FaultAckHeader = [5]byte{0x00, 0x03, 0x01, 0x00, 0x00}
)

func (i *ID) String() string {
	var ss string
	for _, v := range i {
		ss += fmt.Sprintf("%02X", v)
	}
	return ss
}

// Time
// 将时标转换为可读的时间
// 规约 4.4.1 设置时钟发送
func (t *TimeMark) Time() time.Time {
	sec := binary.LittleEndian.Uint16([]byte{t[0], t[1]})
	return time.Date(int(t[6]), time.Month(t[5]), int(t[4]), int(t[3]), int(t[2]), int(sec), 0, time.Local)
}

// 遥信

// NewLogin
// 注册数据
// 扩展规约 4.1
func (frame *Frame) NewLogin() (*Login, error) {

	data := frame.Data

	if len(data) != 8 {
		return nil, fmt.Errorf("frame data error: data expect len 8,got %v", len(frame.Data))
	}

	l := &Login{ID: [6]byte(data[:6])}
	return l, nil
}

// NewFault
// 终端回复故障或上报故障
// 规约 4.6.2
func (frame *Frame) NewFault() (*Fault, error) {

	data := frame.Data

	if len(data) < 21 {
		return nil, fmt.Errorf("frame data error: data expect len >= 21,got %v", len(frame.Data))
	}

	if data[0] != FaultHeader[0] ||
		data[1] != FaultHeader[1] ||
		data[2] != FaultHeader[2] ||
		data[3] != FaultHeader[3] ||
		data[4] != FaultHeader[4] {
		return nil, errors.New("frame data error:  packet format error")
	}

	fault := &Fault{
		TelemeteringNum:      data[5],
		TelemeteringType:     data[6],
		TelemeteringDit:      [2]byte(data[7:9]),
		TelemeteringValue:    data[9],
		TelemeteringTimeMark: [7]byte(data[10:17]),
		TeleindicationNum:    [2]byte(data[17:19]),
		TeleindicationType:   data[19],
	}

	var teleindicationData []TeleindicationData

	for i := 20; i+4 <= len(data[20:])-1; i++ {
		teleindicationData = append(teleindicationData, TeleindicationData{
			TeleindicationDit:   [2]byte(data[i : i+2]),
			TeleindicationValue: [2]byte(data[i+2 : i+4]),
		})
	}

	fault.TeleindicationData = teleindicationData

	return fault, nil
}

// NewFaultAck
// 终端回复故障或上报故障
// 主站回复确认
// 规约 4.6.3
func (frame *Frame) NewFaultAck(fault *Fault) *Frame {
	ackFrame := frame.Copy()
	frame.Ctrl = ServerCtrl

	ackFrame.Data = make([]byte, 5)

	ackFrame.Data[0] = FaultAckHeader[0]
	ackFrame.Data[1] = FaultAckHeader[1]
	ackFrame.Data[2] = FaultAckHeader[2]
	ackFrame.Data[3] = FaultAckHeader[3]
	ackFrame.Data[4] = FaultAckHeader[4]

	ackFrame.Data = append(ackFrame.Data,
		fault.TelemeteringNum,
		fault.TelemeteringType,
		fault.TelemeteringDit[0],
		fault.TelemeteringDit[1],
		fault.TelemeteringValue,
		fault.TelemeteringTimeMark[0],
		fault.TelemeteringTimeMark[1],
		fault.TelemeteringTimeMark[2],
		fault.TelemeteringTimeMark[3],
		fault.TelemeteringTimeMark[4],
		fault.TelemeteringTimeMark[5],
		fault.TelemeteringTimeMark[6],
	)

	return ackFrame
}
