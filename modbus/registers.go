package modbus

var (
	AllRegister = []Register{&Switch, &OverCurrentTripSetting, &OverLoadTripSetting, &OverTemperatureTripSetting, &OverTemperatureTripSetting, &OverVoltageTripSetting, &UnderVoltageTripSetting, &LeakageTripSetting}

	// Switch 开关
	Switch = ControlRegister{
		name:    "Switch",
		address: 0x6001,
	}

	// OverCurrentTripSetting 过流跳闸定值
	OverCurrentTripSetting = ActionRegister{
		name:    "OverCurrentTripSetting",
		address: 0x8229,
		len:     2,
		tag:     0x2D,
	}

	// OverLoadTripSetting 过载跳闸定值
	OverLoadTripSetting = ActionRegister{
		name:    "OverLoadTripSetting",
		address: 0x8230,
		len:     2,
		tag:     0x2D,
	}

	// LeakageTripSetting 漏电跳闸定值
	LeakageTripSetting = ActionRegister{
		name:    "LeakageTripSetting",
		address: 0x8236,
		len:     2,
		tag:     0x2D,
	}

	// OverVoltageTripSetting 过压跳闸定值 0x823CH
	OverVoltageTripSetting = ActionRegister{
		name:    "OverVoltageTripSetting",
		address: 0x8236,
		len:     2,
		tag:     0x2D,
	}

	// UnderVoltageTripSetting 欠压跳闸定值
	UnderVoltageTripSetting = ActionRegister{
		name:    "UnderVoltageTripSetting",
		address: 0x8242,
		len:     2,
		tag:     0x2D,
	}

	// OverTemperatureTripSetting 过温跳闸定值
	OverTemperatureTripSetting = ActionRegister{
		name:    "OverTemperatureTripSetting",
		address: 0x824E,
		len:     2,
		tag:     0x2D,
	}
)

func FindRegister(name string) Register {
	for _, r := range AllRegister {
		if r.Name() == name {
			return r
		}
	}
	return nil
}
