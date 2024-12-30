package custom

import "github.com/lsclh/gtools/log/internal"

const (
	LogFormatJson = "json"
	LogFormatText = "text"
)

func EditDefCnf(save int, fileName string) {
	Def = internal.NewLog(save, fileName, "")
}

func NewLogger(save int, fileName, format string) *internal.LogFile {
	return internal.NewLog(save, fileName, format)
}

var Def *internal.LogFile = nil
