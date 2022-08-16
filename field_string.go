package main

import (
	"fmt"
	"jg-gateway/util"
	"reflect"
)

type StringField struct {
	name  string
	start int
	len   int
}

func NewStringField(name string, start, len int) *StringField {
	return &StringField{
		name:  name,
		start: start,
		len:   len,
	}
}

func (s *StringField) Decode(data []byte, values map[string]any) {
	values[s.name] = util.BytesToString(data)
}

func (s *StringField) Name() string {
	return s.name
}

func (s *StringField) Start() int {
	return s.start
}

func (s *StringField) Len() int {
	return s.len
}

type StringWoField struct {
	name   string
	start  int
	len    int
	encode func(value string, dst []byte) error
}

func NewStringWoField(name string, start int, len int, encode func(value string, dst []byte) error) *StringWoField {
	return &StringWoField{name: name, start: start, len: len, encode: encode}
}

func (s *StringWoField) Encode(params map[string]interface{}, dst []byte) error {
	value, ok := params[s.name]
	if !ok {
		return fmt.Errorf("参数 %v 缺失", s.name)
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("参数 %v 类型错误，期望：string，实际：%v", s.name, reflect.TypeOf(value))
	}

	return s.encode(str, dst)
}

func (s *StringWoField) Name() string {
	return s.name
}

func (s *StringWoField) Start() int {
	return s.start
}

func (s *StringWoField) Len() int {
	return s.len
}
