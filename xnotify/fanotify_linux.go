package xnotify

import (
	"fmt"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"syscall"
	"time"
)

const faInitFlags = FAN_CLOEXEC | FAN_CLASS_CONTENT | FAN_UNLIMITED_MARKS
const faMarkAddFlags = FAN_MARK_ADD
const faMarkDelFlags = FAN_MARK_REMOVE
const faMarkMask = FAN_CLOSE_WRITE | FAN_MODIFY
const faMarkMaskDir = FAN_ONDIR | FAN_EVENT_ON_CHILD

type FaNotify struct {
	fNotify
	bEnabled      bool
	configPerm    bool
	agentPid      int
	dirMonitorMap map[string]uint64
	fa            *NotifyFD
	endChan       chan bool
	EventProcinfo chan ProcInfo
}

func NewFaNotify(endFaChan chan bool) (*FaNotify, error) {
	fa, err := initialize(faInitFlags, os.O_RDONLY|syscall.O_LARGEFILE)
	if err != nil {
		return nil, err
	}

	fant := FaNotify{
		bEnabled:      true,
		configPerm:    false,
		agentPid:      os.Getegid(),
		fa:            fa,
		dirMonitorMap: make(map[string]uint64),
		endChan:       endFaChan,
		EventProcinfo: make(chan ProcInfo),
	}

	fant.configPerm = fant.checkConfigPerm()
	return &fant, nil
}

// 预检查，看当前进程是否有权限操作fanotify的事件
func (fn *FaNotify) checkConfigPerm() bool {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "fan_test")
	if err != nil {
		fmt.Printf("Create temp directory fail error:%v\n", err)
		return false
	}
	defer os.RemoveAll(tmpDir)

	//监控/tmp目录的文件打开事件
	mask := uint64(FAN_OPEN_PERM | FAN_ONDIR)
	err = fn.fa.Mark(faMarkAddFlags, mask, 0, tmpDir)
	if err != nil {
		fmt.Printf("not supported error:%v\n", err)
		return false
	}

	//删除/tmp目录的文件打开事件的监控
	err = fn.fa.Mark(faMarkDelFlags, mask, 0, tmpDir)
	if err != nil {
		fmt.Printf("delete mark fail error:%v\n", err)
	}
	return true
}

func (fn *FaNotify) AddMonitorFile(path string, mask uint64) error {

	if _, ok := fn.dirMonitorMap[path]; ok {
		return fmt.Errorf("already exisit path: %v\n", path)
	}

	fn.dirMonitorMap[path] = mask

	return fn.fa.Mark(faMarkAddFlags, mask, unix.AT_FDCWD, path)
}

func (fn *FaNotify) monitorExit() {
	if fn.fa != nil {
		fn.fa.Close()
	}

	if fn.endChan != nil {
		fn.endChan <- true
	}
}

func (fn *FaNotify) Close() {
	if !fn.bEnabled {
		return
	}

	fn.mux.Lock()
	defer fn.mux.Unlock()

	fn.bEnabled = false

	go func() {
		time.Sleep(5 * time.Second)
		fn.monitorExit()
		fmt.Println("exit FaNotify......")
	}()
}

func (fn *FaNotify) RemoveMonitor() {
	for path, mask := range fn.dirMonitorMap {
		if err := fn.fa.Mark(faMarkDelFlags, mask, 0, path); err != nil {
			fmt.Printf("[FaNotify] removeMonitor path(%s) failed: %v\n", path, err)
		}
	}
}

func (fn *FaNotify) MonitorFileEvents() {
	waitCnt := 0
	pfd := make([]unix.PollFd, 1)
	pfd[0].Fd = fn.fa.GetFd()
	pfd[0].Events = unix.POLLIN

	fmt.Println("[FaNotify MonitorFileEvents] Start Monitor FaNotify............")

	for {
		n, err := unix.Poll(pfd, 1000) // wait 1 sec
		if err != nil && err != unix.EINTR {
			fmt.Printf("[FaNotify MonitorFileEvents] poll error:%v\n", err)
			break
		}
		if n <= 0 {
			if n == 0 && !fn.bEnabled { // timeout at exit stage
				waitCnt += 1
				if waitCnt > 1 { // two chances
					break
				}
			}
			continue
		}

		if (pfd[0].Revents & unix.POLLIN) != 0 {
			if err := fn.handleEvents(); err != nil {
				fmt.Printf("[FaNotify MonitorFileEvents] handle error:%v\n", err)
				break
			}
			waitCnt = 0
		}
	}

	fn.monitorExit()
	fmt.Printf("[FaNotify MonitorFileEvents] exit")
}

func (fn *FaNotify) handleEvents() error {
	if events, err := fn.fa.GetEvents(); err == nil {
		for _, ev := range events {
			pid := int(ev.Pid)
			fd := int(ev.File.Fd())
			fmask := uint64(ev.Mask)

			fmt.Printf("[FaNotify handleEvents] get event,pid: %v, fmask:%v, fd:%+v\n", pid, fmt.Sprintf("0x%08x", fmask), fd)
			fmt.Printf("[FaNotify handleEvents] EventMetadata struct:%+v\n", ev)
			fn.convert2ProcInfo(pid)
			perm := (fmask & (FAN_OPEN_PERM | FAN_ACCESS_PERM)) > 0

			if perm {
				fmt.Printf("[FaNotify handleEvents] get FAN_OPEN_PERM|FAN_ACCESS_PERM event, start write allow or deny")
				fn.fa.Response(ev, true)
			}
			ev.File.Close()

		}
	}
	return nil
}

func (fn *FaNotify) convert2ProcInfo(pid int) {
	pInfo := ProcInfo{}

	pInfo.Pid = pid
	pInfo.Name, pInfo.PPid, _ = GetProcessName(pid)
	pInfo.Path, _ = GetFilePath(pid)
	pInfo.PPath, _ = GetFilePath(pInfo.PPid)
	pInfo.Cmds, _ = ReadCmdLine(pid)

	fn.EventProcinfo <- pInfo
}
