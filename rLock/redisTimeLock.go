package rLock

import (
	"sync"
	"time"
)

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
