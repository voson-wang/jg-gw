package util

import (
	"golang.org/x/mod/modfile"
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
