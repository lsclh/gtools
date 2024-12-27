package internal

type Log interface {
	Println(format string, args ...any)
}

var l Log = nil

func SetLog(lg Log) {
	l = lg
}

type slogger struct{}

var logger = &slogger{}

func (slogger) println(format string, args ...any) {
	if l != nil {
		l.Println(format, args...)
	}
}
