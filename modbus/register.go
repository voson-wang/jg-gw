package modbus

import (
	"encoding/binary"
	"fmt"
)

type Register interface {
	Name() string

	// Start 返回在字节流中的起始地址
	Start() int

	// Len 返回字节长度
	Len() int

	// Addr 返回参数信息地址
	Addr() uint16
}

type Readable interface {
	// Decode 为了适应单个字段，两个字节代表2个参数的情况和
	// 单个字段，每个比特代表不同参数的情况
	// 所以使用了result去获取decode结果
	Decode(data []byte, results map[string]any)
}

type Writable interface {
	// Encode 为了适应单个字段，两个字节代表2个参数的情况和
	// 单个字段，每个比特代表不同参数的情况
	// 所以使用map来作为输入值，让字段自行取用输入值
	Encode(params map[string]any, dst []byte) error
}

type ReadRegister interface {
	Register
	Readable
}

type ReadAndWritableRegister interface {
	Register
	Writable
	Readable
}

type ReadAndWritableRegisters []ReadAndWritableRegister

func (p ReadAndWritableRegisters) Len() int {
	var r int
	for _, w := range p {
		r = r + w.Len()
	}
	return r
}

// Encode 写入不支持跳过字段
func (p ReadAndWritableRegisters) Encode(params map[string]any) ([]byte, error) {
	result := make([]byte, p.Len())
	for _, register := range p {
		start := register.Start()
		end := start + register.Len()
		if err := register.Encode(params, result[start:end]); err != nil {
			return nil, fmt.Errorf("decode error: %v; register: %v, start: %v,end: %v", err, register.Name(), start, end)
		}
	}
	return result, nil
}

// Decode 不支持跳过预留字段
func (p ReadAndWritableRegisters) Decode(data []byte) map[string]any {
	m := make(map[string]any)
	for _, register := range p {
		start := register.Start()
		end := start + register.Len()
		register.Decode(data[start:end], m)
	}
	return m
}

func (p ReadAndWritableRegisters) NewReadRegisters(function uint8, cfg, address [6]byte) ([]byte, error) {
	result := make([]byte, p.Len())
	for _, register := range p {
		start := register.Start()
		end := start + register.Len()
		binary.LittleEndian.PutUint16(result[start:end], register.Addr())
	}
	f := Frame{Function: function, Address: address}
	f.SetData(result)
	return f.Bytes(), nil
}

func (p ReadAndWritableRegisters) Bytes(function uint8, cfg, address [6]byte, params map[string]any) ([]byte, error) {
	data, err := p.Encode(params)
	if err != nil {
		return nil, err
	}
	f := Frame{Function: function, Address: address}
	f.SetData(data)
	return f.Bytes(), nil
}
