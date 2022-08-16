package main

import "github.com/rs/zerolog/log"

type ByteDecode func(b byte) (any, error)

// ByteField 单字节字段
type ByteField struct {
	name   string
	start  int
	decode ByteDecode
}

func NewByteField(name string, start int) *ByteField {
	return &ByteField{name: name, start: start}
}

func NewByteFieldWithDecode(name string, start int, decode ByteDecode) *ByteField {
	return &ByteField{name: name, start: start, decode: decode}
}

func (b *ByteField) Decode(data []byte, values map[string]any) {
	value := data[0]
	if b.decode != nil {
		result, err := b.decode(value)
		if err != nil {
			log.Error().Err(err).Msg(DecodeErr)
			return
		}
		values[b.name] = result
		return
	}
	values[b.name] = value
}

func (b *ByteField) Name() string {
	return b.name
}

func (b *ByteField) Start() int {
	return b.start
}

func (b *ByteField) Len() int {
	return 1
}
