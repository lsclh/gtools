package log

import (
	"github.com/lsclh/gtools/log/internal"
	"go.uber.org/zap"
	"io"
)

type Logger interface {
	Fio() io.Writer
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	Printf(format string, params ...any)
	Zap() *zap.SugaredLogger
}

var Factory = &factory{
	cnf: nil,
}

// Println 控制台输出(生产模式不记录文件 正常输出控制台)
func Println(format string, v ...any) {
	internal.Println(format, v...)
}

// **********************************************内部依赖**********************************************************
func (o *factory) clone() *factory {
	if o.cnf != nil {
		return o
	}
	opt := &factory{
		cnf: &cnf{
			outFile: false,
			outStd:  true,
			dir:     "./logs",
			save:    3,
			format:  "text",
			name:    "log.log",
			skip:    1,
		},
	}
	opt.cnf.opt = opt
	return opt
}

func (o *factory) WithOutfile(out bool) *factory {
	return o.clone().cnf.saveOutFile(out)
}
func (o *factory) WithOutstd(out bool) *factory {
	return o.clone().cnf.saveOutStd(out)
}
func (o *factory) WithSaveDir(logPath string) *factory {
	return o.clone().cnf.saveDir(logPath)
}
func (o *factory) WithSaveDay(day int) *factory {
	return o.clone().cnf.saveDay(day)
}
func (o *factory) WithFormatJson() *factory {
	return o.clone().cnf.saveFmt("json")
}
func (o *factory) WithFormatText() *factory {
	return o.clone().cnf.saveFmt("text")
}
func (o *factory) WithFileName(name string) *factory {
	return o.clone().cnf.saveName(name)
}
func (o *factory) AddCallerSkip(skip int) *factory {
	return o.clone().cnf.saveSkip(skip)
}
func (o *factory) New() Logger {
	return internal.NewLog(o.cnf.outFile, o.cnf.outStd, o.cnf.save, o.cnf.name, o.cnf.dir, o.cnf.format, o.cnf.skip)
}

type factory struct {
	cnf *cnf
}

type cnf struct {
	outFile bool
	outStd  bool
	dir     string
	save    int
	format  string
	name    string
	skip    int
	opt     *factory
}

func (c *cnf) saveOutFile(open bool) *factory {
	c.outFile = open
	return c.opt
}
func (c *cnf) saveOutStd(open bool) *factory {
	c.outStd = open
	return c.opt
}
func (c *cnf) saveDir(logPath string) *factory {
	c.dir = logPath
	return c.opt
}
func (c *cnf) saveDay(day int) *factory {
	c.save = day
	return c.opt
}
func (c *cnf) saveFmt(format string) *factory {
	c.format = format
	return c.opt
}

func (c *cnf) saveName(name string) *factory {
	c.name = name
	return c.opt
}
func (c *cnf) saveSkip(skip int) *factory {
	c.skip = skip
	return c.opt
}
