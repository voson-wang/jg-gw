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
68 10 10 68 80 00 00 00 00 00 00（固定格式） 8b（注册帧控制字）（收到的这里是8d的是心跳包） 18 21 06 23 00 96 （集中器ID）71 00（数据序号，上电初始值为0） 74（cs校验和） 16
*/

// Frame Address: 开关终端地址地址	Function：命令代码	Cfg：指令对应固定参数(主要用于生成下发指令)	Data：数据
type Frame struct {
	Address  []byte
	Function uint8
	Cfg      []byte
	Data     []byte
}

// NewFrame converts a packet to a JG frame.
func NewFrame(packet []byte) (*Frame, error) {

	pLen := len(packet)

	if pLen < 7 {
		return nil, fmt.Errorf("jg: frame error: packet lenght expect >=7, got %v", pLen)
	}

	csExpect := packet[pLen-2]
	csCalc := crcModbus(packet[4 : pLen-2])

	if csExpect != csCalc {
		return nil, fmt.Errorf("jg: frame error: CheckSum (expected 0x%x, got 0x%x)", csExpect, csCalc)
	}

	frame := &Frame{
		Address:  packet[5:11],
		Function: packet[11],
		Data:     packet[12 : pLen-2],
	}

	return frame, nil
}

func (frame *Frame) Copy() *Frame {
	f := *frame
	return &f
}

// Bytes returns the MODBUS byte stream based on the AcrelFrame fields
func (frame *Frame) Bytes() []byte {
	b := make([]byte, 5)

	// 添加定界符
	b[0] = 0x68
	b[3] = 0x68
	b[4] = 0x03

	b = append(b, frame.Address...)

	b = append(b, frame.Function)

	if frame.Cfg != nil {
		b = append(b, frame.Cfg...)
	}

	b = append(b, frame.Data...)

	// Calculate the CheckSum.

	cs := crcModbus(b[4:])

	b = append(b, cs)
	b = append(b, 0x16)
	b[1] = byte(len(b) - 6)
	b[2] = b[1]
	return b
}

// GetFunction returns the Modbus function code.
func (frame *Frame) GetFunction() uint8 {
	return frame.Function
}

// GetData returns the AcrelFrame Data byte field.
func (frame *Frame) GetData() []byte {
	return frame.Data
}

// SetData sets the AcrelFrame Data byte field and updates the frame length
// accordingly.
func (frame *Frame) SetData(data []byte) {
	frame.Data = data
}
