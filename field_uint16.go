package main

import (
	"encoding/binary"
	"github.com/rs/zerolog/log"
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
			log.Error().Err(err).Msg("")
			return
		}
		values[u.name] = value
		return
	}
	values[u.name] = u16
}
