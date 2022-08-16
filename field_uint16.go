package main

import (
	"encoding/binary"
	"fmt"
	"github.com/rs/zerolog/log"
	"math"
	"reflect"
)

type Uint16FieldDecode func(data uint16) (any, error)

type Uint16Field struct {
	name  string
	start int
	// order 大小端
	// 安科瑞协议默认小端，但也有可能大端
	order  binary.ByteOrder
	decode Uint16FieldDecode
}

func NewUint16Field(name string, start int) *Uint16Field {
	return &Uint16Field{name: name, start: start, order: binary.LittleEndian}
}

func NewUint16FieldWithDecode(name string, start int, decode Uint16FieldDecode) *Uint16Field {
	return &Uint16Field{name: name, start: start, decode: decode, order: binary.LittleEndian}
}

func (u *Uint16Field) Name() string {
	return u.name
}

func (u *Uint16Field) Start() int {
	return u.start
}

func (u *Uint16Field) Len() int {
	return 2
}

func (u *Uint16Field) Decode(data []byte, values map[string]any) {
	u16 := u.order.Uint16(data)

	if u.decode != nil {
		value, err := u.decode(u16)
		if err != nil {
			log.Error().Err(err).Msg(DecodeErr)
			return
		}
		values[u.name] = value
		return
	}
	values[u.name] = u16
}

type Uint16FieldEncode func(value float64, params map[string]any) (uint16, error)

type Uint16WoField struct {
	name  string
	start int
	// order 大小端
	// 安科瑞协议默认小端，但也有可能大端
	order  binary.ByteOrder
	decode Uint16FieldDecode
	encode Uint16FieldEncode
}

func NewUint16WoField(name string, start int) *Uint16WoField {
	return &Uint16WoField{name: name, start: start, order: binary.LittleEndian}
}

func (u *Uint16WoField) Name() string {
	return u.name
}

func (u *Uint16WoField) Start() int {
	return u.start
}

func (u *Uint16WoField) Len() int {
	return 2
}

func (u *Uint16WoField) Decode(data []byte, values map[string]any) {
	u16 := u.order.Uint16(data)

	if u.decode != nil {
		value, err := u.decode(u16)
		if err != nil {
			log.Error().Err(err).Msg(DecodeErr)
			return
		}
		values[u.name] = value
		return
	}
	values[u.name] = u16
}

func (u *Uint16WoField) Encode(params map[string]interface{}, dst []byte) error {
	value, ok := params[u.name]
	if !ok {
		return fmt.Errorf("参数 %v 缺失", u.name)
	}

	// 从消息队列中获得的数字
	// 类型经过json序列化后会被转换为float64
	f64, ok := value.(float64)
	if !ok {
		return fmt.Errorf("参数 %v 类型错误，期望：float64，实际：%v", u.name, reflect.TypeOf(value))
	}
	if f64 != math.Trunc(f64) {
		return fmt.Errorf("参数 %v 类型错误，期望：整型，实际：%v", u.name, value)
	}

	var u16 uint16
	var err error

	if u.encode == nil {
		u16 = uint16(f64)
	} else {
		u16, err = u.encode(f64, params)
		if err != nil {
			return err
		}
	}

	u.order.PutUint16(dst, u16)
	return nil
}
