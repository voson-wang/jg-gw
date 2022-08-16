package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/mod/modfile"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var projectName string

var env string

var host string

func Init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	projectName = os.Getenv("PROJECT_NAME")

	if projectName == "" {
		// 获取模块名作为项目名称
		mod, err := os.ReadFile("go.mod")
		if err != nil {
			log.Fatal(err)
		}

		path := modfile.ModulePath(mod)

		if path == "" || strings.Contains(path, "/") {
			log.Fatal(fmt.Sprintf("invalid module path: %v", path))
		}
		projectName = path

	}

	// 读取环境配置
	env = os.Getenv("ENV")
	if "" == env {
		env = "development"
	}

	// 读取环境变量文件，如果不存在则忽略
	_ = godotenv.Load("version", ".env."+env)

	url := os.Getenv("IP_QUERY_ADDRESS")

	// 获取项目运行环境的IP
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err, url)
	}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	host = string(buf)
}

func Env() string {
	return env
}

func ProjectName() string {
	return projectName
}

func Debug() bool {
	return os.Getenv("DEBUG") == "true"
}

func MQTTDebug() bool {
	return os.Getenv("MQTT_DEBUG") == "true"
}

func Host() string {
	return host
}
