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
func NewDb(opts ...mOptions) *gorm.DB {
	opt := &MOptions{
		PollMinConns: 5,
		PollMaxOpen:  10,
	}
	for _, fn := range opts {
		fn(opt)
	}

	return NewMDb(opt)
}

// **********************************************注册参数**********************************************************
type mOptions func(e *MOptions)

// 基础
func WithBase(host string, port int, user, pwd, dbname string) mOptions {
	return func(e *MOptions) {
		e.Host = host
		e.Port = port
		e.User = user
		e.Pass = pwd
		e.Dbname = dbname
	}
}

// 链接池配置
func WithPoll(PollMaxOpen, PollMinConns int) mOptions {
	return func(e *MOptions) {
		e.PollMinConns = PollMinConns
		e.PollMaxOpen = PollMaxOpen
	}
}

// 日志
func WithLog(level logger.LogLevel, std logger.Writer) mOptions {
	return func(e *MOptions) {
		e.Log = &MOptionLog{
			Level: level,
			Std:   std,
		}
	}
}

// ssh代理
func WithSShKey(host, user, publicKey string) mOptions {
	return func(e *MOptions) {
		e.Ssh = &MOptionSSH{
			Host:      host,
			User:      user,
			Pass:      "",
			PublicKey: publicKey,
		}
	}
}

// ssh代理
func WithSShPwd(host, user, pwd string) mOptions {
	return func(e *MOptions) {
		e.Ssh = &MOptionSSH{
			Host:      host,
			User:      user,
			Pass:      pwd,
			PublicKey: "",
		}
	}
}

//**********************************************注册参数**********************************************************
