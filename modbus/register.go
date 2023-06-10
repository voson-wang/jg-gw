package modbus

type Register interface {
	// Name 寄存器名称
	Name() string

	// Address 寄存器地址
	Address() uint16

	// Len 返回字节长度
	Len() uint8
}

type Readable interface {
	// Decode 为了适应单个字段，两个字节代表2个参数的情况和
	// 单个字段，每个比特代表不同参数的情况
	// 所以使用了map去获取decode结果
	Decode(data []byte, results map[string]any)
}

type Writable interface {
	// Encode 为了适应单个字段，两个字节代表2个参数的情况和
	// 单个字段，每个比特代表不同参数的情况
	// 所以使用map来作为输入值，让字段自行取用输入值
	Encode(params map[string]any) ([]byte, error)
}

type RoRegister interface {
	Register
	Readable
}

type RwRegister interface {
	Register
	Readable
	Writable
}

type WoRegister interface {
	Register
	Writable
}
