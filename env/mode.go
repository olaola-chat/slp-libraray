package env

import (
	"fmt"
	"github.com/gogf/gf/frame/g"
	"os"
	"strings"
)

// IsDev 表明当前系统是不是dev
// 不要在init初始化函数中使用
var IsDev bool = false
var RunMode string = "prod"

func init() {
	//通过机器来检测是不是alpha，这样来统一所有的配置和部署
	mode := g.Cfg().GetString("server.RunMode")
	alphaHosts := g.Cfg().GetStrings("server.AlphaHosts")
	if alphaHosts != nil && mode == "prod" {
		host, err := os.Hostname()
		if err == nil {
			host = strings.ToLower(host)
			for _, alpha := range alphaHosts {
				if host == strings.ToLower(alpha) {
					//是alpha服务器
					mode = "alpha"
				}
			}
		}
	}

	IsDev = mode == "dev" || len(mode) == 0
	RunMode = mode

	fmt.Println("server run with", RunMode)
}
