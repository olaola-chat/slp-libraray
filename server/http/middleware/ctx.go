package middleware

import (
	ctx "context"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"

	i18n2 "github.com/olaola-chat/slp-library/i18n"
	"github.com/olaola-chat/slp-library/server/http/context"
)

type AuthUser struct {
	UID      uint32
	Time     uint32
	AppID    uint8
	Salt     string
	Platform string
	Channel  string
}

type AuthFunc func(ctx ctx.Context, token string) (AuthUser, error)

type ctxMiddleware struct {
	auth AuthFunc
}

func NewCtxMiddleware(cb AuthFunc) *ctxMiddleware {
	return &ctxMiddleware{auth: cb}
}

// Ctx 自定义上下文对象
func (c *ctxMiddleware) Ctx(r *ghttp.Request) {
	rbpI18n := i18n2.NewI18n()
	r.SetCtxVar(context.ContextI18nKey, rbpI18n)

	ctxUser := context.NewContextUserFromRequest(r)
	r.SetCtxVar(context.ContextUserKey, ctxUser)

	//获取用户的token
	//app放在header里面
	//todo... 不确定golang的自定义header key格式
	token := r.GetHeader("User-Token")
	if len(token) == 0 {
		//从cookie中获取
		token = r.Cookie.Get("token")
	}
	if len(token) == 0 {
		//从query中获取
		token = r.GetQueryString("token")
	}
	if len(token) > 5 {
		user, err := c.auth(r.Context(), token)
		if err != nil {
			//这是系统的失败，不能确定token是否有问题，不能把用户踢出去...
			g.Log("exception").Error(err)
			r.Response.WriteStatusExit(500, "something error")
		}
		if user.UID > 0 {
			//认证成功
			ctxUser.UID = user.UID
			ctxUser.Time = user.Time
			ctxUser.AppID = user.AppID
			ctxUser.Salt = user.Salt
			ctxUser.Platform = user.Platform
			ctxUser.Channel = user.Channel
		} else {
			//这是token认证失败，过期或者，被封禁
			//如果是cookie，则删除cookie
			//后面会验证，对有要求登录的接口，会把用户踢出来，客户端会删除本地的token
			r.Cookie.Remove("token")
		}
	}
	// TODO: 过滤支持的语言
	rbpI18n.SetLanguage(ctxUser.Language) //从header中获取的
	r.Validator.SetLanguage(ctxUser.Language)
	// 执行下一步请求逻辑
	r.Middleware.Next()
}
