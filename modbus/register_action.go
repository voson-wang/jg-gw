package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// ActionRegister 动作参数寄存器 可读写
type ActionRegister struct {
	name    string
	address uint16
	len     uint8
	tag     byte
}

func (r *ActionRegister) ReadFrame(id ID) *Frame {
	data := make([]byte, 5)

	data[0] = 1 // 本程序默认只支持单个参数写入
	data[1] = MultiParamsHeader[0]
	data[2] = MultiParamsHeader[1]
	data[3] = MultiParamsHeader[2]
	data[4] = MultiParamsHeader[3]

	data = append(data, 0x00, 0x00) // 定值区间
	address := make([]byte, 2)
	binary.LittleEndian.PutUint16(address, r.address)
	data = append(data, address...) // 信息地址
	return &Frame{
		Ctrl:     ServerCtrl3,
		ID:       id,
		Function: MultiReadFun,
		Data:     data,
	}
}

func (r *ActionRegister) NewWriteFrame(id ID, val []byte) *Frame {
	data := make([]byte, 5)
	data[0] = 1 // 本程序默认只支持单个参数写入
	data[1] = MultiParamsHeader[0]
	data[2] = MultiParamsHeader[1]
	data[3] = MultiParamsHeader[2]
	data[4] = MultiParamsHeader[3]
	data = append(data, 0x00, 0x00) // 定值区间
	data = append(data, 0x01)       // 特征标识
	address := make([]byte, 2)
	binary.LittleEndian.PutUint16(address, r.address)
	data = append(data, address...) // 信息地址
	data = append(data, r.tag)      // Tag类型，见扩展规约 附件1：数据类型
	data = append(data, r.len)      // 数据长度
	data = append(data, val...)     // 值
	return &Frame{
		Ctrl:     ServerCtrl3,
		ID:       id,
		Function: MultiWriteFun,
		Data:     data,
	}
}

func (r *ActionRegister) Name() string {
	return r.name
}

func (r *ActionRegister) Address() uint16 {
	return r.address
}

func (r *ActionRegister) Len() uint8 {
	return r.len
}

func (r *ActionRegister) Decode(data []byte, results map[string]any) {
	results[r.name] = binary.LittleEndian.Uint16(data)
}

func (r *ActionRegister) Encode(params map[string]any) ([]byte, error) {
	value, ok := params[r.name]
	if !ok {
		return nil, fmt.Errorf("参数 %v 缺失", r.name)
	}

	param, err := ConvertToUint16(value)
	if err != nil {
		return nil, err
	}

	dst := make([]byte, 2)
	binary.LittleEndian.PutUint16(dst, param)
	return dst, nil
}

func (r *ActionRegister) ParserWriteResp(frame *Frame) (bool, error) {
	if frame.Ctrl != DeviceCtrl80 {
		return false, fmt.Errorf("expect ctrl 0x%X, got 0x%X", DeviceCtrl80, frame.Ctrl)
	}

	if frame.Function != MultiWriteFun {
		return false, fmt.Errorf("expect function 0x%X, got 0x%X", MultiWriteFun, frame.Function)
	}

	data := frame.Data

	if len(data) != 6 {
		return false, fmt.Errorf("frame error: packet lenght expect 6, got %v", len(data))
	}

	if data[0] != MultiWriteAckHeader[0] ||
		data[1] != MultiWriteAckHeader[1] ||
		data[2] != MultiWriteAckHeader[2] ||
		data[3] != MultiWriteAckHeader[3] ||
		data[4] != MultiWriteAckHeader[4] {
		return false, errors.New("invalid telecontrol ack header")
	}

	// 遥控成功
	if data[5] == 0x00 {
		return true, nil
	}

	// 失败
	return false, nil
}

func (r *ActionRegister) ParserReadResp(frame *Frame) (map[string]any, error) {
	if frame.Ctrl != DeviceCtrl83 {
		return nil, fmt.Errorf("expect ctrl 0x%X, got 0x%X", DeviceCtrl83, frame.Ctrl)
	}

	if frame.Function != MultiReadFun {
		return nil, fmt.Errorf("expect function 0x%X, got 0x%X", MultiReadFun, frame.Function)
	}

	data := frame.Data

	l := len(data)

	if l == 5 {
		return nil, fmt.Errorf("错误的信息地址或组号")
	}

	if len(data) < 12 {
		return nil, fmt.Errorf("frame error: packet lenght expect >= 6, got %v", len(data))
	}

	if data[0] != 0x01 ||
		data[1] != MultiReadAckHeader[0] ||
		data[2] != MultiReadAckHeader[1] ||
		data[3] != MultiReadAckHeader[2] ||
		data[4] != MultiReadAckHeader[3] ||
		data[5] != 0x00 ||
		data[6] != 0x00 ||
		// 特征标识
		data[7] != 0x00 {
		return nil, errors.New("invalid telecontrol ack header")
	}

	address := binary.LittleEndian.Uint16(data[8:10])
	if address != r.address {
		return nil, fmt.Errorf("expect address 0x%X, got 0x%X", r.address, address)
	}

	// Tag类型，见扩展规约 附件1：数据类型
	tag := data[10]
	if tag != r.tag {
		return nil, fmt.Errorf("expect tag 0x%X, got 0x%X", r.tag, tag)
	}

	dataLen := data[11]
	if dataLen != r.len {
		return nil, fmt.Errorf("expect len 0x%X, got 0x%X", r.len, dataLen)
	}

	result := make(map[string]any)

	r.Decode(data[12:12+l], result)

	return result, nil
}

func ConvertToUint16(val interface{}) (uint16, error) {
	switch v := val.(type) {
	case uint16:
		return v, nil
	case int:
		if v < 0 || v > math.MaxUint16 {
			return 0, fmt.Errorf("value out of range for uint16: %v", v)
		}
		return uint16(v), nil
	case float64:
		if v < 0 || v > math.MaxUint16 {
			return 0, fmt.Errorf("value out of range for uint16: %v", v)
		}
		return uint16(v), nil
	default:
		return 0, fmt.Errorf("unsupported type %T for uint16 conversion", val)
	}
}
