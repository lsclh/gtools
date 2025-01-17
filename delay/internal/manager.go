package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/lsclh/gtools/timer"
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
)

// ********************************************对外********************************************
func NewTask(opt *Options) *Task {
	t := &Task{
		name:           opt.Name,
		method:         opt.Method,
		params:         opt.Params,
		loadTimeoutRun: opt.LoadTimeoutRun,
	}
	if t.name == "" {
		t.name = uuid.NewV4().String()
	}
	return t
}

// 添加任务 具体的运行时间
func (t *Task) AddForTime(timestamp int64) {
	t.runTime = timestamp
	_ = TaskManager.createTask(t)
}

// 添加任务 当前之后多长时间运行
func (t *Task) AddForAfter(after time.Duration) {
	t.runTime = time.Now().Add(after).Unix()
	_ = TaskManager.createTask(t)
}

//********************************************对外********************************************

const (
	taskList = "TASK_LIST"
	taskAdd  = "TASK_ADD"
)

type Options struct {
	Name           string //任务名称
	Method         string //任务类型
	Params         string //执行参数
	LoadTimeoutRun bool   //如果因为主机挂掉 重新加载时已超过了执行时间 是否再次执行
}

type JobFn func(params string)

type Task struct {
	name           string //任务名称
	method         string //任务类型
	runTime        int64  //执行时间
	params         string //执行参数
	loadTimeoutRun bool   //如果因为主机挂掉 重新加载时已超过了执行时间 是否再次执行
	//依赖
	timeHandel *timer.Entry
}

// // 1.用于管理主从选择
// // 2.用于管理定时器与任务列表
type manager struct {
	localLock sync.Mutex
	handel    map[string]Task  //name => task
	registry  map[string]JobFn //method => fn
	wait      sync.WaitGroup
	isMaster  bool
	workerId  int64
}

// // task持久化
type save struct {
	Name           string `json:"n"`
	Method         string `json:"m"`
	Runtime        int64  `json:"r"`
	Params         string `json:"p"`
	LoadTimeoutRun bool   `json:"ltr"` //如果因为主机挂掉 重新加载时已超过了执行时间 是否再次执行
	Add            bool   `json:"a"`
}

// 任务管理器对象
var TaskManager = &manager{
	handel:   map[string]Task{},
	registry: map[string]JobFn{},
	isMaster: false,
}

//当被选为master时 启动延时任务
func DelayStart() {
	if TaskManager.isMaster {
		return
	}
	TaskManager.workerId++
	TaskManager.isMaster = true
	TaskManager.Setup()
	go TaskManager.MargeTask()
	logger.info("节点延时任务处理启动")
}

func DelayStop() {
	if !TaskManager.isMaster {
		return
	}
	TaskManager.isMaster = false
	logger.info("节点延时任务处理停止")
}

// 注册任务
func (m *manager) AddModelToRegistry(method string, fn JobFn) {
	m.registry[method] = fn
}

func (m *manager) Setup() {

	res, err := rds.HGetAll(taskList)
	if err != nil {
		return
	}
	for taskId, params := range res {
		task := &save{}
		err := json.Unmarshal([]byte(params), task)
		if err != nil {
			rds.HDel(taskList, taskId)
			continue
		}
		to := m.initTaskForSave(task)
		if to != nil {
			rds.HDel(taskList, taskId)
		}
	}
}

// 根据缓存记录初始化任务
func (m *manager) initTaskForSave(ts *save) error {
	return m.createTask(&Task{
		name:           ts.Name,
		method:         ts.Method,
		runTime:        ts.Runtime,
		params:         ts.Params,
		loadTimeoutRun: ts.LoadTimeoutRun,
	})
}

// 到时间任务立即执行
func (m *manager) runImmediateTask(t *Task, fn JobFn) {
	m.wait.Add(1)
	go func() {
		defer func() {
			m.wait.Done()
			if err := recover(); err != nil {
				logger.error(
					"任务运行异常 method: %s params: %s runTime: %s",
					t.method,
					t.params,
					time.Unix(t.runTime, 0).Format("2006-01-02 15:04:05"),
				)
			}
		}()
		fn(t.params)
	}()
}

