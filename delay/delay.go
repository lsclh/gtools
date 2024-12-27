package delay

import (
	"github.com/lsclh/gtools/delay/internal"
	"github.com/spf13/cast"
)

// 注册对应任务的执行模型
func RegisterMethod(method string, fn internal.JobFn) {
	internal.TaskManager.AddModelToRegistry(method, fn)
}

func SetLog(l internal.Log) {
	internal.SetLog(l)
}

func SetRds(r internal.Rds) {
	internal.SetRds(r)
}

func DelayStart() {
	internal.DelayStart()
}

// 退出程序是等待当前的运行完成
func WaitStopTask() {
	internal.TaskManager.WaitStopTask()
}

// 获取一个任务对象 并创建添加任务
func New(method string, opts ...options) *internal.Task {
	opt := &internal.Options{
		Method:         method,
		Name:           "",
		Params:         "",
		LoadTimeoutRun: true,
	}
	for _, fn := range opts {
		fn(opt)
	}

	return internal.NewTask(opt)
}

// 删除一个任务
func Del(name string) {
	internal.TaskManager.DeleteTask(name)
}

// **********************************************注册参数**********************************************************
type options func(e *internal.Options)

// 如果后续需要删除任务则需要注册此任务名称 用于后续指定名称删除使用
func WithName(name string) options {
	return func(e *internal.Options) {
		e.Name = name
	}
}

// 如果任务携带参数 可以用此函数添加参数
func WithParams(params any) options {
	return func(e *internal.Options) {
		//根据params类型 转化为string 在存放上去

		e.Params = cast.ToString(params)
	}
}

// 如果服务因为异常停止
// 过了一段期间才启动起来 此时次任务过了预定执行时间 是否还要执行 还是丢弃掉
// 默认执行
func WithLoadTimeoutRun(run bool) options {
	return func(e *internal.Options) {
		e.LoadTimeoutRun = run
	}
}

//**********************************************注册参数**********************************************************
