package wrap

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/smallnest/rpcx/share"
)

const (
	// ContextKey 上下文变量存储键名，前后端系统共享
	ContextKey = "ContextKey"
	// TrackKey 上下文传递
	TrackKey = "TraceId"
	// TraceingEnabled 当前请求是否开启
	TraceingEnabled = "TraceingEnabled"
)

// StartOpentracingSpan 从Context里产生一个新的节点
func StartOpentracingSpan(ctx context.Context, name string) (opentracing.Span, context.Context) {
	value := ctx.Value(share.OpentracingSpanServerKey)
	if value != nil {
		span, ok := value.(opentracing.Span)
		if ok {
			spanCtx := opentracing.ContextWithSpan(ctx, span)
			return opentracing.StartSpanFromContext(spanCtx, name)
		}
		return nil, ctx
	}
	if !GetOpentracingEnabled(ctx) {
		return nil, ctx
	}
	//这个是正常的http请求来的
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		return opentracing.StartSpanFromContext(ctx, name)
	} else {
		return nil, ctx
	}
}

func GetOpentracingEnabled(ctx context.Context) bool {
	val := ctx.Value(TraceingEnabled)
	if v, ok := val.(bool); ok {
		return v
	}
	return false
}
