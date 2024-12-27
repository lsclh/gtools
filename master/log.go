package master

type Log interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
}

var l Log = nil

func SetLog(lg Log) {
	if l == nil {
		l = lg
	}
}

type slogger struct{}

var logger = &slogger{}

func (slogger) error(format string, args ...any) {
	if l != nil {
		l.Error(format, args...)
	}
}
func (slogger) info(format string, args ...any) {
	if l != nil {
		l.Info(format, args...)
	}
}
