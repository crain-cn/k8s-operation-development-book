package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"golang.org/x/text/encoding/simplifiedchinese"
	"k8s.io/klog/v2"

	"os/exec"
	"strings"
)

const (
	CalicoFile   = "/etc/cni/net.d/10-calico.conflist"
	FlannelFile  = "/etc/cni/net.d/10-flannel.conflist"

	// bpftrace 时间
	Intervel = "500"
)

var cniType = ""

// 删除 ifb3 net 和  qdisc
var tcDelete = []string{
	"tc qdisc del dev ifb3 root",
	"ip link del ifb3 type ifb",
}

type NetSuppressor struct {
}


func NewNetSuppress() *NetSuppressor {
	return &NetSuppressor{}
}

func execShell(commandStr string) error {
	if err := exec.Command("/bin/bash", "-c", commandStr).Run(); err != nil {
		klog.Warningf(err.Error())
		klog.Warningf("shell -- " + commandStr + " -- run err")
		return err
	}
	return nil
}

// 创建Pod的veth关联备注ifb3
func (net *NetSuppressor) createIfbConnectVeth(vethName string) {
	tcAddVethQdisc := fmt.Sprintf("tc qdisc add dev %s ingress handle ffff: ", vethName)
	if err := execShell(tcAddVethQdisc); err != nil {
		return
	}
	tcAddVethFilter := "tc filter add dev " + vethName + " parent ffff: protocol ip u32 match u32 0 0 flowid 1:1 action mirred egress redirect dev ifb3 "
	if err := execShell(tcAddVethFilter); err != nil {
		return
	}
}

// 删除Pod对应的veth名称，filter侧罗和qdisc
func (net *NetSuppressor) deleteIfbConnectVeth(vethName string) {
	tcDeleteVethFilter := fmt.Sprintf("tc filter delete dev %s parent ffff: ", vethName)
	klog.Infof(vethName)
	execShell(tcDeleteVethFilter)
	tcDeleteVethQdisc := "tc qdisc delete dev " + vethName + " ingress handle ffff: "
	execShell(tcDeleteVethQdisc)
}

// 判断CNI类型
func (net *NetSuppressor) setCniType() {
	if _, err := os.Stat(CalicoFile); !os.IsNotExist(err) {
		cniType = "Calico"
		return
	}
	if _, err := os.Stat(FlannelFile); !os.IsNotExist(err) {
		cniType = "Flannel"
		return
	}
	klog.Warningf("file dont exit dont know cni type")
}

// 通过podIP获取veth名称
func (net *NetSuppressor) getVethName(podIP string) (string, error) {
	if cniType == "" {
		klog.Warningf("can not get cniType")
		return "", errors.New("can not get cniType")
	}
	// CNI 为Calico的获取方式
	if cniType == "Calico" {
		vethName, err := exec.Command("/bin/bash", "-c", "route | grep "+podIP+" | awk ' {print $NF }'").Output()
		if err != nil {
			klog.Warningf("cannot find vethName in Calico podIP is " + string(podIP))
			return "", err
		}
		return strings.Replace(string(vethName), "\n", "", -1), nil
	}
	// CNI 为Flannel的获取方式
	if cniType == "Flannel" {
		arpAddr, err := exec.Command("/bin/bash", "-c", "arp -e | grep "+podIP+" | awk ' {print $3}'").Output()
		if err != nil {
			klog.Warningf("cannot find vethName in flannel podIP is " + string(podIP))
			return "", err
		}
		vethName, err := exec.Command("/bin/bash", "-c", "bridge fdb show | grep "+strings.Replace(string(arpAddr), "\n", "", -1)+" | awk '{print $3}'").Output()
		if err != nil {
			klog.Warningf("cannot find vethName in flannel arpAddr is " + string(arpAddr))
			return "", err
		}
		return strings.Replace(string(vethName), "\n", "", -1), nil
	}
	return "", errors.New("can not get VethName")
}

