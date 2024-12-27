package log

import (
	"github.com/lsclh/gtools/log/internal"
	"io"
)

func Init(opts ...options) {
	opt := &internal.Options{}
	for _, fn := range opts {
		fn(opt)
	}
	internal.New(opt)
	if Default == nil {
		SetDefaultLog(7, "log")
	}
}

func SetDefaultLog(save int, fileName string) {
	Default = &BaseLog{
		Save: save,
		File: fileName,
	}
	internal.Initlog(Default)
}

// **********************************************注册参数**********************************************************
type options func(e *internal.Options)

func WithDebug(debug bool) options {
	return func(e *internal.Options) {
		e.Debug = debug
	}
}

func WithLogPath(path string) options {
	return func(e *internal.Options) {
		e.Dir = path
	}
}

//**********************************************注册参数**********************************************************

func InitCustomLogger(l *BaseLog) {
	internal.Initlog(l)
}

var Default *BaseLog = nil

type BaseLog struct {
	*internal.LogFile
	//存储时间 (天)
	Save int
	//文件名
	File string `json:"file"`
	//存储格式 json/console
	Format string `json:"format"`
}

func (b *BaseLog) GetWriter() io.Writer {
	return b.Fio
}

func (b *BaseLog) GetConfig() (string, int) {
	return b.File, b.Save
}
func (b *BaseLog) GetFormat() string {
	return b.Format
}
func (b *BaseLog) Init(file *internal.LogFile) {
	b.LogFile = file
}

// Println 控制台输出(生产模式不记录文件 正常输出控制台)
func Println(format string, v ...interface{}) {
	internal.Println(format, v...)
}

// Debug uses fmt.Sprintf to log a templated message.
// Debug("debug %s", "value")
func Debug(format string, args ...interface{}) {
	Default.Z.Debugf(format, args...)
}

// Info uses fmt.Sprintf to log a templated message.
func Info(format string, args ...interface{}) {
	Default.Z.Infof(format, args...)
}

// Warn uses fmt.Sprintf to log a templated message.
func Warn(format string, args ...interface{}) {
	Default.Z.Warnf(format, args...)
}

// Error uses fmt.Sprintf to log a format message.
func Error(format string, args ...interface{}) {
	Default.Z.Errorf(format, args...)
}

// Debugw Debugw("debug", "key", "value", "key2", "value")
// Debugw("debug", zap.String("key", "value"))
func Debugw(format string, keysAndValues ...interface{}) {
	Default.Z.Debugw(format, keysAndValues...)
}

func Infow(format string, keysAndValues ...interface{}) {
	Default.Z.Infow(format, keysAndValues...)
}

func Warnw(format string, keysAndValues ...interface{}) {
	Default.Z.Warnw(format, keysAndValues...)
}

func Errorw(format string, keysAndValues ...interface{}) {
	Default.Z.Errorw(format, keysAndValues...)
}
