package rLock

import (
	"sync"
	"time"
)

var Factory = factory{}
var Engine = &engine{}

// 锁模版
type Mutex interface {
	//TryLock 尝试加锁 成功返回true 失败返回false
	TryLock() bool
	//Lock 尝试加锁 成功往下执行 失败阻塞自旋
	Lock()
	//Unlock 解锁
	Unlock() bool
}

type factory struct{}

func (f factory) NewNullLock() Mutex {
	return &nullLock{}
}

// NewRedisLoopLocker 获取一个RedisMutex锁 可应用于集群
// 特性:
//  1. 同name 同key 不同的对象也可上锁成功 比如 key传机器码 同name当前机器各处都可以加锁成功 其他机器无法加锁成功
//     key传空表示不使用此功能
//  2. 不解锁不会进行自动释放 将永远锁死
//
// lName 锁名称 key 标记符 rdb redis操作连接对象
func (f factory) NewRedisLoopLocker(lName string, key string) Mutex {
	return &redisLoopLock{
		name: "lLock:" + lName,
		key:  key,
	}
}

// NewRedisTimeLocker 获取一个RedisMutex锁 可应用于集群
// 特性:
//  1. 同name 同key 不同的对象也可上锁成功
//     比如 key传机器码 同name当前机器各处都可以加锁成功且会延长过期时间 其他机器无法加锁成功
//  2. 同name 不同key 无法解锁成功
//  3. key传空表示不使用1,2功能
//  4. 不解锁到期将进行自动释放 使用时慎重
//
// lName 锁名称 key 标记符 rdb redis操作连接对象
func (f factory) NewRedisTimeLocker(lName, key string, ttl time.Duration) Mutex {
	return &redisTimeLock{
		name:    "tLock:" + lName,
		ttl:     ttl,
		backoff: time.Millisecond * 500,
		key:     key,
		mux:     new(sync.Mutex),
	}
}

type engine struct{}

func (engine) RegRds(r rds) {
	if rdbClient == nil {
		rdbClient = r
	}
}

// redis操作接口
type rds interface {
	//return true成功(说明没值写入成功) false=失败(说明有值写入失败)
	SetNX(key string, value interface{}, expiration time.Duration) (bool, error)
	//return false失败(说明没值无需删除) true成功(说明有值并删除了)
	Del(key string) (bool, error)
	//return 值 true存在key,false不存在key
	Get(key string) (string, bool, error)
	//return true成功(说明有值并更新了过期时间) false失败(说明没值无法更新过期时间)
	Expire(key string, expiration time.Duration) (bool, error)
}

var rdbClient rds = nil