// 修改qdisc的ifb3设备带宽限制
func (net *NetSuppressor) tcChangeBandWidth(beRate string) error {
	tcSuppress := "tc qdisc change dev ifb3  root tbf  rate " + beRate + "  limit  " + beRate + " burst " + beRate
	if err := execShell(tcSuppress); err != nil {
		return err
	}
	klog.V(5).Info("shell - " + tcSuppress)
	return nil
}

func (net *NetSuppressor) RunInit(stopCh <-chan struct{}) error {
	net.setCniType() //获取CNI类型
	go func() {
		// 通过bpftrace 统计BE使用带宽流量数据
		<-net.bpftraceInit("./bpftrace/beNetGather.bt", net.insertBENetData, stopCh)
	}()
	go func() {
		// 通过bpftrace 统计带宽总流量数据
		<-net.bpftraceInit("./bpftrace/totalNetGather.bt", net.insertTotalNetData, stopCh)
	}()
	if err := net.tcInit(); err != nil {
		klog.Warningf("NetSuppressor " + err.Error())
		return err
	}
	return nil
}

// 初始化tc信息
func (net *NetSuppressor) tcInit() error {
	netBEUpper := "1Gbit" //带宽限制
	klog.Infof("NetSuppressor init")
	for i := 0; i < len(tcDelete); i++ {
		execShell(tcDelete[i])
	}
	var tcInit = []string{
		"modprobe ifb",		 									// 加载ifb模块
		"ip link add ifb3 type ifb",							// 设置ifb3设备为ifb
		"ip link set dev ifb3 up txqueuelen 1000",  			// 设置ifb3设备缓冲区的储存长度为1000
		"tc qdisc add dev ifb3 root tbf rate " + netBEUpper +	// 设置ifb3 tbf队列限速
			" burst " + netBEUpper + " limit " + netBEUpper ,
	}
	for i := 0; i < len(tcInit); i++ {
		if err := execShell(tcInit[i]); err != nil {
			return err
		}
	}
	klog.Infof("NetSuppressor inited success")
	return nil
}

// 运行bpftrace 并采集信息
func (net *NetSuppressor) bpftraceInit(bpftracePath string, insertFunc func(str string) error, stopCh <-chan struct{}) <-chan struct{} {
	klog.Infof("bpftrace creating")
	defer klog.Infof("bpftrace over")

	// 运行bpftrace
	cmd := exec.Command("bpftrace", bpftracePath, "10000", "100000", Intervel, Intervel)
	closed := make(chan struct{})

	//通过管道获取监听脚本数据
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		klog.Warningf("bpftrace error: cant find bpftrace or file dont exit")
		return nil
	}
	defer stdoutPipe.Close()
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		// 在命令执行过程中，实时获取其输出
		for scanner.Scan() {
			// 防止垃圾数据
			data, err := simplifiedchinese.GB18030.NewDecoder().Bytes(scanner.Bytes())
			if err != nil {
				klog.Warningf("transfer error with bytes:", scanner.Bytes())
				continue
			}
			//
			if err := insertFunc(string(data)); err != nil {
				klog.Warningf("storage info to metricDB err")
				return
			}
			select {
			case <-stopCh:
				close(closed)
				return
			default:
			}
		}
	}()
	if err := cmd.Run(); err != nil {
		klog.Warningf("bpftrace run err")
		return nil
	}
	return closed
}

// 获取BE带宽数据
func (net *NetSuppressor) insertBENetData(data string) error {
	fmt.Println(fmt.Sprintf("insertBENetData: %s", data))  // 输出数据，这里可以考虑写入监控，或者是数据库，方便使用
	return nil
}

// 获得带宽统计数据
func (net *NetSuppressor) insertTotalNetData(data string) error {
	fmt.Println(fmt.Sprintf("insertTotalNetData: %s", data)) // 输出数据，这里可以考虑写入监控，或者是数据库，方便使用
	return nil
}

func main()  {
	net := NewNetSuppress()
	stopCh := make(chan struct{})
	net.RunInit(stopCh)
	defer func() { stopCh <- struct{}{} }()
}
