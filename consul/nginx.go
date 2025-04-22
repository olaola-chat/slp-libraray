package consul

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/olaola-chat/slp-library/env"
	"github.com/olaola-chat/slp-library/tool"

	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/hashicorp/consul/api"
)

const (
	//ServiceName 向注册中心注册的服务名字
	ServiceName = "slp/nginx"
)

var _nginx *nginx
var nginxOnce sync.Once

func GetNginx() *nginx {
	nginxOnce.Do(func() {
		_nginx = &nginx{}
		_nginx.init()
	})

	return _nginx
}

func (ng *nginx) init() {
	cfg := &DiscoverConfig{}
	err := g.Cfg().GetStruct("rpc.discover", cfg)
	if err != nil {
		panic(err)
	}

	addr := g.Cfg().GetString("server.Address")
	if len(addr) == 0 {
		panic(gerror.New("error with config server.Address, because of empty"))
	}
	rs := strings.Split(addr, ":")
	if len(rs) != 2 {
		panic(gerror.New("error with config server.Address, because of format"))
	}

	port, err := strconv.Atoi(rs[1])
	if err != nil {
		panic(gerror.New("error with config server.Address, because of port"))
	}

	ipv4, err := tool.IP.LocalIPv4s()
	if err != nil {
		panic(err)
	}

	ng.Ipv4 = ipv4
	ng.Port = port
	ng.Cfg = cfg
}

type nginx struct {
	Ipv4   string
	Port   int
	Cfg    *DiscoverConfig
	closed bool
}

func (ng *nginx) Close() error {
	if ng.Cfg.Type != "consul" {
		return nil
	}
	if !ng.closed {
		g.Log().Println("nginx close")
		client, err := ng.getClient()
		if err != nil {
			g.Log().Println("nginx getClient error", err)
			return err
		}
		res, err := client.Agent().Members(false)
		if err == nil {
			g.Log().Println("nginx get members len", len(res))
			for _, agent := range res {
				c, e := ng.getDiretClient(fmt.Sprintf("%s:8500", agent.Addr))
				if e == nil {
					e := c.Agent().ServiceDeregister(ng.getID())
					if e == nil {
						g.Log().Println("nginx close from", agent.Addr)
					}
				} else {
					g.Log().Println("get getDiretClient error", e)
				}
			}
		} else {
			err := client.Agent().ServiceDeregister(ng.getID())
			g.Log().Println("nginx close from default", err)
		}

		ng.closed = true
	}
	return nil
}

func (ng *nginx) Query() ([]string, error) {
	client, err := ng.getClient()
	if err != nil {
		return []string{}, err
	}
	res, err := client.Agent().Checks()
	if err != nil {
		return []string{}, err
	}
	data := []string{}
	for _, cfg := range res {
		if cfg.ServiceName == ServiceName && cfg.Status == "passing" {
			data = append(data, cfg.ServiceID)
		}
	}
	return data, nil
}

func (ng *nginx) Regist(prefixs []string) error {
	if ng.Cfg.Type != "consul" {
		return nil
	}
	if len(prefixs) == 0 {
		panic("The slice prefixs must not be empty")
	}

	//把mode写入consul，用于前端nginx agent识别机器...
	mode := env.GetRunMode()
	var tags []string
	if mode == env.RUNMODE_PROD {
		tags = append(tags, "nginx")
		tags = append(tags, string(mode))
	}
	for i := 0; i < len(prefixs); i++ {
		prefix := prefixs[i]
		if len(prefix) < 3 ||
			!strings.HasPrefix(prefix, "/") ||
			!strings.HasSuffix(prefix, "/") ||
			strings.Count(prefix, "/") != 2 {
			panic("The string prefix must start and end with /")
		}
		tags = append(tags, prefix)
	}

	client, err := ng.getClient()
	if err != nil {
		return err
	}
	//创建一个新服务。
	registration := new(api.AgentServiceRegistration)
	registration.ID = ng.getID()
	registration.Name = ServiceName
	registration.Tags = tags
	registration.Address = ng.Ipv4
	registration.Port = ng.Port

	//增加check。
	check := new(api.AgentServiceCheck)
	check.HTTP = fmt.Sprintf("http://%s:%d%s", registration.Address, registration.Port, "/ping")
	check.Timeout = "1s"                         //设置超时 1s
	check.Interval = "3s"                        //设置间隔 5s
	check.DeregisterCriticalServiceAfter = "30s" //check失败后30秒删除本服务，注销时间，相当于过期时间

	registration.Check = check

	return client.Agent().ServiceRegister(registration)
}

func (ng *nginx) getID() string {
	return fmt.Sprintf("%s:%d", ng.Ipv4, ng.Port)
}

func (ng *nginx) getClient() (*api.Client, error) {
	return ng.getDiretClient(ng.Cfg.Addr[0])
}

func (ng *nginx) getDiretClient(ip string) (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = ip
	config.Scheme = "http"
	return api.NewClient(config)
}
