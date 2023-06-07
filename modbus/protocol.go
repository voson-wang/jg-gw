// Package modbus
// 协议文件（2个）：如下
// KS用电终端通讯规约 2020-7-7
// KSV3扩展规约 2021-8-4
package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
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

	ID [6]byte // 集中器ID（网关ID）、微端通讯地址

	Login struct {
		ID ID
	}

	HeartBeat struct {
		ID      ID
		NodeIDs []ID
	}
)

// 终端、集中器控制
// 回复主站
const (
	DeviceCtrl80 Ctrl = 0x80 // 注册、掉电、心跳
	DeviceCtrl83 Ctrl = 0x83 // 故障
	DeviceCtrl88 Ctrl = 0x88 // 遥信
)

// 主站控制
// 发送给终端
const (
	ServerCtrlA Ctrl = 0x0A // 遥信、遥测
	ServerCtrl3 Ctrl = 0x03 // 故障回复确认
)

// 命令码
const (
	FaultFun     Function = 0x2A
	TeleFun      Function = 0x64 // 遥信、遥测
	RegisterFun  Function = 0x8B
	PowerDownFun Function = 0x8C
	HeartBeatFun Function = 0x8D
)

// Frame.Data的数据头
// 遥信、遥测用的是相同的控制和命名码，所不同的是在Header
var (
	FaultHeader           = [5]byte{0x00, 0x03, 0x00, 0x00, 0x00}
	FaultAckHeader        = [5]byte{0x00, 0x03, 0x01, 0x00, 0x00}
	TelemeteringHeader    = [8]byte{0x80, 0x06, 0x00, 0x00, 0x00, 0x01, 0x00, 0x20}
	TelemeteringAckHeader = [8]byte{0x80, 0x07, 0x00, 0x00, 0x00, 0x01, 0x00, 0x20}
	TeleindicationHeader  = [8]byte{0x80, 0x06, 0x00, 0x00, 0x00, 0x01, 0x40, 0x20}
	// TeleindicationAckHeader 规约遥测数据头和实际不相符，实际的数据头最后一个报文是0x00，规约上是0x20
	TeleindicationAckHeader = [8]byte{0x80, 0x07, 0x00, 0x00, 0x00, 0x01, 0x40, 0x00}
)

func (i ID) String() string {
	var s string
	for _, v := range i {
		s += fmt.Sprintf("%02X", v)
	}
	return s
}

