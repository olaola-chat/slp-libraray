package plugins

import (
	"context"
	"net"

	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/juju/ratelimit"
	"github.com/smallnest/rpcx/protocol"
)

// InfoHandler 定义了有个RPC server插件
type InfoHandler struct {
	bucket *ratelimit.Bucket
}

// NewInfoPlugin 实例化一个 InfoHandler
func NewInfoPlugin(limit, capacity int64) *InfoHandler {
	return &InfoHandler{
		bucket: ratelimit.NewBucketWithRate(float64(limit), capacity),
	}
}

// HeartbeatRequest 心跳的回调
func (h *InfoHandler) HeartbeatRequest(ctx context.Context, req *protocol.Message) error {
	//conn := ctx.Value(server.RemoteConnContextKey).(net.Conn)
	//println("OnHeartbeat:", conn.RemoteAddr().String(), req.SerializeType())
	return nil
}

// HandleConnAccept 当有客户端建立链接时回调
func (h *InfoHandler) HandleConnAccept(conn net.Conn) (net.Conn, bool) {
	g.Log().Println("has accept this connection:", conn.RemoteAddr().String())
	return conn, true
}

// HandleConnClose 当有客户端关闭链接时回调
func (h *InfoHandler) HandleConnClose(conn net.Conn) bool {
	g.Log().Println("has closed this connection:", conn.RemoteAddr().String())
	return true
}

// PreCall 当有客户端发起函数调用是回调
// 这里是限流
func (h *InfoHandler) PreCall(ctx context.Context, serviceName, methodName string, args interface{}) (interface{}, error) {
	//conn, ok := ctx.Value(server.RemoteConnContextKey).(net.Conn)
	//if ok {
	//	g.Log().Println("call", conn.RemoteAddr().String(), serviceName, methodName)
	//}
	ok := h.bucket.TakeAvailable(1) > 0
	var err error = nil
	if !ok {
		err = gerror.Newf("rpc call is limited, service %s %s", serviceName, methodName)
	}
	return args, err
}
