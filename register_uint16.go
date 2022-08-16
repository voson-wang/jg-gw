package main

import (
	"encoding/binary"
	"fmt"
	"github.com/rs/zerolog/log"
	"math"
	"reflect"
)

type Uint16RegisterValidate func(value uint16) error

type Uint16RegisterDecode func(data uint16, values map[string]any) (any, error)

type Uint16RegisterEncode func(value float64, params map[string]any) (uint16, error)

type Uint16RwRegister struct {
	name  string
	start int
	// order 默认小端
	order    binary.ByteOrder
	len      int
	address  uint16
	validate Uint16RegisterValidate
	decode   Uint16RegisterDecode
	encode   Uint16RegisterEncode
}

func NewUint16RwRegister(name string, start int, address uint16) *Uint16RwRegister {
	return &Uint16RwRegister{name: name, start: start, address: address, order: binary.LittleEndian}
}

func (u *Uint16RwRegister) Start() int {
	return u.start
}

func (u *Uint16RwRegister) Len() int {
	return 2
}

func (u *Uint16RwRegister) Name() string {
	return u.name
}

func (u *Uint16RwRegister) Addr() uint16 {
	return u.address
}

func (u *Uint16RwRegister) Decode(data []byte, values map[string]any) {
	u16 := u.order.Uint16(data)
	if u.decode == nil {
		values[u.name] = u16
		return
	}

	value, err := u.decode(u16, values)
	if err != nil {
		log.Error().Err(err).Msg(DecodeErr)
		return
	}

	values[u.name] = value
}

func (u *Uint16RwRegister) Encode(params map[string]any, dst []byte) error {
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

	if u.validate != nil {
		if err := u.validate(u16); err != nil {
			return err
		}
	}

	u.order.PutUint16(dst, u16)
	return nil
}
