package rpc

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/olaola-chat/rbp-library/acm"
	"github.com/olaola-chat/rbp-library/env"
	"github.com/olaola-chat/rbp-library/server/rpc/plugins"

	"github.com/olaola-chat/rbp-library/loghook"
	"github.com/olaola-chat/rbp-library/tool"
	_ "github.com/olaola-chat/rbp-library/tracer"

	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/libkv/store"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	"github.com/urfave/cli"
)

var config *discoverConfig

type discoverConfig struct {
	Type string
	Addr []string
	Path string
}

var myRand *rand.Rand

func init() {
	myRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	config = &discoverConfig{}
	err := g.Cfg().GetStruct("rpc.discover", config)
	if err != nil {
		panic(gerror.Wrap(err, "rpc discover config error"))
	}

}

// NewServer 创建rpc server服务
func NewServer(addr string, limit ...int64) *server.Server {
	fmt.Println("NewServer", addr, config.Type, config.Addr, config.Path)
	rpcServer := server.NewServer(
		server.WithReadTimeout(time.Second*3),
		server.WithWriteTimeout(time.Second*3),
	)
	//一秒产生1万个令牌，桶最大存储两倍于每秒产生的令牌
	var lt int64 = 10000
	if len(limit) > 0 && limit[0] > 0 {
		lt = limit[0]
	}
	rpcServer.Plugins.Add(plugins.NewInfoPlugin(lt, lt*2))
	rpcServer.Plugins.Add(serverplugin.OpenTracingPlugin{})
	rpcServer.DisableHTTPGateway = true

	switch config.Type {
	case "redis":
		discover := &serverplugin.RedisRegisterPlugin{
			ServiceAddress: "tcp@" + addr,
			RedisServers:   config.Addr,
			BasePath:       config.Path,
			Metrics:        metrics.NewRegistry(),
			UpdateInterval: time.Second * 3,
			Options: &store.Config{
				PersistConnection: true,
			},
		}
		err := discover.Start()
		if err != nil {
			panic(err)
		}
		rpcServer.Plugins.Add(discover)
	case "consul":
		discover := &serverplugin.ConsulRegisterPlugin{
			ServiceAddress: "tcp@" + addr,
			ConsulServers:  config.Addr,
			BasePath:       config.Path,
			Metrics:        metrics.DefaultRegistry,
			UpdateInterval: time.Second * 10, //这个更新的是Metrics
		}
		err := discover.Start()
		if err != nil {
			panic(err)
		}
		rpcServer.Plugins.Add(discover)
		//startMetrics()
	default:
		panic(gerror.Newf("error discover type %s", config.Type))
	}

	return rpcServer
}

// LocalIPWithPort 自动生成ip:port
func LocalIPWithAutoPort() string {
	ip, err := tool.IP.LocalIPv4s()
	if err != nil {
		panic(err)
	}
	// rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s:%d", ip, 10000+myRand.Int31n(10000))
}

func CreateRpcServer(sCfg *ServerCfg, closed chan bool) {
	var rpcServer *server.Server
	go func() {
		addr := LocalIPWithAutoPort()
		rpcServer = NewServer(addr, 20000)
		rpcServer.DisableHTTPGateway = false
		err := rpcServer.RegisterName(
			sCfg.RegisterName,
			sCfg.Server(),
			fmt.Sprintf("group=%s", env.GetRunMode()),
		)
		if err != nil {
			panic(err)
		}
		err = rpcServer.Serve("tcp", addr)
		if err != nil && err != server.ErrServerClosed {
			panic(err)
		}
	}()

	<-closed
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	err := rpcServer.UnregisterAll()
	if err != nil {
		g.Log().Println(err)
	}
	err = rpcServer.Shutdown(ctx)
	if err != nil {
		g.Log().Println(err)
	}
	cancel()
}

var (
	closed chan bool = make(chan bool)
)

type ServerCfg struct {
	RegisterName string
	Server       func() interface{}
}

// run 根据name实例化RPC服务
func run(name string, servers map[string]*ServerCfg, pwg *sync.WaitGroup) {
	defer func() {
		if pwg != nil {
			pwg.Done()
		}
	}()
	if name == "all" {
		wg := sync.WaitGroup{}
		for n := range servers {
			g.Log().Info("run rpc", name)
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				run(name, servers, pwg)
			}(n)
		}
		wg.Wait()
		return
	}
	s, ok := servers[name]
	if !ok {
		panic(fmt.Sprintf("error rpc name with %s \n all rpc instance blow:\n\t%+v", name, servers))
	}
	CreateRpcServer(s, closed)
}

func Run(servers map[string]*ServerCfg) {
	g.Log().SetAsync(true)
	g.Log().SetHeaderPrint(true)
	g.Log().SetFlags(glog.F_FILE_SHORT)
	g.Log().SetStack(false)
	g.Log().Info("work begin")

	acm.GetAcm()

	var serviceName string
	var cfgName string

	ca := cli.NewApp()
	ca.Name = "banban rpc server"
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
			Usage:       "rpc name",
			Value:       "",
			Destination: &serviceName,
		},
	}
	ca.Action = func(ctx *cli.Context) error {
		if len(serviceName) == 0 {
			panic(fmt.Errorf("error args service name"))
		}
		return nil
	}

	err := ca.Run(os.Args)
	if err != nil {
		panic(err)
	}

	g.Log().SetWriter(loghook.NewLogWriter("RbpRpc." + tool.Str.FirstToUpper(serviceName)))

	var pwg sync.WaitGroup
	pwg.Add(1)

	go run(serviceName, servers, &pwg)

	sign := make(chan os.Signal, 1)
	signal.Notify(sign, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	v := <-sign
	g.Log().Info("rpc server receive sign", v)
	if serviceName == "all" {
		//dev
		for range servers {
			closed <- true
		}
	} else {
		closed <- true
	}
	pwg.Wait()
	g.Log().Info("rpc server closed complete")
}