// 创建一个任务一个任务
func (m *manager) createTask(t *Task) error {
	fn, ok := m.registry[t.method]
	if !ok {
		return errors.New("未知的任务类型")
	}
	if !m.isMaster {
		//将任务添加到队列 主服务器加载执行
		data, _ := json.Marshal(save{
			Method:         t.method,
			Name:           t.name,
			Params:         t.params,
			Runtime:        t.runTime,
			LoadTimeoutRun: t.loadTimeoutRun,
			Add:            true,
		})
		_ = rds.RPush(taskAdd, string(data))
		return nil
	}

	if t.runTime <= time.Now().Unix() {
		if t.loadTimeoutRun {
			m.runImmediateTask(t, fn)
		}
		return errors.New("已运行完成")
	}

	m.localLock.Lock()
	defer m.localLock.Unlock()

	if job, taskDuplication := m.handel[t.name]; taskDuplication {
		job.timeHandel.Stop()
		logger.info("旧任务替换停止 method: %s params: %s runTime: %s: ", job.method, job.params, time.Unix(job.runTime, 0).Format("2006-01-02 15:04:05"))
	}

	t.timeHandel = timer.AddOnce(
		context.Background(),
		time.Duration(t.runTime-time.Now().Unix())*time.Second,
		func(ctx context.Context) {
			m.wait.Add(1)
			defer m.wait.Done()
			TaskManager.clearTask(t.name)
			fn(t.params)
		})
	m.handel[t.name] = *t

	d, _ := json.Marshal(save{
		Name:           t.name,
		Method:         t.method,
		Params:         t.params,
		Runtime:        t.runTime,
		LoadTimeoutRun: t.loadTimeoutRun,
		Add:            true,
	})
	rds.HSet(taskList, t.name, string(d))
	return nil
}

// 删除一个task任务
func (m *manager) DeleteTask(name string) {
	if name == "" {
		return
	}
	if !m.isMaster {
		data, _ := json.Marshal(save{
			Name: name,
			Add:  false,
		})
		_ = rds.RPush(taskAdd, string(data))
		return
	}
	m.localLock.Lock()
	defer m.localLock.Unlock()
	rds.HDel(taskList, name)
	t, ok := m.handel[name]
	if !ok {
		return
	}
	t.timeHandel.Stop()
	delete(m.handel, name)
}

// 等待当前运行的task执行完成
func (m *manager) WaitStopTask() {
	m.wait.Wait()
}

// 合并集群其他机器抛出的任务
func (m *manager) MargeTask() {
	workerId := m.workerId
	defer func() {
		if err := recover(); err != nil {
			logger.error(fmt.Sprintf("合并集群任务出现异常 %v", err))
		}
	}()
	for {

		res, err := rds.BLPop(time.Second*5, taskAdd)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				if !m.isMaster || workerId != m.workerId {
					return
				}
				time.Sleep(time.Second * 3)
			}
			continue
		}
		if len(res) != 2 {
			continue
		}
		if !m.isMaster || workerId != m.workerId {
			_ = rds.RPush(taskAdd, res[1])
			return
		}

		ts := &save{}
		if err := json.Unmarshal([]byte(res[1]), ts); err != nil {
			continue
		}
		if ts.Add {
			_ = m.createTask(&Task{
				name:           ts.Name,
				method:         ts.Method,
				runTime:        ts.Runtime,
				params:         ts.Params,
				loadTimeoutRun: ts.LoadTimeoutRun,
				timeHandel:     nil,
			})
		} else {
			m.DeleteTask(ts.Name)
		}
	}

}

func (m *manager) clearTask(name string) {
	m.localLock.Lock()
	defer m.localLock.Unlock()
	rds.HDel(taskList, name)
	if val, ok := m.handel[name]; ok {
		val.timeHandel.Stop()
		delete(m.handel, name)
	}
}
