package middleware

import (
	"context"

	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/util/gconv"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"

	context2 "github.com/olaola-chat/slp-library/server/http/context"
)

const LogTrackKey = "TraceId"
const TraceingEnabled = "TraceingEnabled"

func GetSpan(ctx context.Context) opentracing.Span {
	return opentracing.SpanFromContext(ctx)

}

func Trace(r *ghttp.Request) {
	rootSpan, ctx := opentracing.StartSpanFromContext(r.Context(), r.Request.URL.Path)
	r.SetCtx(ctx)

	//把traceID写入context，便于其他组件获取
	//比如 日志 g.Log().Ctx(r.Context())
	//是否开启了链路追踪...
	r.SetCtxVar(LogTrackKey, rootSpan.Context().(jaeger.SpanContext).TraceID().String())
	r.SetCtxVar(TraceingEnabled, true)
	defer rootSpan.Finish()

	defer func() {
		//提前关闭session
		//框架默认在整个业务之后关闭，会导致链路无法追踪
		//代码做了更改，可以多次close
		// r.Session.Close()
		// 删除敏感数据
		if len(r.GetString("token")) > 0 {
			// 链路追踪的一些数据
			values := r.URL.Query()
			values.Del("token")
			r.URL.RawQuery = values.Encode()
			rootSpan.SetTag("http.url", r.URL.String())
		} else {
			rootSpan.SetTag("http.url", r.RequestURI)
		}

		ctxUser, ok := r.GetCtxVar(context2.ContextUserKey).Interface().(*context2.ContextUser)
		if !ok {
			ctxUser = &context2.ContextUser{}
		}
		if ctxUser.UID > 0 {
			//认证成功
			if rootSpan != nil {
				rootSpan.SetTag("uid", gconv.String(ctxUser.UID))
				// TODO: 根据配置支持的language过滤，FallDown到一个默认语言
				rootSpan.SetTag("language", ctxUser.Language)
			}
		}

		rootSpan.SetTag("sign", r.GetQueryString("_sign"))
		rootSpan.SetTag("http.method", r.Method)
		rootSpan.SetTag("http.status_code", r.Response.Status)
		if err := r.GetError(); err != nil {
			rootSpan.SetTag("error", err.Error())
			rootSpan.LogFields(
				log.String("event", "error"),
				log.String("stack", err.Error()),
			)
		}
	}()

	r.Middleware.Next()
}
