package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

func Start(daemonName, workerPath, exeName string, args ...string) {
	if args == nil {
		args = make([]string, 0)
	}
	go daemon(daemonName, workerPath, exeName, args...)
}

func Stop(daemonName string) {
	pro, ok := engine.GetProcess(daemonName)
	if ok {
		pro.runMode = 1
		if pro.cmd.Process != nil {
			//优先尝试主动退出 如果退出不来则进行进程强制kill
			err := pro.cmd.Process.Signal(syscall.SIGTERM)
			if err == nil {
				time.Sleep(time.Second * 5)
			}
			if pro.cmd.ProcessState == nil {
				pro.cmd.Process.Kill()
			}

		}
	}
}

func Restart(daemonName string) {
	pro, ok := engine.GetProcess(daemonName)
	if ok {

		pro.cmd.Process.Kill()
	}
}

func daemon(daemonName, path string, exeName string, args ...string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(fmt.Sprintf("守护进程异常 %s %v", daemonName, err))
			engine.DelProcess(daemonName)
		}
	}()
	for {
		pro, ok := engine.GetProcess(daemonName)
		if ok {
			if pro.runMode == 1 {
				fmt.Println("目标程序停止运行...")
				engine.DelProcess(daemonName)
				return
			}
			fmt.Println(fmt.Sprintf("目标程序已退出，准备重启..."))
		}

		targetCmd := exec.Command(path+"/"+exeName, args...)
		targetCmd.Stdin = os.Stdin
		targetCmd.Stdout = os.Stdout
		targetCmd.Stderr = os.Stderr
		err := targetCmd.Start()
		if err != nil {
			fmt.Println(fmt.Sprintf("启动目标程序失败: %s %s", daemonName, err.Error()))
			engine.DelProcess(daemonName)
			break
		}
		fmt.Println("run oasis bydeamon,pid=", targetCmd.Process.Pid, ",ppid=", os.Getpid(), "args=", args, "time=", time.Now())

		// 等待目标程序退出，同时捕获其退出状态
		waitStatus := make(chan struct{})
		go func() {
			err := targetCmd.Wait()
			waitStatus <- struct{}{}
			if err != nil {
				fmt.Println(fmt.Sprintf("目标程序意外退出: %s", err.Error()))
			}
		}()

		if ok {
			pro.cmd = targetCmd
			pro.runTime = time.Now().Unix()
			pro.restartNum += 1
		} else {
			engine.SetProcess(daemonName, &process{
				cmd:            targetCmd,
				runTime:        time.Now().Unix(),
				restartNum:     0,
				workingDirPath: path,
				exeName:        exeName,
				args:           args,
				daemonName:     daemonName,
			})
		}
		<-waitStatus
		close(waitStatus)
	}
}

var engine = &storage{}

type storage struct {
	mp sync.Map
}

type process struct {
	cmd            *exec.Cmd
	runMode        int   //0=运行 1=停止
	runTime        int64 //运行时间
	restartNum     int64 //重启次数
	exeName        string
	daemonName     string
	workingDirPath string
	args           []string
}

func (e *storage) GetProcess(daemonName string) (*process, bool) {
	v, ok := e.mp.Load(daemonName)
	if !ok {
		return nil, false
	}
	return v.(*process), true
}

func (e *storage) SetProcess(daemonName string, p *process) {
	e.mp.Store(daemonName, p)
}

func (e *storage) DelProcess(daemonName string) {
	e.mp.Delete(daemonName)
}
