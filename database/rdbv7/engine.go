package rdbv7

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"strings"
	"github.com/lsclh/gtools/log"
)

const (
	MethodOne      = "one"      //redis单节点
	MethodCluster  = "cluster"  //redis集群
	MethodFailover = "failover" //redis哨兵
)

func NewRDb(method string, opts ...rOptions) redis.Cmdable {
	opt := &options{
		Method:       method,
		MinIdleConns: 10,
		PoolSize:     300,
	}
	for _, fn := range opts {
		fn(opt)
	}

	ropt = opt
	if client == nil {
		return rdbInit()
	}
	return client
}

// **********************************************注册参数**********************************************************
type rOptions func(e *options)

// 基础信息
func WithBase(host, pwd string, db int) rOptions {
	return func(e *options) {
		e.Host = host
		e.Pwd = pwd
		e.Db = db
	}
}

// 哨兵模式使用
func WithMaster(name string) rOptions {
	return func(e *options) {
		e.Master = name
	}
}

// 链接池配置
func WithPoll(poolSize, minIdleConns int) rOptions {
	return func(e *options) {
		e.PoolSize = poolSize
		e.MinIdleConns = minIdleConns
	}
}

//**********************************************注册参数**********************************************************

type options struct {
	Method       string `json:"method"`
	Master       string `json:"master"`
	Host         string `json:"host"`
	Pwd          string `json:"pwd"`
	Db           int    `json:"db"`
	MinIdleConns int    `json:"minIdleConns"`
	PoolSize     int    `json:"pool_size"`
}

var ropt *options = nil

var client redis.Cmdable

func rdbInit() redis.Cmdable {
	switch ropt.Method {
	case MethodOne:
		client = redis.NewClient(&redis.Options{
			Addr:         ropt.Host,         //链接地址
			Password:     ropt.Pwd,          //密码
			DB:           ropt.Db,           //选择的db
			MinIdleConns: ropt.MinIdleConns, //最小维持连接数
			PoolSize:     ropt.PoolSize,     //连接池大小 6倍cpu核心数
		})
	case MethodFailover:
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    ropt.Master,
			SentinelAddrs: strings.Split(ropt.Host, ","),
			Password:      ropt.Pwd,
			DB:            ropt.Db,
			MinIdleConns:  ropt.MinIdleConns,
			PoolSize:      ropt.PoolSize,
		})
	case MethodCluster:
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        strings.Split(ropt.Host, ","),
			Password:     ropt.Pwd,
			MinIdleConns: ropt.MinIdleConns,
			PoolSize:     ropt.PoolSize,
		})

	}
	if _, err := client.Ping().Result(); err != nil {
		panic(fmt.Sprintf("RedisConnectError: %s", err.Error()))
		return nil
	}

	log.Println("RedisConnectSuccess")
	return client
}
