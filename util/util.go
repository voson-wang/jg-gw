package util

import (
	"fmt"
	"strconv"
)

func BytesToString(data []byte) string {
	var ss string
	for _, v := range data {
		ss += fmt.Sprintf("%02X", v)
	}
	return ss
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
	for index := range bs {
		m, err := strconv.Atoi(s[index*2 : index*2+2])
		if err != nil {
			return nil, err
		}
		bs[index] = byte(((m / 10) << 4) + m%10)
	}
	return bs, nil
}
