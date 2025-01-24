package internal

import (
	"encoding/json"
	"github.com/lsclh/gtools/timer"
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
)

type Task struct {
	name    string //任务名称
	method  string //任务类型
	runTime int64  //执行时间
	params  string //执行参数
	//依赖
	timeHandel *timer.Entry
}

// ********************************************对外********************************************
func NewTask(opt *Options) *Task {
	t := &Task{
		name:   opt.Name,
		method: opt.Method,
		params: opt.Params,
	}
	if t.name == "" {
		t.name = uuid.NewV4().String()
	}
	return t
}

// 添加任务 具体的运行时间
func (t *Task) AddForTime(timestamp int64) {
	t.runTime = timestamp
	TaskManager.EmitAddTask(t)
}

// 添加任务 当前之后多长时间运行
func (t *Task) AddForAfter(after time.Duration) {
	t.runTime = time.Now().Add(after).Unix()
	TaskManager.EmitAddTask(t)
}

type taskHandle struct {
	di sync.Map
}

func (h *taskHandle) Get(name string) *Task {
	t, ok := h.di.Load(name)
	if !ok {
		return nil
	}
	return t.(*Task)
}

func (h *taskHandle) Set(name string, t *Task) {
	h.di.Store(name, t)
	// 持久化存储
	d, _ := json.Marshal(save{
		Name:    t.name,
		Method:  t.method,
		Params:  t.params,
		Runtime: t.runTime,
		Add:     true,
	})
	_ = rds.HSet(taskList, t.name, string(d))
}

func (h *taskHandle) Del(name string) {
	_ = rds.HDel(taskList, name)
	t, loaded := h.di.LoadAndDelete(name)
	if loaded && t != nil {
		t.(*Task).timeHandel.Stop()
	}
}

func (h *taskHandle) ClearLocal() {
	h.di.Range(func(key, value interface{}) bool {
		h.di.Delete(key)
		value.(*Task).timeHandel.Stop()
		return true
	})
}

func (h *taskHandle) Clear() {
	h.di.Range(func(key, value interface{}) bool {
		t := value.(*Task)
		t.timeHandel.Stop()
		h.di.Delete(key)
		_ = rds.HDel(taskList, t.name)
		return true
	})
}
