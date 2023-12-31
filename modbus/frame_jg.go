package modbus

import (
	"fmt"
)

/*
京硅微断modbus协议
使用可变帧长。数据传输顺序为低位在前，高位在后；低字节在前，高字节在后。

1字节	1字节	1字节	1字节	1字节	6字节	1字节	N字节	1字节	1字节
68H		长度L	长度L	68H		控制		终端地址	命令码	用户数据	校验CS	16H

长度L：从控制域到用户数据的字节总长度，第2个报文长度L与第一个报文长度L相同。
控制字：1个字节，定义数据传送方向及数据帧种类。
终端地址：6个字节，选址范围为000000000001H～FFFFFFFFFFFFH，其中FFFFFFFFFFFFH为广播地址，000000000000H为无效地址。
校验和CS：1个字节，是控制字、终端地址、命令码、用户数据的字节的八位位组算术和，不考虑溢出位，即：CS＝（控制字+终端地址+命令码+用户数据）MOD 256。

注册包、心跳包
注册包可能会出现1～3次，随后可能时故障主动上传和遥信主动上传，故障需要主站回复确认
68 10 10 68 80 00 00 00 00 00 00（固定格式） 8b（注册帧控制字）（收到的这里是8d的是心跳包） 18 21 06 23 00 96 （集中器ID）71 00（数据序号，上电初始值为0） 74（cs校验和） 16
*/

type Ctrl byte

type Function byte

type Frame struct {
	Ctrl     Ctrl     // 控制
	ID       ID       // 终端地址，开关（节点）编号
	Function Function // 命令码
	Data     []byte   // 用户数据
}

const startFlag byte = 0x68 // 起始字符
const endFlag byte = 0x16   // 终止符号

// NewFrame converts a packet to a JG frame.
func NewFrame(packet []byte) (*Frame, error) {

	pLen := len(packet)

	if pLen < 14 {
		return nil, fmt.Errorf("frame error: packet lenght expect >=14, got %v", pLen)
	}

	if packet[0] != startFlag || packet[3] != startFlag || packet[pLen-1] != endFlag {
		return nil, fmt.Errorf("frame error: packet format error")
	}

	// 获取长度L
	l := len(packet[4 : pLen-2])

	if byte(l) != packet[1] || packet[1] != packet[2] {
		return nil, fmt.Errorf("frame error: packet lenght error")
	}

	// 校验和
	csExpect := packet[pLen-2]
	csCalc := crcModbus(packet[4 : pLen-2])

	if csExpect != csCalc {
		return nil, fmt.Errorf("frame error: CheckSum (expected 0x%X, got 0x%X)", csExpect, csCalc)
	}

	frame := &Frame{
		Ctrl:     Ctrl(packet[4]),
		ID:       [6]byte(packet[5:11]),
		Function: Function(packet[11]),
		Data:     packet[12 : pLen-2],
	}

	return frame, nil
}

func (f *Frame) Copy() *Frame {
	frame := *f
	return &frame
}

// Bytes returns the MODBUS byte stream based on the Frame fields
func (f *Frame) Bytes() []byte {
	b := make([]byte, 12)

	// 添加定界符
	b[0] = startFlag
	b[3] = startFlag
	b[4] = byte(f.Ctrl)
	b[5] = f.ID[0]
	b[6] = f.ID[1]
	b[7] = f.ID[2]
	b[8] = f.ID[3]
	b[9] = f.ID[4]
	b[10] = f.ID[5]
	b[11] = byte(f.Function)

	b = append(b, f.Data...)

	// Calculate the CheckSum.
	cs := crcModbus(b[4:])

	b = append(b, cs)
	b = append(b, endFlag)
	b[1] = byte(len(b) - 6)
	b[2] = b[1]
	return b
}

// GetFunction returns the Modbus function code.
func (f *Frame) GetFunction() Function {
	return f.Function
}

// GetData returns the Frame Data byte field.
func (f *Frame) GetData() []byte {
	return f.Data
}

// SetData sets the Frame Data byte field and updates the frame length accordingly.
func (f *Frame) SetData(data []byte) {
	f.Data = data
}
