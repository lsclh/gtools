package master

import (
	"github.com/lsclh/gtools/rLock"
	uuid "github.com/satori/go.uuid"
	"time"
)

const masterKey = "MASTER_LOCK"

type m struct {
	isMaster     bool
	lock         rLock.Mutex
	startHandler []func()
	stopHandler  []func()
}

var manager = &m{
	isMaster:     false,
	startHandler: []func(){},
	stopHandler:  []func(){},
}

// 启动master节点判断
func Start() {
	manager.lock = rLock.NewRedisTimeLocker(masterKey, uuid.NewV4().String(), time.Second*11)
	//从服务器 10秒检测一次
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.error("master节点选择出现异常 %v", err)
			}
		}()
		t := time.NewTicker(time.Second * 10) //从服务器 10秒检测一次
		for {
			if manager.lock.TryLock() {
				if !manager.isMaster {
					logger.info("当前节点已成为master")
					manager.isMaster = true
					for _, fn := range manager.startHandler {
						go func(fc func()) {
							defer func() {
								if err := recover(); err != nil {
									logger.error("通知监控任务出现异常 %v", err)
								}
							}()
							fc()
						}(fn)
					}
				}
			} else if manager.isMaster {
				logger.info("当前节点已退出master")
				manager.isMaster = false
				for _, fn := range manager.stopHandler {
					go func(fc func()) {
						defer func() {
							if err := recover(); err != nil {
								logger.error("通知监控任务出现异常 %v", err)
							}
						}()
						fc()
					}(fn)
				}
			}
			<-t.C
		}
	}()
}

func IsMaster() bool {
	return manager.isMaster
}

// 当节点成为master时进行调用事件
func MStartEvent(fn func()) {
	manager.startHandler = append(manager.startHandler, fn)
}

func MStopEvent(fn func()) {
	manager.stopHandler = append(manager.stopHandler, fn)
}
