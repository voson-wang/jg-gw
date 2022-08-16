package main

import (
	"e.coding.net/ricnsmart/service/jg-modbus"
)

var (
	RWPacket = jg.ReadAndWritableRegisters{Status, OverCurrentValue, OverCurrentDelay, OverLoadValue,
		OverLoadDelay, LeakageValue, LeakageDelay, OverVoltageValue, OverVoltageDelay, UnderVoltageValue, UnderVoltageDelay, OverTemperatureValue, OverTemperatureDelay, ShortValue, ShortDelay}

	AlarmSettingPacket = jg.ReadAndWritableRegisters{OverCurrentValue, OverCurrentDelay, OverLoadValue, OverLoadDelay,
		LeakageValue, LeakageDelay, OverVoltageValue, OverVoltageDelay, UnderVoltageValue, UnderVoltageDelay, OverTemperatureValue, OverTemperatureDelay, ShortValue, ShortDelay}

	SwitchPacket = jg.ReadAndWritableRegisters{Status}

	Status = NewByteRwRegister("Status", 0, 24577)

	OverCurrentValue     = NewUint16RwRegister("OverCurrentValue", 0, 0x822c)
	OverCurrentDelay     = NewUint16RwRegister("OverCurrentDelay", 2, 0x822d)
	OverLoadValue        = NewUint16RwRegister("OverLoadValue", 4, 0x8233)
	OverLoadDelay        = NewUint16RwRegister("OverLoadDelay", 6, 0x8234)
	LeakageValue         = NewUint16RwRegister("LeakageValue", 8, 0x8239)
	LeakageDelay         = NewUint16RwRegister("LeakageDelay", 10, 0x823a)
	OverVoltageValue     = NewUint16RwRegister("OverVoltageValue", 12, 0x823f)
	OverVoltageDelay     = NewUint16RwRegister("OverVoltageDelay", 14, 0x8240)
	UnderVoltageValue    = NewUint16RwRegister("UnderVoltageValue", 16, 0x8245)
	UnderVoltageDelay    = NewUint16RwRegister("UnderVoltageDelay", 18, 0x8246)
	OverTemperatureValue = NewUint16RwRegister("OverTemperatureValue", 20, 0x8251)
	OverTemperatureDelay = NewUint16RwRegister("OverTemperatureDelay", 22, 0x8252)
	ShortValue           = NewUint16RwRegister("ShortValue", 24, 0x8225)
	ShortDelay           = NewUint16RwRegister("ShortDelay", 26, 0x8226)
)
