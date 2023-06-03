package modbus

// Field 京硅协议中没有寄存器地址的消息体
type Field interface {
	// Name 字段名称
	Name() string
	// Start 字段起始位置
	Start() int
	// Len 字段长度，单位：字节
	Len() int
}

type Decoder interface {
	Decode(data []byte, values map[string]any)
}

type DecodableField interface {
	Field
	Decoder
}

// DecodableFields
// len不一定等于fields中字段的长度
// 因此需要len字段来确定字段组整体长度
type DecodableFields struct {
	len int

	fields []DecodableField
}

func NewDecodableFields(len int, fields []DecodableField) *DecodableFields {
	return &DecodableFields{
		len:    len,
		fields: fields,
	}
}

func (p *DecodableFields) Len() int {
	return p.len
}

func (p *DecodableFields) Decode(data []byte, values map[string]any) {
	for _, fd := range p.fields {
		start := fd.Start()
		end := start + fd.Len()
		fd.Decode(data[start:end], values)
	}
}

type Encoder interface {
	Encode(params map[string]interface{}, dst []byte) error
}
