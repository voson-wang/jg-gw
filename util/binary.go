package util

import (
	"encoding/binary"
)

func TwoByteToBits(data []byte, order binary.ByteOrder) (dst []byte) {
	_ = data[1]
	dst = make([]byte, 16)
	first := ByteToBits(data[0])
	second := ByteToBits(data[1])

	if order == binary.BigEndian {
		dst = append(second, first...)
	} else {
		dst = append(first, second...)
	}
	return
}

func BitsToTwoByte(bits []byte, order binary.ByteOrder, dst []byte) {

	first := BitsToByte(bits[0:8])
	second := BitsToByte(bits[8:16])

	if order == binary.BigEndian {
		dst[0] = second
		dst[1] = first
	} else {
		dst[0] = first
		dst[1] = second
	}
}

// ByteToBits 字节转比特
// 十进制转二进制数组
func ByteToBits(d byte) []byte {
	b := make([]byte, 8)
	for i := 0; d > 0; d /= 2 {
		b[i] = d % 2
		i++
	}
	return b
}

// BitsToByte 比特转字节
// 二进制组转十进制
// 大于255的二进制会被切割，u最大返回255
func BitsToByte(buf []byte) (u byte) {
	for i := 0; i < len(buf); i++ {
		u += buf[i] << i
	}
	return u
}
