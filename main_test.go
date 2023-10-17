package common_library

import (
	"fmt"
	"github.com/hugoww/common-library/utils"
	"github.com/hugoww/common-library/xnotify"
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func TestMainXnotify(t *testing.T) {
	var end chan bool = make(chan bool)
	fa, err := xnotify.NewFaNotify(end)
	if err != nil {
		fmt.Println("NewFaNotify error:", err)
		os.Exit(1)
	}

	// 只有在对文件注册的监控事件类型为"FAN_ACCESS_PERM|FAN_OPEN_PERM"类型的时候
	// 需要把对这些事件类型允许,还是拒绝的结果写回到fanotify的文件描述符(即写回内核中)，从而判断进程是否有权限对文件的操作；
	err = fa.AddMonitorFile("/tmp/dir1", xnotify.FAN_ACCESS|xnotify.FAN_CLOSE_WRITE|xnotify.FAN_OPEN|xnotify.FAN_MODIFY)
	if err != nil {
		fmt.Println("NewFaNotify add monitor file error:", err)
	}

	go fa.MonitorFileEvents()
	go func() {
		var pInfo xnotify.ProcInfo
		pInfo = <-fa.EventProcinfo
		fmt.Printf("Get Fanotify Event info about process info: %+v\n", pInfo)
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-end:
		fa.RemoveMonitor()
		fa.Close()
	case <-quit:
		fa.RemoveMonitor()
		fa.Close()
	}

}

func IPNet2Subnet(ipnet *net.IPNet) *net.IPNet {
	return &net.IPNet{
		IP:   ipnet.IP.Mask(ipnet.Mask),
		Mask: ipnet.Mask,
	}
}

func TestSubnetSet(t *testing.T) {
	subnets := utils.NewSet()
	ipnet1 := net.IPNet{IP: net.ParseIP("1.2.3.4"), Mask: net.CIDRMask(16, 32)}
	ipnet2 := net.IPNet{IP: net.ParseIP("1.2.4.5"), Mask: net.CIDRMask(16, 32)}
	str1 := IPNet2Subnet(&ipnet1).String()
	str2 := IPNet2Subnet(&ipnet2).String()
	subnets.Add(str1)
	subnets.Add(str2)

	if subnets.Cardinality() != 1 {
		t.Errorf("Wrong subnet set size: %v\n", subnets.Cardinality())
	}

	ip := net.ParseIP("1.2.5.6")
	for str := range subnets.Iter() {
		if _, subnet, err := net.ParseCIDR(str.(string)); err != nil {
			t.Errorf("Subnet set error: %v\n", err.Error())
		} else if !subnet.Contains(ip) {
			t.Errorf("Subnet not contain IP: %v %v\n", subnets, ip.String())
		}
	}
}
