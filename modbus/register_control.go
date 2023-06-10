package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// ControlRegister 控制量 单字节只写寄存器
type ControlRegister struct {
	name    string
	address uint16
}

func (r *ControlRegister) NewWriteFrame(id ID, val []byte) *Frame {
	data := make([]byte, 5)
	data[0] = TelecontrolHeader[0]
	data[1] = TelecontrolHeader[1]
	data[2] = TelecontrolHeader[2]
	data[3] = TelecontrolHeader[3]
	data[4] = TelecontrolHeader[4]
	address := make([]byte, 2)
	binary.LittleEndian.PutUint16(address, r.address)
	data = append(data, address...)
	data = append(data, val...)
	return &Frame{
		Ctrl:     ServerCtrl3,
		ID:       id,
		Function: Telecontrol,
		Data:     data,
	}
}

func (r *ControlRegister) Name() string {
	return r.name
}

func (r *ControlRegister) Address() uint16 {
	return r.address
}

func (r *ControlRegister) Len() uint8 {
	return 1
}

func (r *ControlRegister) Encode(params map[string]any) ([]byte, error) {
	value, ok := params[r.name]
	if !ok {
		return nil, fmt.Errorf("参数 %v 缺失", r.name)
	}

	param, err := ConvertToUint8(value)
	if err != nil {
		return nil, err
	}

	return []byte{param}, nil
}

func (r *ControlRegister) ParserWriteResp(frame *Frame) (bool, error) {
	if frame.Ctrl != DeviceCtrl80 {
		return false, fmt.Errorf("expect ctrl 0x%X, got 0x%X", DeviceCtrl80, frame.Ctrl)
	}

	if frame.Function != Telecontrol {
		return false, fmt.Errorf("expect function 0x%X, got 0x%X", Telecontrol, frame.Function)
	}

	data := frame.Data

	if len(data) != 8 {
		return false, fmt.Errorf("frame error: packet lenght expect 8, got %v", len(data))
	}

	if data[0] != TelecontrolAckHeader[0] ||
		data[1] != TelecontrolAckHeader[1] ||
		data[2] != TelecontrolAckHeader[2] ||
		data[3] != TelecontrolAckHeader[3] ||
		data[4] != TelecontrolAckHeader[4] {
		return false, errors.New("invalid telecontrol ack header")
	}

	address := binary.LittleEndian.Uint16(data[5:7])

	if address != r.address {
		return false, fmt.Errorf("expect address 0x%X, got 0x%X", r.address, address)
	}

	// 遥控成功
	if data[7] == 0x00 {
		return true, nil
	}

	// 失败
	return false, nil
}

func ConvertToUint8(val interface{}) (uint8, error) {
	switch v := val.(type) {
	case uint8:
		return v, nil
	case int:
		if v < 0 || v > math.MaxUint8 {
			return 0, fmt.Errorf("value out of range for uint8: %v", v)
		}
		return uint8(v), nil
	case float64:
		if v < 0 || v > math.MaxUint8 {
			return 0, fmt.Errorf("value out of range for uint8: %v", v)
		}
		return uint8(v), nil
	default:
		return 0, fmt.Errorf("unsupported type %T for uint8 conversion", val)
	}
}
