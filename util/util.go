package util

import (
	"golang.org/x/mod/modfile"
	"net"
	"os"
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
