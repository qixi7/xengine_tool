package main

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

// 配置模板变量字段
type configVar struct {
	// 确定的变量
	GSCode int // gs编号前缀=1

	// 动态生成的变量
	SelfIP   string // 本机IP
	Suffix   string // 后缀.取本机IP地址最后三位
	ServerID int    // 服务器物理机编号

	// 读取配置的变量
	// ---------------项目全局字段--------------
	ProjName string // 项目名字
	Encoding string // 网络包压缩算法
	GameCode int    // 项目编号
	WorldID  int    // 世界标记ID

	// -------------业务服务器通用字段-----------
	MultiServerID int // 多服标识ID
	MaxClient     int // 最大客户端负载
	LogLevel      int // log 层级

	// -----------------网络通用---------------
	ClientTimeout int    // 客户端心跳超时时间
	KeyFile       string // 加密KeyFile
	CertFile      string // 加密CertFile

	// 端口监听
	PingPort     int // ping端口
	PortBegin    int
	MaxPortRange int
}

// new
func newConfigVar() configVar {
	cfg := configVar{
		GSCode: 1,
		SelfIP: getIP(),
	}
	ipPart := strings.Split(cfg.SelfIP, ".")
	cfg.Suffix = ipPart[len(ipPart)-1]
	return cfg
}

// 获取ip地址最后一组数字
// 返回本机IP
func getIP() string {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Errorf("getIPFinalNum err=%v\n", err.Error())
		os.Exit(1)
	}
	for _, address := range addr {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	fmt.Errorf("getIP failed.\n")
	os.Exit(2)
	return ""
}

// load
func (c *configVar) LoadJsonFile(path string) bool {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("load json file, path=%s, err=%v\n", path, err)
		return false
	}
	err = jsoniter.Unmarshal(f, c)
	if err != nil {
		fmt.Printf("Unmarshal json file, path=%s, err=%v\n", path, err)
		return false
	}
	return true
}
