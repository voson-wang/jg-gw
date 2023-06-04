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

// Frame Address: 开关终端地址地址	 Function：命令代码  Data：数据
type Frame struct {
	Size     byte
	Ctrl     byte
	Address  [6]byte
	Function byte
	Data     []byte
	CS       byte
}

const flag byte = 0x68
const ender byte = 0x16

// NewFrame converts a packet to a JG frame.
func NewFrame(packet []byte) (*Frame, error) {

	pLen := len(packet)

	if pLen < 14 {
		return nil, fmt.Errorf("[ModBus]: frame error: packet lenght expect >=14, got %v", pLen)
	}

	if packet[0] != flag || packet[3] != flag || packet[pLen-1] != ender {
		return nil, fmt.Errorf("[ModBus]: frame error: packet format error")
	}

	// 获取长度L
	l := len(packet[4 : pLen-2])

	if byte(l) != packet[1] {
		return nil, fmt.Errorf("[ModBus]: frame error: packet lenght error")
	}

	// 校验和
	csExpect := packet[pLen-2]
	csCalc := crcModbus(packet[4 : pLen-2])

	if csExpect != csCalc {
		return nil, fmt.Errorf("[ModBus]: frame error: CheckSum (expected 0x%x, got 0x%x)", csExpect, csCalc)
	}

	frame := &Frame{
		Size:     packet[1],
		Ctrl:     packet[4],
		Address:  [6]byte(packet[5:11]),
		Function: packet[11],
		Data:     packet[12 : pLen-2],
		CS:       csExpect,
	}

	return frame, nil
}

func (frame *Frame) Copy() *Frame {
	f := *frame
	return &f
}

// Bytes returns the MODBUS byte stream based on the Frame fields
func (frame *Frame) Bytes() []byte {
	b := make([]byte, 11)

	// 添加定界符
	b[0] = flag
	b[3] = flag
	b[4] = frame.Ctrl
	b[5] = frame.Address[0]
	b[6] = frame.Address[1]
	b[7] = frame.Address[2]
	b[8] = frame.Address[3]
	b[9] = frame.Address[4]
	b[10] = frame.Address[5]
	b[11] = frame.Function

	b = append(b, frame.Data...)

	// Calculate the CheckSum.
	cs := crcModbus(b[4:])

	b = append(b, cs)
	b = append(b, ender)
	b[1] = byte(len(b) - 6)
	b[2] = b[1]
	return b
}

// GetFunction returns the Modbus function code.
func (frame *Frame) GetFunction() uint8 {
	return frame.Function
}

// GetData returns the Frame Data byte field.
func (frame *Frame) GetData() []byte {
	return frame.Data
}

// SetData sets the Frame Data byte field and updates the frame length accordingly.
func (frame *Frame) SetData(data []byte) {
	frame.Data = data
}
