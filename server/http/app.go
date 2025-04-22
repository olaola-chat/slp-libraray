package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/olaola-chat/slp-library/consul"
	_ "github.com/olaola-chat/slp-library/tracer"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/swagger"
)

const (
	//prefixURI 所有请求的统一前缀，便于前端负载进行分发
	prefixURI = "/go/"
)

// Run Http服务启动入口
func appRun(route func(server *ghttp.Server)) {
	server := g.Server()
	server.SetClientMaxBodySize(1024 * 1024 * 20) //20MB
	server.SetNameToUriType(ghttp.URI_TYPE_CAMEL)
	server.SetErrorStack(true)
	server.SetFileServerEnabled(true)
	server.SetKeepAlive(true)
	// server.SetSessionIdName("bbsid")
	// server.SetSessionCookieMaxAge(time.Hour * 24)
	// server.SetSessionStorage(session.NewStorageRedisV8())
	server.Plugin(&swagger.Swagger{})
	// TODO: 支持Swagger Token
	// server.BindHandler("/swagger/token", api.Debug.Token)
	//增加健康检测
	server.BindHandler("/ping", func(r *ghttp.Request) {
		r.Response.Status = http.StatusOK
		r.Response.WriteExit("pong")
	})
	//增加关闭服务接口
	server.BindHandler("/shutdown", func(r *ghttp.Request) {
		//todo... 安全验证
		//先取消注册服务
		_ = consul.GetNginx().Close()
		//等一会，其他服务需要时间
		time.Sleep(time.Second * 3)
		r.Response.Write("ok")
		//一秒后停止服务
		gtimer.SetTimeout(time.Second, func() {
			err := g.Server().Shutdown()
			if err != nil {
				g.Log().Error("Shutdown Server Error", err)
			}
		})
	})
	server.BindHandler("/unregister", func(r *ghttp.Request) {
		//先取消注册服务
		_ = consul.GetNginx().Close()
		//等一会，其他服务需要时间
		time.Sleep(time.Second * 3)
		r.Response.Write("ok")
	})

	route(server)

	err := server.Start()
	if err != nil {
		panic(err)
	}

	//注册当前http服务的请求前缀
	//nginx会根据uri前缀匹配来转发
	//路径格式必须如 如 /xxxx/ , 且不能包含多级路径
	//系统没有做更加严格的校验，但最好路径名字只包含 [a-z0-9]
	//自动读取路由，把符合条件的注册到服务
	items := server.GetRouterArray()
	tags := []string{}
	tagsHash := map[string]bool{}
	prefixURLLength := len(prefixURI)
	for _, item := range items {
		if !strings.HasPrefix(item.Route, prefixURI) {
			continue
		}
		path := item.Route[prefixURLLength-1:]
		rs := strings.Split(path, "/")
		if len(rs) < 3 || len(rs[1]) == 0 {
			continue
		}
		prefix := rs[1]
		if _, ok := tagsHash[prefix]; !ok {
			tagsHash[prefix] = true
			tags = append(tags, fmt.Sprintf("/%s/", prefix))
		}
	}
	err = consul.GetNginx().Regist(tags)
	if err != nil {
		panic(err)
	}

	g.Wait()

	//关闭注册服务
	_ = consul.GetNginx().Close()
}
