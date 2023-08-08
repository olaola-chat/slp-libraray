package env

import (
	"github.com/gogf/gf/frame/g"
	"os"
	"strings"
	"sync"
)

type runMode string

const RUNMODE_PROD runMode = "prod"
const RUNMODE_ALPH runMode = "alpha"
const RUNMODE_DEV runMode = "dev"

func GetRunMode() runMode {
	getMode()
	return _mode
}

func IsDev() bool {
	getMode()
	return _mode == RUNMODE_DEV
}

var _mode runMode
var modeOnce sync.Once

func getMode() {
	modeOnce.Do(func() {
		mode := g.Cfg().GetString("server.RunMode")
		alphaHosts := g.Cfg().GetStrings("server.AlphaHosts")

		if alphaHosts != nil && mode == string(RUNMODE_PROD) {
			host, err := os.Hostname()
			if err == nil {
				host = strings.ToLower(host)
				for _, alpha := range alphaHosts {
					if host == strings.ToLower(alpha) {
						//是alpha服务器
						mode = string(RUNMODE_ALPH)
					}
				}
			}
		}

		switch mode {
		case string(RUNMODE_PROD):
			_mode = RUNMODE_PROD
		case string(RUNMODE_ALPH):
			_mode = RUNMODE_ALPH
		default:
			_mode = RUNMODE_DEV
		}

		g.Log().Info("server run with ", _mode)
	})
}
