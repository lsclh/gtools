package internal

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/redis/go-redis/v9"
)

const (
	RedisMethodOne      = "one"      //redis单节点
	RedisMethodCluster  = "cluster"  //redis集群
	RedisMethodFailover = "failover" //redis哨兵
)

var Client redis.Cmdable

type ROptions struct {
	Method string `json:"method"`
	Master string `json:"master"`
	Host   string `json:"host"`
	Pwd    string `json:"pwd"`
	Db     int    `json:"db"`
}

var ropt *ROptions = nil

func NewRdb(o *ROptions) *ROptions {
	ropt = o
	return ropt
}

func (ROptions) Init() {
	ctx := context.Background()
	switch ropt.Method {
	case RedisMethodOne:
		Client = redis.NewClient(&redis.Options{
			Addr:         ropt.Host, //链接地址
			Password:     ropt.Pwd,  //密码
			DB:           ropt.Db,   //选择的db
			MinIdleConns: 10,        //最小维持连接数
			PoolSize:     300,       //连接池大小 6倍cpu核心数
		})
	case RedisMethodFailover:
		Client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    ropt.Master,
			SentinelAddrs: gstr.Split(ropt.Host, ","),
			Password:      ropt.Pwd,
			DB:            ropt.Db,
			MinIdleConns:  10,
			PoolSize:      300,
		})
	case RedisMethodCluster:
		Client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        gstr.Split(ropt.Host, ","),
			Password:     ropt.Pwd,
			MinIdleConns: 10,
			PoolSize:     300,
		})

	}
	if _, err := Client.Ping(ctx).Result(); err != nil {
		panic(fmt.Sprintf("RedisConnectError: %s", err.Error()))
	}

	logger.println("RedisConnectSuccess")

}
