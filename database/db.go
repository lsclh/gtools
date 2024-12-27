package database

import (
	"github.com/lsclh/gtools/database/internal"
	"gorm.io/gorm/logger"
)

// 获取一个任务对象 并创建添加任务
func NewDb(opts ...mOptions) *internal.MOptions {
	opt := &internal.MOptions{}
	for _, fn := range opts {
		fn(opt)
	}

	return internal.NewDb(opt)
}

const (
	RDbMethodOne      = "one"      //redis单节点
	RDbMethodCluster  = "cluster"  //redis集群
	RDbMethodFailover = "failover" //redis哨兵
)

func NewRDb(method string, opts ...rOptions) *internal.ROptions {
	opt := &internal.ROptions{
		Method: method,
	}
	for _, fn := range opts {
		fn(opt)
	}

	return internal.NewRdb(opt)
}

func SetLog(l internal.Log) {
	internal.SetLog(l)
}

// **********************************************注册参数**********************************************************
type mOptions func(e *internal.MOptions)

func DbWithBase(host string, port int, user, pwd, dbname string) mOptions {
	return func(e *internal.MOptions) {
		e.Host = host
		e.Port = port
		e.User = user
		e.Pass = pwd
		e.Dbname = dbname
	}
}
func DbWithConn(PollMaxOpen, PollMinConns int) mOptions {
	return func(e *internal.MOptions) {
		e.PollMinConns = PollMinConns
		e.PollMaxOpen = PollMaxOpen
	}
}
func DbWithLog(level logger.LogLevel, std logger.Writer) mOptions {
	return func(e *internal.MOptions) {
		e.Log = &internal.MOptionLog{
			Level: level,
			Std:   std,
		}
	}
}
func DBWithSShKey(host, user, publicKey string) mOptions {
	return func(e *internal.MOptions) {
		e.Ssh = &internal.MOptionSSH{
			Host:      host,
			User:      user,
			Pass:      "",
			PublicKey: publicKey,
		}
	}
}
func DBWithSShPwd(host, user, pwd string) mOptions {
	return func(e *internal.MOptions) {
		e.Ssh = &internal.MOptionSSH{
			Host:      host,
			User:      user,
			Pass:      pwd,
			PublicKey: "",
		}
	}
}

//**********************************************注册参数**********************************************************

// **********************************************注册参数**********************************************************
type rOptions func(e *internal.ROptions)

func RDbWithBase(host, pwd string, db int) rOptions {
	return func(e *internal.ROptions) {
		e.Host = host
		e.Pwd = pwd
		e.Db = db
	}
}

func RDbWithMaster(name string) rOptions {
	return func(e *internal.ROptions) {
		e.Master = name
	}
}

//**********************************************注册参数**********************************************************
