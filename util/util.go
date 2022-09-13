package util

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func BytesToString(data []byte) string {
	var ss string
	for _, v := range data {
		ss += fmt.Sprintf("%02X", v)
	}
	return ss
}

// GetLineModel 京硅线路的编号开头两位决定了线路型号
func GetLineModel(data string) (string, error) {
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

// Shuffle 数组洗牌
func Shuffle[T any](arr []T) []T {
	slc := make([]T, len(arr))
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for i, randIndex := range r.Perm(len(arr)) {
		slc[i] = arr[randIndex]
	}
	return slc
}
