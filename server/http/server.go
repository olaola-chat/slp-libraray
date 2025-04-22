package http

import (
	"os"

	"github.com/gogf/gf/net/ghttp"

	"github.com/olaola-chat/slp-library/acm"
	"github.com/olaola-chat/slp-library/loghook"

	_ "github.com/olaola-chat/slp-library/tracer"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/urfave/cli"
)

func Run(route func(server *ghttp.Server)) {

	g.Log().SetAsync(true)
	g.Log().SetHeaderPrint(true)
	g.Log().SetFlags(glog.F_FILE_SHORT)
	g.Log().SetStack(false)
	g.Log().Info("work begin")

	acm.GetAcm()

	var cfgName string

	ca := cli.NewApp()
	ca.Name = "http server"
	ca.Version = "0.0.1"
	ca.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "gf.gcfg.file",
			Usage:       "config name",
			Value:       "config.toml",
			Destination: &cfgName,
		},
	}
	ca.Action = func(c *cli.Context) error {
		return nil
	}
	err := ca.Run(os.Args)
	if err != nil {
		panic(err)
	}

	//设置日志
	g.Log().SetWriter(loghook.NewLogWriter("SlpGoHttp"))

	appRun(route)
}
