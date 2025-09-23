package mdb

import (
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	LogDebug = logger.Silent
	LogError = logger.Error
	LogWarn  = logger.Warn
	LogInfo  = logger.Info
)

// 获取一个任务对象 并创建添加任务
func NewDb(opts ...options) *gorm.DB {
	opt := &mOptions{
		PollMinConns: 5,
		PollMaxOpen:  10,
	}
	for _, fn := range opts {
		fn(opt)
	}

	return newMDb(opt)
}

// **********************************************注册参数**********************************************************
type options func(e *mOptions)

// 默认使用mysql 注册此函数使用oracle数据库
func WithOracle(serviceName string) options {
	return func(e *mOptions) {
		e.Oracle = &mOptionOracle{
			ServiceName: serviceName,
		}
	}
}

// WithBase 基础
func WithBase(host string, port int, user, pwd, dbname string) options {
	return func(e *mOptions) {
		e.Host = host
		e.Port = port
		e.User = user
		e.Pass = pwd
		e.Dbname = dbname
	}
}

// WithPoll 链接池配置
func WithPoll(PollMaxOpen, PollMinConns int) options {
	return func(e *mOptions) {
		e.PollMinConns = PollMinConns
		e.PollMaxOpen = PollMaxOpen
	}
}

// WithLog 日志
func WithLog(level logger.LogLevel, std logger.Writer) options {
	return func(e *mOptions) {
		e.Log = &mOptionLog{
			Level: level,
			Std:   std,
		}
	}
}

// WithSShKey ssh代理 仅支持mysql数据库
func WithSShKey(host, user, publicKey string) options {
	return func(e *mOptions) {
		e.Ssh = &mOptionSSH{
			Host:      host,
			User:      user,
			Pass:      "",
			PublicKey: publicKey,
		}
	}
}

// WithSShPwd ssh代理 仅支持mysql数据库
func WithSShPwd(host, user, pwd string) options {
	return func(e *mOptions) {
		e.Ssh = &mOptionSSH{
			Host:      host,
			User:      user,
			Pass:      pwd,
			PublicKey: "",
		}
	}
}

//**********************************************注册参数**********************************************************
