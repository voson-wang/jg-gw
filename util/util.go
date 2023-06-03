package util

import (
	"fmt"
	"golang.org/x/mod/modfile"
	"net"
	"os"
	"strconv"
	"strings"
)

func GetProjectNameFromModule() string {
	// 本地开发环境时，尝试获取模块名作为项目名称
	mod, err := os.ReadFile("go.mod")
	if err == nil {
		pathStr := modfile.ModulePath(mod)

		paths := strings.Split(pathStr, "/")

		return paths[len(paths)-1]
	}
	return ""
}

func GetLocalIP() (string, error) {
	adders, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range adders {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", nil
}

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
