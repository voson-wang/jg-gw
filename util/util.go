package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// RemoveZero 移除ASCII中的空字符（即0x00字节）
// 默认从左往右读取，遇到空字符即停止
func RemoveZero(data []byte) string {
	for index, b := range data {
		if b == 0 {
			return string(data[0:index])
		}
	}
	return string(data)
}

// ConvertBytesToTime
// 将物联网设备中常见的形如20, 3, 12, 17, 19, 00的字节流
// 转换为形如2020-03-12 17:19:00这样的格式
func ConvertBytesToTime(packet []byte) string {
	// 边界检查
	_ = packet[5]
	var b strings.Builder
	year := packet[0]
	month := packet[1]
	day := packet[2]
	hour := packet[3]
	minute := packet[4]
	second := packet[5]
	b.WriteString(time.Now().Format("2006")[0:2])
	b.WriteString(fmt.Sprintf(`%v-`, year))
	if month >= 10 {
		b.WriteString(fmt.Sprintf(`%v-`, month))
	} else {
		b.WriteString(fmt.Sprintf(`0%v-`, month))
	}
	if day >= 10 {
		b.WriteString(fmt.Sprintf(`%v `, day))
	} else {
		b.WriteString(fmt.Sprintf(`0%v `, day))
	}
	if hour >= 10 {
		b.WriteString(fmt.Sprintf(`%v:`, hour))
	} else {
		b.WriteString(fmt.Sprintf(`0%v:`, hour))
	}
	if minute >= 10 {
		b.WriteString(fmt.Sprintf(`%v:`, minute))
	} else {
		b.WriteString(fmt.Sprintf(`0%v:`, minute))
	}
	if second >= 10 {
		b.WriteString(fmt.Sprintf(`%v`, second))
	} else {
		b.WriteString(fmt.Sprintf(`0%v`, second))
	}
	return b.String()
}

func IPToBytes(ip string, dst []byte) {
	a := strings.Split(ip, ".")
	for index, b := range a {
		i, _ := strconv.Atoi(b)
		dst[index] = byte(i)
	}
}

func BytesToString(data []byte) string {
	var ss string
	for _, v := range data {
		ss += fmt.Sprintf("%02X", v)
	}
	return ss
}

func CheckSum(bs []byte) (b byte) {
	u := uint16(0x00)
	for _, d := range bs {
		u = u + uint16(d)
	}
	u = u % 256
	return byte(u)
}

func GetModel(data string) (string, error) {
	var model string
	switch data[0:2] {
	case "04":
		model = "4P_L"
	case "07":
		model = "2P_L"
	case "08":
		model = "1P"
	case "10":
		model = "2P"
	case "11":
		model = "3P"
	case "12":
		model = "4P"
	default:
		return "", fmt.Errorf("未知型号")
	}

	return model, nil
}

func SetByteSN(s string) ([]byte, error) {

	bs := make([]byte, 6)
	for index, _ := range bs {
		m, err := strconv.Atoi(s[index*2 : index*2+2])
		if err != nil {
			return nil, err
		}
		bs[index] = byte(((m / 10) << 4) + m%10)
	}
	return bs, nil
}
