package tool

import (
	"fmt"
	"net"
	"path"

	"banban/library/tool/region"
)

var (
	//ErrorEmptyInterfaceAddrs 定义错误，无法找到网卡信息时
	ErrorEmptyInterfaceAddrs = fmt.Errorf("empty found in InterfaceAddrs")
)

// IP 单例ip，并且导出
var IP = &ip{}
var ipServ = &region.IP2Region{}
var ipDataPath string

func init() {
	dir := Path.ExecRootPath()
	ipDataPath = path.Join(dir, "config", "ip2region.db")
}

type ip struct {
}

// for test
func (*ip) setIPDbDir(dir string) {
	ipDataPath = path.Join(dir, "config", "ip2region.db")
}

// GetAddr 根据IP地址
func (*ip) GetAddr(ipv4 string) (region.IPInfo, error) {
	// 线程安全的
	if !ipServ.Initialized() {
		err := ipServ.Init(ipDataPath)
		if err != nil {
			return region.IPInfo{}, err
		}
	}
	return ipServ.MemorySearch(ipv4)
}

// GetAddrLocation 根据IP属地
func (*ip) GetAddrLocation(ipv4 string) string {

	//如果是内网IP，直接返回空
	if IP.IsLanIP(ipv4) {
		return ""
	}

	info, err := IP.GetAddr(ipv4)
	if err != nil || info.Province == "0" {
		return ""
	}

	return info.Province
}

// LocalIPv4s 获取本机局域网ipv4地址
func (*ip) LocalIPv4s() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}

	return "", ErrorEmptyInterfaceAddrs
}

// IsIPV4 判断字符串是否是ipv4
func (*ip) IsIPV4(ipv4 string) bool {
	address := net.ParseIP(ipv4)
	return address != nil
}

// IsLanIP 判断字符串是否是内网IP
func (*ip) IsLanIP(ipv4 string) bool {
	ip := net.ParseIP(ipv4)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return true
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return true
		case ip4[0] == 192 && ip4[1] == 168:
			return true
		default:
			return false
		}
	}
	return false
}
