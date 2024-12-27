package log

import (
	"github.com/lsclh/gtools/log/basics"
	"github.com/lsclh/gtools/log/internal"
)

// Println 控制台输出(生产模式不记录文件 正常输出控制台)
func Println(format string, v ...interface{}) {
	internal.Println(format, v...)
}

// Debug uses fmt.Sprintf to log a templated message.
// Debug("debug %s", "value")
func Debug(format string, args ...interface{}) {
	basics.Def.Z.Debugf(format, args...)
}

// Info uses fmt.Sprintf to log a templated message.
func Info(format string, args ...interface{}) {
	basics.Def.Z.Infof(format, args...)
}

// Warn uses fmt.Sprintf to log a templated message.
func Warn(format string, args ...interface{}) {
	basics.Def.Z.Warnf(format, args...)
}

// Error uses fmt.Sprintf to log a format message.
func Error(format string, args ...interface{}) {
	basics.Def.Z.Errorf(format, args...)
}

// Debugw Debugw("debug", "key", "value", "key2", "value")
// Debugw("debug", zap.String("key", "value"))
func Debugw(format string, keysAndValues ...interface{}) {
	basics.Def.Z.Debugw(format, keysAndValues...)
}

func Infow(format string, keysAndValues ...interface{}) {
	basics.Def.Z.Infow(format, keysAndValues...)
}

func Warnw(format string, keysAndValues ...interface{}) {
	basics.Def.Z.Warnw(format, keysAndValues...)
}

func Errorw(format string, keysAndValues ...interface{}) {
	basics.Def.Z.Errorw(format, keysAndValues...)
}
