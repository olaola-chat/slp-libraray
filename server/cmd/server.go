package cmd

import (
	"fmt"
	"github.com/olaola-chat/rbp-library/acm"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/olaola-chat/rbp-library/loghook"
	"github.com/olaola-chat/rbp-library/tool"
	_ "github.com/olaola-chat/rbp-library/tracer"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/urfave/cli"
)

func run(name string, action string, servers []interface{}) {
	name = strings.ToLower("Cmd") + strings.ToLower(name)
	for _, pt := range servers {
		refT := reflect.TypeOf(pt).String()
		refTName := strings.Split(refT, ".")
		if strings.ToLower(refTName[len(refTName)-1]) == name {
			refV := reflect.ValueOf(pt)
			method := refV.MethodByName(action)
			if method.Kind() == reflect.Func {
				go heartBeat(name, action)

				serverName := "RbpCmd" +
					tool.Str.FirstToUpper(name) +
					tool.Str.FirstToUpper(action)
				g.Log().SetAsync(true)
				g.Log().SetFlags(glog.F_FILE_SHORT)
				g.Log().SetStack(false)
				g.Log().SetWriter(loghook.NewLogWriter(serverName))

				method.Call([]reflect.Value{})
				return
			}
			panic(fmt.Sprintf("error action name with %s", action))
		}
	}
	panic(fmt.Sprintf("error cmd name with %s", name))
}

func heartBeat(name, action string) {
	hbTicker := time.NewTicker(3 * time.Second)
	for {
		<-hbTicker.C
		m := &runtime.MemStats{}
		runtime.ReadMemStats(m)
		g.Log().Printf("%s.%s ### Current memory usage: %dKb ###", name, action, m.Alloc/1024)
	}
}

func Run(servers []interface{}) {
	var cmdName string
	var cmdActionName string
	var cfgName string

	g.Log().SetAsync(true)
	g.Log().SetHeaderPrint(true)
	g.Log().SetFlags(glog.F_FILE_SHORT)
	g.Log().SetStack(false)

	acm.GetAcm()

	ca := cli.NewApp()
	ca.Name = "banban cli server"
	ca.Version = "0.0.1"
	ca.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "gf.gcfg.file",
			Usage:       "config name",
			Value:       "config.toml",
			Destination: &cfgName,
		},
		cli.StringFlag{
			Name:        "name",
			Usage:       "cmd name",
			Value:       "",
			Destination: &cmdName,
		},
		cli.StringFlag{
			Name:        "action",
			Usage:       "action name",
			Value:       "Main",
			Destination: &cmdActionName,
		},
	}
	ca.Action = func(c *cli.Context) error {
		if len(cmdName) == 0 {
			panic(fmt.Errorf("error args cmd name"))
		}
		return nil
	}

	err := ca.Run(os.Args)
	if err != nil {
		panic(err)
	}

	g.Log().SetWriter(loghook.NewLogWriter("RbpGoCmd_" + cmdName + "_" + cmdActionName))

	run(cmdName, cmdActionName, servers)
}
