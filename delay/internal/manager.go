package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/lsclh/gtools/timer"
	"sync"
	"time"
)

var (
	engineSignal  = make(chan bool)
	runTaskSignal = make(chan *Task, 99999)
	taskSignal    = make(chan *Task, 99999)
)

func init() {
	go TaskManager.monitoring()
}

//当被选为master时 启动延时任务
func DelayStart() {
	engineSignal <- true
}

func DelayStop() {
	engineSignal <- false
}

//********************************************对外********************************************

// // 1.用于管理主从选择
// // 2.用于管理定时器与任务列表
type manager struct {
	handel       taskHandle       //name => task
	registry     map[string]JobFn //method => fn
	isMaster     bool
	workerId     int64
	margeRunning bool
	runWorker    sync.WaitGroup
}

// 任务管理器对象
var TaskManager = &manager{
	handel:       taskHandle{},
	registry:     map[string]JobFn{},
	isMaster:     false,
	margeRunning: false,
}

func (m *manager) monitoring() {
	for {
		select {
		case start := <-engineSignal:
			if start {
				if TaskManager.isMaster {
					return
				}
				TaskManager.isMaster = true
				// 从历史记录里面加在任务
				TaskManager.loading()
				// 启动主节点监控 用于将其他节点的任务投递到主节点
				go TaskManager.marge()

				logger.info("节点延时任务处理启动")
			} else {
				if !TaskManager.isMaster {
					return
				}
				TaskManager.isMaster = false
				// 清理本地节点任务
				TaskManager.clearLocalTask()

				close(runTaskSignal)
				for _ = range runTaskSignal {
				}
				m.runWorker.Wait()
				runTaskSignal = make(chan *Task, 99999)
				logger.info("节点延时任务处理停止")
			}
		case t := <-runTaskSignal:
			//这里只是把将要执行的任务抛进来 等待执行 此时redis里还有handle里面 还是有值的
			m.runImmediateTask(t)
		//运行过后 会清除redis和handle里的数据
		case t := <-taskSignal:
			if t.method == "" {
				TaskManager.delTask(t.name)
			} else {
				_ = TaskManager.addTask(t)
			}

		}

	}
}

func (m *manager) EmitDelTask(name string) {
	taskSignal <- &Task{
		name: name,
	}
}

func (m *manager) EmitAddTask(t *Task) {
	taskSignal <- t
}

// 合并集群其他机器抛出的任务
func (m *manager) marge() {
	if m.margeRunning {
		return
	}
	m.margeRunning = true
	defer func() {
		if err := recover(); err != nil {
			logger.error(fmt.Sprintf("合并集群任务出现异常 %v", err))
		}
		m.margeRunning = false
		fmt.Println("任务退出")
	}()

	for {
		res, err := rds.BLPop(time.Second*5, taskAdd)
		if err != nil {
			//不是主节点了就退出吧
			if !m.isMaster {
				return
			}
			// 这里有两种错误
			// 一种是没有读出数据来 属于正常的无任务投递
			if errors.Is(err, redis.Nil) {
				continue
			}
			// 另一种是redis断开连接 等待个几秒钟 需要再次重试
			time.Sleep(time.Second * 3)
			continue

		}
		if len(res) != 2 {
			continue
		}
		if !m.isMaster {
			_ = rds.RPush(taskAdd, res[1])
			return
		}

		ts := &save{}
		if err := json.Unmarshal([]byte(res[1]), ts); err != nil {
			continue
		}
		if ts.Add {
			taskSignal <- &Task{
				name:       ts.Name,
				method:     ts.Method,
				runTime:    ts.Runtime,
				params:     ts.Params,
				timeHandel: nil,
			}
		} else {
			taskSignal <- &Task{name: ts.Name}
		}
	}

}

// 删除一个task任务
func (m *manager) delTask(name string) {
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

	m.handel.Del(name)
}

// 根据一系列信息 添加一个任务到管理器
func (m *manager) addTask(t *Task) error {
	_, ok := m.registry[t.method]
	if !ok {
		_ = rds.HDel(taskList, t.name)
		return errors.New("未知的任务类型")
	}
	// 当前非运行主节点 则将任务添加到队列之中
	if !m.isMaster {
		//将任务添加到队列 主服务器加载执行
		data, _ := json.Marshal(save{
			Method:  t.method,
			Name:    t.name,
			Params:  t.params,
			Runtime: t.runTime,
			Add:     true,
		})
		err := rds.RPush(taskAdd, string(data))
		if err != nil {
			logger.error("任务异步投递失败 %s", string(data))
		}
		return err
	}

	// 删除重复的旧任务
	m.handel.Del(t.name)

	// 已到运行时间 直接执行新任务
	if t.runTime <= time.Now().Unix() {
		m.runImmediateTask(t)
		return nil
	}

	// 添加新任务
	t.timeHandel = timer.AddOnce(context.Background(), time.Duration(t.runTime-time.Now().Unix())*time.Second, func(ctx context.Context) {
		runTaskSignal <- t
	})
	m.handel.Set(t.name, t)

	return nil
}

// 加在历史未完成的任务
func (m *manager) loading() {
	res, err := rds.HGetAll(taskList)
	if err != nil {
		return
	}
	for taskId, params := range res {
		task := &save{}
		err = json.Unmarshal([]byte(params), task)
		if err != nil {
			_ = rds.HDel(taskList, taskId)
			continue
		}
		_ = m.addTask(&Task{
			name:    task.Name,
			method:  task.Method,
			runTime: task.Runtime,
			params:  task.Params,
		})

	}
}

// 到时间任务立即执行
func (m *manager) runImmediateTask(t *Task) {
	m.runWorker.Add(1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.error(
					"任务运行异常 method: %s params: %s runTime: %s",
					t.method,
					t.params,
					time.Unix(t.runTime, 0).Format("2006-01-02 15:04:05"),
				)
			}
			m.runWorker.Done()
		}()
		m.handel.Del(t.name)
		logger.info("任务运行 %s %s", t.method, t.params)
		m.registry[t.method](t.params)
	}()
}

// 清空本地挂载任务 当前节点失去主节点时 此节点上的定时器需要停止 留给其他节点运行
func (m *manager) clearLocalTask() {
	m.handel.ClearLocal()
}

const (
	taskList = "TASK_LIST"
	taskAdd  = "TASK_ADD"
)

// 注册任务
func (m *manager) AddModelToRegistry(method string, fn JobFn) {
	m.registry[method] = fn
}

type Options struct {
	Name   string //任务名称
	Method string //任务类型
	Params string //执行参数
}

type JobFn func(params string)

// // task持久化
type save struct {
	Name           string `json:"n"`
	Method         string `json:"m"`
	Runtime        int64  `json:"r"`
	Params         string `json:"p"`
	LoadTimeoutRun bool   `json:"ltr"` //如果因为主机挂掉 重新加载时已超过了执行时间 是否再次执行
	Add            bool   `json:"a"`
}
