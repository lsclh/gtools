package internal

import (
	"context"
	"fmt"
	gLog "github.com/lsclh/gtools/log"
	"github.com/redis/go-redis/v9"
	"strings"
)

const (
	RedisMethodOne      = "one"      //redis单节点
	RedisMethodCluster  = "cluster"  //redis集群
	RedisMethodFailover = "failover" //redis哨兵
)

var Client redis.Cmdable

type ROptions struct {
	Method       string `json:"method"`
	Master       string `json:"master"`
	Host         string `json:"host"`
	Pwd          string `json:"pwd"`
	Db           int    `json:"db"`
	MinIdleConns int    `json:"minIdleConns"`
	PoolSize     int    `json:"pool_size"`
}

var ropt *ROptions = nil

func NewRdb(o *ROptions) redis.Cmdable {
	ropt = o
	if Client == nil {
		return rdbInit()
	}
	return Client
}

func rdbInit() redis.Cmdable {
	ctx := context.Background()
	switch ropt.Method {
	case RedisMethodOne:
		Client = redis.NewClient(&redis.Options{
			Addr:         ropt.Host,         //链接地址
			Password:     ropt.Pwd,          //密码
			DB:           ropt.Db,           //选择的db
			MinIdleConns: ropt.MinIdleConns, //最小维持连接数
			PoolSize:     ropt.PoolSize,     //连接池大小 6倍cpu核心数
		})
	case RedisMethodFailover:
		Client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    ropt.Master,
			SentinelAddrs: strings.Split(ropt.Host, ","),
			Password:      ropt.Pwd,
			DB:            ropt.Db,
			MinIdleConns:  ropt.MinIdleConns,
			PoolSize:      ropt.PoolSize,
		})
	case RedisMethodCluster:
		Client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        strings.Split(ropt.Host, ","),
			Password:     ropt.Pwd,
			MinIdleConns: ropt.MinIdleConns,
			PoolSize:     ropt.PoolSize,
		})

	}
	if _, err := Client.Ping(ctx).Result(); err != nil {
		panic(fmt.Sprintf("RedisConnectError: %s", err.Error()))
		return nil
	}

	gLog.Println("RedisConnectSuccess")
	return Client
}
