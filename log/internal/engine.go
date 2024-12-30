package internal

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"runtime"
	"time"
)

// var blog *logs.BeeLogger
var (
	yellow  = string([]byte{27, 91, 51, 51, 109})
	blue2   = string([]byte{27, 91, 51, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
	pathSep = string(os.PathSeparator)
)

type Options struct {
	Debug bool
	Dir   string
}

var opt *Options = nil

func New(o *Options) {
	opt = o
}

func NewLog(save int, fileName, format string) *LogFile {
	if opt == nil {
		opt = &Options{
			Debug: false,
			Dir:   "./logs",
		}
	}
	fio := getWriter(fileName, save)
	z := NewZapLogger(fio, format)
	return &LogFile{
		fio:  fio,
		Z:    z,
		sync: z.Sync,
	}
}

type LogFile struct {
	fio  io.Writer
	Z    *zap.SugaredLogger
	sync func() error
}

func (f *LogFile) Fio() io.Writer {
	return f.fio
}

// Debug uses fmt.Sprintf to log a templated message.
func (l *LogFile) Debug(format string, args ...interface{}) {
	l.Z.Debugf(format, args...)
}

// Info uses fmt.Sprintf to log a templated message.
func (l *LogFile) Info(format string, args ...interface{}) {
	l.Z.Infof(format, args...)
}

// Warn uses fmt.Sprintf to log a templated message.
func (l *LogFile) Warn(format string, args ...interface{}) {
	l.Z.Warnf(format, args...)
}

// Error uses fmt.Sprintf to log a templated message.
func (l *LogFile) Error(format string, args ...interface{}) {
	l.Z.Errorf(format, args...)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
// s.With(keysAndValues).Debug(msg)
func (l *LogFile) Debugw(msg string, keysAndValues ...interface{}) {
	l.Z.Debugw(msg, keysAndValues...)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *LogFile) Infow(msg string, keysAndValues ...interface{}) {
	l.Z.Infow(msg, keysAndValues...)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *LogFile) Warnw(msg string, keysAndValues ...interface{}) {
	l.Z.Warnw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *LogFile) Errorw(msg string, keysAndValues ...interface{}) {
	l.Z.Errorw(msg, keysAndValues...)
}

func (l *LogFile) Printf(format string, params ...interface{}) {
	l.Z.Info(fmt.Sprintf(format, params...))
}

// NewZapLogger 创建 ZapLogger
func NewZapLogger(f io.Writer, format string) *zap.SugaredLogger {
	// 动态日志等级
	//文件+控制台
	var core zapcore.Core
	if opt.Debug {
		w := zapcore.AddSync(f)
		if format == "json" {
			core = zapcore.NewTee(
				zapcore.NewCore(zapcore.NewJSONEncoder(NewEncoderConfig(true)), os.Stdout, zap.DebugLevel),
				zapcore.NewCore(zapcore.NewConsoleEncoder(NewEncoderConfig(false)), w, zap.DebugLevel),
			)
		} else {
			core = zapcore.NewTee(
				zapcore.NewCore(zapcore.NewConsoleEncoder(NewEncoderConfig(true)), os.Stdout, zap.DebugLevel),
				zapcore.NewCore(zapcore.NewConsoleEncoder(NewEncoderConfig(false)), w, zap.DebugLevel),
			)
		}

	} else {
		w := zapcore.AddSync(f)
		if format == "json" {
			core = zapcore.NewTee(
				zapcore.NewCore(zapcore.NewJSONEncoder(NewEncoderConfig(false)), w, zapcore.InfoLevel),
			)
		} else {
			core = zapcore.NewTee(
				zapcore.NewCore(zapcore.NewConsoleEncoder(NewEncoderConfig(false)), w, zapcore.InfoLevel),
			)
		}

	}

	//向上跳一层 到调用位置
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return logger.Sugar()
}

// TimeEncoder 格式化时间
func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func NewEncoderConfig(color bool) zapcore.EncoderConfig {
	lc := zapcore.CapitalLevelEncoder
	if color {
		lc = zapcore.CapitalColorLevelEncoder
	}
	return zapcore.EncoderConfig{
		// Keys can be interface{}thing except the empty string.
		TimeKey:        "T",                           // json时时间键
		LevelKey:       "L",                           // json时日志等级键
		NameKey:        "N",                           // json时日志记录器名
		CallerKey:      "C",                           // json时日志文件信息键
		MessageKey:     "M",                           // json时日志消息键
		StacktraceKey:  "S",                           // json时堆栈键
		LineEnding:     zapcore.DefaultLineEnding,     // 友好日志换行符
		EncodeTime:     TimeEncoder,                   // 友好日志时日期格式化
		EncodeDuration: zapcore.StringDurationEncoder, // 时间序列化
		EncodeCaller:   zapcore.ShortCallerEncoder,    // 日志文件信息（包/文件.go:行号）
		EncodeLevel:    lc,                            // 友好日志等级名大小写（info INFO）

		//EncodeCaller: zapcore.FullCallerEncoder, // 日志文件信息（路径/文件.go:行号）
	}
}

// 按时间划分
func getWriter(filename string, saveTime int) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存7天内的日志，每1小时(整点)分割一次日志
	output_dir := opt.Dir + pathSep
	hook, err := rotatelogs.New(
		// 没有使用go风格反人类的format格式
		output_dir+filename+".%Y_%m_%d",
		rotatelogs.WithLinkName(output_dir+filename),
		rotatelogs.WithMaxAge(time.Hour*24*time.Duration(saveTime)), //保存天数
		rotatelogs.WithRotationTime(time.Hour*24),                   //24小时轮转
	)
	if err != nil {
		panic(err)
	}
	return hook
}

// 控制台输出(生产模式不记录文件 正常输出控制台)
func Println(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	fmt.Println(fmt.Sprintf(" %s | %s | %s ", fmt.Sprint(yellow, "Debug", reset), fmt.Sprint(blue2, fmt.Sprintf(format, v...), reset), fmt.Sprintf("%s:%d", file, line)))
}
