package log

import (
	"github.com/lsclh/gtools/log/internal"
	"sync"
)

var once sync.Once

const (
	LogFormatJson = "json"
	LogFormatText = "text"
)

type Logger struct {
	*internal.LogFile
}

func GetLogger(save int, fileName, format string) *Logger {
	return &Logger{
		internal.NewLog(save, fileName, format),
	}
}

func New(debug bool, logPath string) {
	opt := &internal.Options{
		Debug: debug,
		Dir:   logPath,
	}
	internal.New(opt)
}

// Println 控制台输出(生产模式不记录文件 正常输出控制台)
func Println(format string, v ...interface{}) {
	internal.Println(format, v...)
}
