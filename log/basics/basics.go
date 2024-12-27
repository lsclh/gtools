package basics

import "github.com/lsclh/gtools/log/internal"

func New(opts ...options) {
	opt := &internal.Options{
		Debug: false,
	}
	for _, fn := range opts {
		fn(opt)
	}
	internal.New(opt)
	internal.NewLog(7, "log.txt", "")
}

func SetDefaultLog(save int, fileName string) {
	internal.NewLog(save, fileName, "")
}

func NewCustomLog(save int, fileName, format string) *internal.LogFile {
	return internal.NewLog(save, fileName, format)
}

// **********************************************注册参数**********************************************************
type options func(e *internal.Options)

func WithDebug() options {
	return func(e *internal.Options) {
		e.Debug = true
	}
}

func WithLogPath(path string) options {
	return func(e *internal.Options) {
		e.Dir = path
	}
}

//**********************************************注册参数**********************************************************

var Def *internal.LogFile = nil
