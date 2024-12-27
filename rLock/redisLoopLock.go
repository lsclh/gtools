package rLock

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// NewRedisLoopLocker 获取一个RedisMutex锁 可应用于集群
// 特性:
//  1. 同name 同key 不同的对象也可上锁成功 比如 key传机器码 同name当前机器各处都可以加锁成功 其他机器无法加锁成功
//     key传空表示不使用此功能
//  2. 不解锁不会进行自动释放 将永远锁死
//
// lName 锁名称 key 标记符 rdb redis操作连接对象
func NewRedisLoopLocker(lName string, key string) *RedisLoopLock {
	r := &RedisLoopLock{
		name: "lLock:" + lName,
		key:  key,
	}
	return r
}

type RedisLoopLock struct {
	l      sync.Mutex
	name   string
	key    string
	ctx    context.Context
	cancel context.CancelFunc
}

func (r *RedisLoopLock) TryLock() bool {
	r.l.Lock()
	defer r.l.Unlock()
	return r.lock()
}

func (r *RedisLoopLock) Lock() {
	r.l.Lock()
	defer r.l.Unlock()
	for {
		if r.lock() {
			return
		}
		time.Sleep(time.Second)
	}
}

func (r *RedisLoopLock) Unlock() bool {
	r.l.Lock()
	defer r.l.Unlock()
	if r.cancel != nil {
		r.cancel()
	}
	_, err := rdb.Del(r.name)
	return err == nil
}

func (r *RedisLoopLock) lock() bool {
	ok, err := rdb.SetNX(r.name, r.key, time.Second*15)
	if err == nil && ok {
		signal := make(chan struct{})
		go r.loop(signal)
		<-signal
		close(signal)
		return true
	}
	if r.key == "" {
		return false
	}
	key, _, _ := rdb.Get(r.name)
	return key == r.key
}

// 这里不使用无时间限制的锁 而是使用定时去延期
// 好处: 出现程序因为上线 或者 异常崩溃 再或者 忘记了写解锁 重启服务的时候 锁会自然释放掉 用无限制时间的锁 会锁死无法自己恢复需要手动去redis删除对应的key
func (r *RedisLoopLock) loop(signal chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(fmt.Sprintf("异步锁出现异常: %s %v", r.name, err))
		}
	}()
	t := time.NewTicker(time.Second * 10)
	r.ctx, r.cancel = context.WithCancel(context.Background())
	signal <- struct{}{}
	for {
		select {
		case <-t.C:
			if ok, err := rdb.Expire(r.name, time.Second*15); err != nil || !ok {
				r.cancel()
			}
		case <-r.ctx.Done():
			t.Stop()
			r.ctx = nil
			r.cancel = nil
			return
		}
	}
}
