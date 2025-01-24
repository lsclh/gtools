package master

import (
	"github.com/lsclh/gtools/rLock"
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
)

func init() {
	Engine.lock = rLock.Factory.NewRedisTimeLocker(masterKey, uuid.NewV4().String(), time.Second*11)
	// 监控启动停止指令
	go Engine.monitoring()
}

const masterKey = "MASTER_LOCK"

var (
	startSignal    = make(chan struct{})
	stopSignal     = make(chan struct{})
	stoppingSignal = make(chan struct{})
)

type engine struct {
	isMaster     bool
	lock         rLock.Mutex
	startHandler []func()
	stopHandler  []func()
	close        chan struct{}
	start        bool
	waitClose    sync.WaitGroup
}

var Engine = &engine{
	isMaster:     false,
	startHandler: []func(){},
	stopHandler:  []func(){},
	close:        make(chan struct{}),
}

// 启动 master 判断任务
func (e *engine) Start() {

	startSignal <- struct{}{}
}

// 停止 master 判断任务
func (e *engine) Stop() {

	stopSignal <- struct{}{}
	<-stoppingSignal
}

func (e *engine) RegLog(lg log) *engine {
	l = lg
	return e
}

func (e *engine) RegRds(r rds) *engine {
	rLock.Engine.RegRds(r)
	return e
}

func (e *engine) IsMaster() bool {
	return e.isMaster
}

// 监控启动与停止指令
func (e *engine) monitoring() {
	for {
		select {
		case <-startSignal:
			if e.start {
				return
			}
			e.start = true
			go e.scrambleForMaster()
		case <-stopSignal:
			if !e.start {
				return
			}
			e.start = false
			e.close <- struct{}{}
			e.waitClose.Wait()
			stoppingSignal <- struct{}{}
		}
	}
}

// 通知所有的监控任务
func (e *engine) emitStopMaster() {
	logger.info("当前节点已退出master")
	e.isMaster = false
	var wait sync.WaitGroup
	for _, fn := range e.stopHandler {
		wait.Add(1)
		go func(fc func()) {
			defer func() {
				if err := recover(); err != nil {
					logger.error("通知监控任务出现异常 %v", err)
				}
				wait.Done()
			}()
			fc()
		}(fn)
	}
	wait.Wait()
}

// 通知所有的监控任务
func (e *engine) emitStartMaster() {
	logger.info("当前节点已成为master")
	e.isMaster = true
	var wait sync.WaitGroup
	for _, fn := range e.startHandler {
		wait.Add(1)
		go func(fc func()) {
			defer func() {
				if err := recover(); err != nil {
					logger.error("通知监控任务出现异常 %v", err)
				}
				wait.Done()
			}()
			fc()
		}(fn)
	}
	wait.Wait()
}

// 选举master
func (e *engine) scrambleForMaster() {
	e.waitClose.Add(1)
	defer func() {
		if err := recover(); err != nil {
			logger.error("master节点选择出现异常 %v", err)
		}
		e.lock.Unlock()
		e.waitClose.Done()
	}()
	t := time.NewTicker(time.Second * 10) //从服务器 10秒检测一次
	for {
		if e.lock.TryLock() {
			if !e.isMaster {
				e.emitStartMaster()
			}
		} else if e.isMaster {
			e.emitStopMaster()
		}
		select {
		case <-t.C:
		case <-e.close:
			t.Stop()
			if e.isMaster {
				e.emitStopMaster()
			}
			return
		}
	}
}

type event struct{}

var Event = &event{}

// 选中为主节点事件
func (event) ChooseMasterNode(fn func()) {
	Engine.startHandler = append(Engine.startHandler, fn)
}

//
func (event) LoseMasterNode(fn func()) {
	Engine.stopHandler = append(Engine.stopHandler, fn)
}
