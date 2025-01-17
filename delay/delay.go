package delay

import (
	"github.com/lsclh/gtools/delay/internal"
	"github.com/spf13/cast"
)

var Engine = &engine{}

// 获取一个任务对象 并创建添加任务
var WorkerOptions = &workerOptions{}

var Method = &method{}

// New 获取一个任务对象
// @method 任务类型 必填
func New(method string, workerOptions *workerOptions) *internal.Task {
	if method == "" {
		return nil
	}
	return internal.NewTask(&internal.Options{
		Name:           workerOptions.cnf.name,
		Method:         method,
		Params:         workerOptions.cnf.params,
		LoadTimeoutRun: workerOptions.cnf.loadTimeoutRun,
	})
}

//

// 删除一个任务
func Del(name string) {
	internal.TaskManager.DeleteTask(name)
}

// **********************************************注册参数**********************************************************

type method struct{}

//注册对应任务类型的执行模型
func (*method) Register(method string, fn internal.JobFn) {
	internal.TaskManager.AddModelToRegistry(method, fn)
}

type workerOptions struct {
	cnf *cnf
}

func (w *workerOptions) clone() *workerOptions {
	if w.cnf != nil {
		return w
	}
	opt := &workerOptions{
		cnf: &cnf{
			name:           "",
			params:         "",
			loadTimeoutRun: false,
		},
	}
	opt.cnf.opt = opt
	return opt
}

// WithName 如果后续需要删除任务则需要注册此任务名称 用于后续指定名称删除使用
func (w *workerOptions) WithName(name string) *workerOptions {
	return w.clone().cnf.saveName(name)
}

// WithParams 如果任务携带参数 可以用此函数添加参数
func (w *workerOptions) WithParams(params any) *workerOptions {
	return w.clone().cnf.saveParams(params)
}

// WithLoadTimeoutRun 如果服务因为异常停止
// 过了一段期间才启动起来 此时次任务过了预定执行时间 是否还要执行 还是丢弃掉
// 默认执行
func (w *workerOptions) WithLoadTimeoutRun(run bool) *workerOptions {
	return w.clone().cnf.saveLoadTimeoutRun(run)
}

type cnf struct {
	opt            *workerOptions
	name           string //任务名称
	params         string //执行参数
	loadTimeoutRun bool   //如果因为主机挂掉 重新加载时已超过了执行时间 是否再次执行
}

func (c *cnf) saveName(name string) *workerOptions {
	c.name = name
	return c.opt
}

func (c *cnf) saveParams(params any) *workerOptions {
	c.params = cast.ToString(params)
	return c.opt
}
func (c *cnf) saveLoadTimeoutRun(run bool) *workerOptions {
	c.loadTimeoutRun = run
	return c.opt
}

//**********************************************注册参数**********************************************************

type engine struct{}

func (f *engine) Start() {
	internal.DelayStart()
}

func (f *engine) Stop() {
	internal.DelayStop()
}

// 退出程序是等待当前的运行完成
func (f *engine) WaitStop() {
	internal.TaskManager.WaitStopTask()
}

// RegLog 注册日志组件
func (f *engine) RegLog(l internal.Log) *engine {
	internal.SetLog(l)
	return f
}

// RegRds 注册redis组件 必须
func (f *engine) RegRds(r internal.Rds) *engine {
	internal.SetRds(r)
	return f
}
