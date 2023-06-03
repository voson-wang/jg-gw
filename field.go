package main

import (
	"github.com/shopspring/decimal"
	"ricn-smart/ricn-jg-gw/modbus"
)

var SNField = NewStringField("SN", 0, 6)

// connReg 注册包
var connReg = modbus.NewDecodableFields(124, []modbus.DecodableField{
	SNField,
})

// livedataReg 实时数据包
var livedataReg = modbus.NewDecodableFields(66, []modbus.DecodableField{
	NewUint16FieldWithDecode("Ua", 14, unit10),
	NewUint16FieldWithDecode("Ub", 16, unit10),
	NewUint16FieldWithDecode("Uc", 18, unit10),
	NewUint16FieldWithDecode("Ia", 20, unit100),
	NewUint16FieldWithDecode("Ib", 22, unit100),
	NewUint16FieldWithDecode("Ic", 24, unit100),
	NewUint16Field("Leakage", 26),
	NewUint16Field("T1", 58),
	NewUint16Field("T2", 60),
	NewUint16Field("T3", 62),
	NewUint16Field("T4", 64),
})

func unit10(data uint16) (any, error) {
	return decimal.NewFromInt32(int32(data)).Mul(decimal.NewFromFloat(.1)).IntPart(), nil
}

func unit100(data uint16) (any, error) {
	return decimal.NewFromInt32(int32(data)).Mul(decimal.NewFromFloat(.01)).IntPart(), nil
}
