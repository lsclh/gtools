package rLock

import (
	"sync"
	"time"
)

// NewRedisTimeLocker 获取一个RedisMutex锁 可应用于集群
// 特性:
//  1. 同name 同key 不同的对象也可上锁成功
//     比如 key传机器码 同name当前机器各处都可以加锁成功且会延长过期时间 其他机器无法加锁成功
//  2. 同name 不同key 无法解锁成功
//  3. key传空表示不使用1,2功能
//  4. 不解锁到期将进行自动释放 使用时慎重
//
// lName 锁名称 key 标记符 rdb redis操作连接对象
func NewRedisTimeLocker(lName, key string, ttl time.Duration) *RedisTimeLock {
	lock := &RedisTimeLock{
		name:    "tLock:" + lName,
		ttl:     ttl,
		backoff: time.Millisecond * 500,
		key:     key,
		mux:     new(sync.Mutex),
	}
	return lock
}

type RedisTimeLock struct {
	name    string
	backoff time.Duration
	ttl     time.Duration
	key     string
	mux     *sync.Mutex
}

func (l *RedisTimeLock) Lock() {
	for {
		if l.TryLock() {
			return
		}
		time.Sleep(l.backoff)
	}
}

func (l *RedisTimeLock) Unlock() bool {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.unlock()
}

func (l *RedisTimeLock) TryLock() bool {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.lock()
}

func (l *RedisTimeLock) lock() bool {
	ok, err := rdb.SetNX(l.name, l.key, l.ttl)
	if err == nil && ok {
		return true
	}
	if l.key == "" {
		return false
	}

	res, _, _ := rdb.Get(l.name)
	if res == l.key {
		// 是本次设置的锁 更新锁过期时间ttl
		if l.ttl > 0 {
			rdb.Expire(l.name, l.ttl)
		}
		//这个地方返回true会导致本机不会锁住
		return true
	}
	return false
}

func (l *RedisTimeLock) unlock() bool {
	oldVal, b, err := rdb.Get(l.name)
	if err != nil { //redis操作失败解锁失败
		return false
	}
	if !b { //redis没有这个key解锁成功
		return true
	}
	// 不是本次设置的锁 解锁失败
	if oldVal == l.key || oldVal == "" {
		// 是本次设置的锁 删除key
		_, err = rdb.Del(l.name)
		return err == nil //reids操作成功解锁成功
	}
	return false
}
