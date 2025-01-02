package log

import (
	"github.com/lsclh/gtools/log/internal"
)

type Logger struct {
	*internal.LogFile
}

var Factory = &factory{
	cnf: nil,
}

// Println 控制台输出(生产模式不记录文件 正常输出控制台)
func Println(format string, v ...interface{}) {
	internal.Println(format, v...)
}

// **********************************************内部依赖**********************************************************
func (o *factory) clone() *factory {
	if o.cnf != nil {
		return o
	}
	opt := &factory{
		cnf: &cnf{
			debug:  false,
			dir:    "./logs",
			save:   3,
			format: "text",
			name:   "log.log",
		},
	}
	opt.cnf.opt = opt
	return opt
}

func (o *factory) WithModeDebug() *factory {
	return o.clone().cnf.saveMode(true)
}
func (o *factory) WithModeRelease() *factory {
	return o.clone().cnf.saveMode(false)
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
func (o *factory) New() *Logger {
	return &Logger{
		internal.NewLog(o.cnf.debug, o.cnf.save, o.cnf.name, o.cnf.dir, o.cnf.format),
	}
}

type factory struct {
	cnf *cnf
}

type cnf struct {
	debug  bool
	dir    string
	save   int
	format string
	name   string
	opt    *factory
}

func (c *cnf) saveMode(debug bool) *factory {
	c.debug = debug
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
