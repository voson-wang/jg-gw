package log

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strconv"
)

func Init(debug bool, path string) {
	// 必须添加这句，否则还是会成为debug level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// 显示文件和行号
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	multi := zerolog.MultiLevelWriter(&zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: "2006/01/02 15:04:05"}, &lumberjack.Logger{
		Filename:   path,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true, // 开启压缩
	})

	log.Logger = zerolog.New(multi).With().Timestamp().Caller().Logger()
}
