package main

import (
	"jg-gw/util"
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