func NodesString(is []ID) string {
	var node []string
	for _, id := range is {
		node = append(node, id.String())
	}
	return strings.Join(node, ",")
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
func (f *Frame) NewLogin() (*Login, error) {

	if f.Ctrl != DeviceCtrl80 {
		return nil, fmt.Errorf("frame ctrl error: ctrl expect 0x%X,got 0x%X", DeviceCtrl80, f.Ctrl)
	}

	data := f.Data

	if len(data) < 8 {
		return nil, fmt.Errorf("frame data error: data expect len >8,got %v", len(data))
	}

	l := &Login{ID: [6]byte(data[:6])}
	return l, nil
}

// NewHeartBeat
// 心跳
func (f *Frame) NewHeartBeat() (*HeartBeat, error) {
	data := f.Data

	if len(data) < 8 {
		return nil, fmt.Errorf("frame data error: data expect len >8,got %v", len(f.Data))
	}

	h := &HeartBeat{
		ID: [6]byte(data[:6]),
	}

	var nodeID []ID

	for i := 6; i+6 <= len(data)-1; i = i + 6 {
		nodeID = append(nodeID, ID(data[i:i+6]))
	}

	h.NodeIDs = nodeID
	return h, nil
}

// NewFault
// 终端回复故障或上报故障
// 规约 4.6.2
func (f *Frame) NewFault() (*Fault, error) {

	if f.Ctrl != DeviceCtrl83 {
		return nil, fmt.Errorf("frame ctrl error: ctrl expect 0x%X,got 0x%X", DeviceCtrl83, f.Ctrl)
	}

	data := f.Data

	if len(data) < 21 {
		return nil, fmt.Errorf("frame data error: data expect len >= 21,got %v", len(data))
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

	for i := 20; i+4 <= len(data)-1; i = i + 4 {
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
func (f *Frame) NewFaultAck(fault *Fault) *Frame {
	ackFrame := f.Copy()
	f.Ctrl = ServerCtrl3

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

// NewTelemetering
// 创建一个主站发送的遥信数据
// 规约 4.1.1
func NewTelemetering(address ID) *Frame {
	f := &Frame{
		Ctrl:     ServerCtrlA,
		Address:  address,
		Function: TeleFun,
	}

	f.Data = make([]byte, 8)
	f.Data[0] = TelemeteringHeader[0]
	f.Data[1] = TelemeteringHeader[1]
	f.Data[2] = TelemeteringHeader[2]
	f.Data[3] = TelemeteringHeader[3]
	f.Data[4] = TelemeteringHeader[4]
	f.Data[5] = TelemeteringHeader[5]
	f.Data[6] = TelemeteringHeader[6]
	f.Data[7] = TelemeteringHeader[7]

	return f
}

// NewTeleindication
// 创建一个主站发送的遥测数据
// 规约 4.2.1
func NewTeleindication(address ID) *Frame {
	f := &Frame{
		Ctrl:     ServerCtrlA,
		Address:  address,
		Function: TeleFun,
	}

	f.Data = make([]byte, 8)
	f.Data[0] = TeleindicationHeader[0]
	f.Data[1] = TeleindicationHeader[1]
	f.Data[2] = TeleindicationHeader[2]
	f.Data[3] = TeleindicationHeader[3]
	f.Data[4] = TeleindicationHeader[4]
	f.Data[5] = TeleindicationHeader[5]
	f.Data[6] = TeleindicationHeader[6]
	f.Data[7] = TeleindicationHeader[7]

	return f
}

// NewTelemeteringAck
// 终端回复的遥信数据
// 规约 4.1.2
func (f *Frame) NewTelemeteringAck() (map[string]any, error) {

	if f.Ctrl != DeviceCtrl88 {
		return nil, fmt.Errorf("frame ctrl error: ctrl expect 0x%X,got 0x%X", DeviceCtrl88, f.Ctrl)
	}

	if f.Function != TeleFun {
		return nil, fmt.Errorf("frame function error: function expect 0x%X,got 0x%X", TeleFun, f.Function)
	}

	data := f.Data

	if len(data) < 25 {
		return nil, fmt.Errorf("frame data error: data expect len >= 25,got %v", len(data))
	}

	if data[0] != TelemeteringAckHeader[0] ||
		data[1] != TelemeteringAckHeader[1] ||
		data[2] != TelemeteringAckHeader[2] ||
		data[3] != TelemeteringAckHeader[3] ||
		data[4] != TelemeteringAckHeader[4] ||
		data[5] != TelemeteringAckHeader[5] ||
		data[6] != TelemeteringAckHeader[6] ||
		data[7] != TelemeteringAckHeader[7] {
		return nil, errors.New("frame data error:  packet format error")
	}

	actualData := data[8:]

	return map[string]any{
		"Switch":         actualData[0],
		"LeakageProtect": actualData[25],
	}, nil
}

// AnalogQuantity 模拟量
// 参数地址分配（2020）_MCB_2021.08.132
type AnalogQuantity struct {
	Num         int // 序号，从1开始
	Name        string
	Coefficient float64 // 系数
}

var analogQuantities = []*AnalogQuantity{
	{
		Num:         4,
		Name:        "Ua",
		Coefficient: 0.1,
	},
	{
		Num:         5,
		Name:        "Ub",
		Coefficient: 0.1,
	},
	{
		Num:         6,
		Name:        "Uc",
		Coefficient: 0.1,
	},
	{
		Num:         8,
		Name:        "Ia",
		Coefficient: 0.01,
	},
	{
		Num:         9,
		Name:        "Ib",
		Coefficient: 0.01,
	},
	{
		Num:         10,
		Name:        "Ic",
		Coefficient: 0.01,
	},
	{
		Num:         11,
		Name:        "Leakage",
		Coefficient: 1,
	},
	{
		Num:         12,
		Name:        "P",
		Coefficient: 0.01,
	},
	{
		Num:         14,
		Name:        "PF",
		Coefficient: 0.01,
	},
	{
		Num:         15,
		Name:        "EPI", // 正向有功总和、总电量L
		Coefficient: 0.01,
	},
	{
		Num:         26,
		Name:        "Ta",
		Coefficient: 1,
	},
	{
		Num:         27,
		Name:        "Tb",
		Coefficient: 1,
	},
	{
		Num:         28,
		Name:        "Tc",
		Coefficient: 1,
	},
	{
		Num:         29,
		Name:        "TN",
		Coefficient: 1,
	},
}

// NewTeleindicationAck
// 终端回复的遥测数据
// 规约 4.2.2
func (f *Frame) NewTeleindicationAck() (map[string]any, error) {
	if f.Ctrl != DeviceCtrl88 {
		return nil, fmt.Errorf("frame ctrl error: ctrl expect 0x%X,got 0x%X", DeviceCtrl88, f.Ctrl)
	}

	if f.Function != TeleFun {
		return nil, fmt.Errorf("frame function error: function expect 0x%X,got 0x%X", TeleFun, f.Function)
	}

	data := f.Data

	if data[0] != TeleindicationAckHeader[0] ||
		data[1] != TeleindicationAckHeader[1] ||
		data[2] != TeleindicationAckHeader[2] ||
		data[3] != TeleindicationAckHeader[3] ||
		data[4] != TeleindicationAckHeader[4] ||
		data[5] != TeleindicationAckHeader[5] ||
		data[6] != TeleindicationAckHeader[6] ||
		data[7] != TeleindicationAckHeader[7] {
		return nil, errors.New("frame data error:  packet format error")
	}

	actualData := data[8:]

	m := make(map[string]any)

	for _, a := range analogQuantities {
		m[a.Name] = decimal.NewFromInt(int64(binary.LittleEndian.Uint16(actualData[(a.Num-1)*2 : a.Num*2]))).Mul(decimal.NewFromFloat(a.Coefficient))
	}

	return m, nil
}
