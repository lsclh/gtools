package rLock

import "time"

// 锁模版
type Mutex interface {
	//TryLock 尝试加锁 成功返回true 失败返回false
	TryLock() bool
	//Lock 尝试加锁 成功往下执行 失败阻塞自旋
	Lock()
	//Unlock 解锁
	Unlock() bool
}

// redis操作接口
type Rds interface {
	//return true成功(说明没值写入成功) false=失败(说明有值写入失败)
	SetNX(key string, value interface{}, expiration time.Duration) (bool, error)
	//return false失败(说明没值无需删除) true成功(说明有值并删除了)
	Del(key string) (bool, error)
	//return 值 true存在key,false不存在key
	Get(key string) (string, bool, error)
	//return true成功(说明有值并更新了过期时间) false失败(说明没值无法更新过期时间)
	Expire(key string, expiration time.Duration) (bool, error)
}

var rdb Rds = nil

func SetRds(r Rds) {
	if rdb == nil {
		rdb = r
	}
}
