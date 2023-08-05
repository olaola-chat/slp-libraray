package tool

import (
	"strings"
	"testing"
)

var data = []string{
	"127.0.0.1",
	"192.168.1.1",
	"47.241.98.216",
	"61.183.129.50",
}

func init() {

}

func TestGetAddr(t *testing.T) {
	sp := "/banban/"
	p := Path.GetFilePath()
	index := strings.Index(p, sp) + len(sp) - 1
	t.Log("path", p, p[0:index])

	IP.setIPDbDir(p[0:index])

	for _, ip := range data {
		addr, err := IP.GetAddr(ip)
		t.Log("ip", ip, "=>", addr.String(), err)
	}
}

func TestIsIpV4(t *testing.T) {
	res := IP.IsIPV4("1.0.0.1")
	res2 := IP.IsIPV4("0.1")
	t.Error(res, res2)
}
