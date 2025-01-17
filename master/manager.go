package master

import (
	"github.com/lsclh/gtools/rLock"
	uuid "github.com/satori/go.uuid"
	"time"
)

const masterKey = "MASTER_LOCK"

type engine struct {
	isMaster     bool
	lock         rLock.Mutex
	startHandler []func()
	stopHandler  []func()
	cl           chan struct{}
	start        bool
}

var Engine = &engine{
	isMaster:     false,
	startHandler: []func(){},
	stopHandler:  []func(){},
}

// 启动master节点判断
func (e *engine) Start() {
	if e.start {
		return
	}
	e.start = true
	e.lock = rLock.Factory.NewRedisTimeLocker(masterKey, uuid.NewV4().String(), time.Second*11)
	//从服务器 10秒检测一次
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.error("master节点选择出现异常 %v", err)
			}
			e.lock.Unlock()
		}()
		t := time.NewTicker(time.Second * 10) //从服务器 10秒检测一次
		e.cl = make(chan struct{})
		for {
			if e.lock.TryLock() {
				if !e.isMaster {
					logger.info("当前节点已成为master")
					e.isMaster = true
					for _, fn := range e.startHandler {
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
			} else if e.isMaster {
				logger.info("当前节点已退出master")
				e.isMaster = false
				for _, fn := range e.stopHandler {
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
			select {
			case <-t.C:
			case <-e.cl:
				t.Stop()
				close(e.cl)
				e.start = false
				return
			}
		}
	}()
}

func (e *engine) Stop() {
	e.cl <- struct{}{}
	time.Sleep(time.Second)
}

func (e *engine) IsMaster() bool {
	return e.isMaster
}

func (e *engine) RegLog(lg log) *engine {
	l = lg
	return e
}

func (e *engine) RegRds(r rds) *engine {
	rLock.Engine.RegRds(r)
	return e
}

type event struct{}

var Event = &event{}

// 当节点成为master时进行调用事件
func (e *event) IsMaster(fn func()) {
	Engine.startHandler = append(Engine.startHandler, fn)
}

func (*event) RmMaster(fn func()) {
	Engine.stopHandler = append(Engine.stopHandler, fn)
}
