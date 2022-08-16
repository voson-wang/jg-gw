package main

import (
	"fmt"
	"math"
	"reflect"
)

// ByteRwRegister 每个字节代表一个参数
// 寄存器默认为1，参数个数<=2
type ByteRwRegister struct {
	name    string
	start   int
	address uint16
}

func NewByteRwRegister(name string, start int, address uint16) *ByteRwRegister {
	return &ByteRwRegister{name: name, start: start, address: address}
}

func (b *ByteRwRegister) Start() int {
	return b.start
}

func (b *ByteRwRegister) Len() int {
	return 1
}

func (b *ByteRwRegister) Name() string {
	return b.name
}

func (b *ByteRwRegister) Addr() uint16 {
	return b.address
}

func (b *ByteRwRegister) Decode(data []byte, values map[string]any) {
	value := data[0]
	values[b.name] = value
}

func (b *ByteRwRegister) Encode(params map[string]any, dst []byte) error {

	value, ok := params[b.name]
	if !ok {
		return fmt.Errorf("参数 %v 缺失", b.name)
	}

	f64, ok := value.(float64)
	if !ok {
		return fmt.Errorf("参数 %v 类型错误，期望：float64，实际：%v", b.name, reflect.TypeOf(value))
	}

	if f64 != math.Trunc(f64) {
		return fmt.Errorf("参数 %v 类型错误，期望：整型，实际：%v", b.name, f64)
	}

	if f64 < 0 || f64 > 255 {
		return fmt.Errorf("参数 %v 值越界，范围：0~255，实际：%v", b.name, f64)
	}

	dst[0] = byte(f64)

	return nil
}
