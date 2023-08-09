package response

import (
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/util/gconv"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func getRequestFormat(r *ghttp.Request) string {
	value := r.GetString("format")
	if len(value) > 0 {
		return value
	}
	value = r.GetHeader("Accept")
	if len(value) > 0 {
		switch value {
		case "application/json":
			return "json"
		case "application/protobuf":
			return "pb"
		}
	}
	return ""
}

// Output 输出数据，自动选择
func Output(r *ghttp.Request, data interface{}) {
	value := getRequestFormat(r)
	if len(value) > 0 {
		switch value {
		case "pb":
			content, ok := data.(proto.Message)
			if ok {
				OutputPb(r, content)
			} else {
				OutputServerError(r, gerror.New("The server cannot output pb format data"))
			}
			return
		case "json":
			OutputJSON(r, data)
			return
		}
	}
	switch data := data.(type) {
	case proto.Message:
		OutputPb(r, data)
	default:
		OutputJSON(r, data)
	}
}

// OutputServerError 格式化500错误
func OutputServerError(r *ghttp.Request, err error) {
	r.Response.Header().Set("Content-Type", "text/html")
	r.Response.Status = 500
	r.Response.Write(err.Error())
	r.Exit()
}

// OutputJSON 输出JSON数据
func OutputJSON(r *ghttp.Request, data interface{}) {
	pbMessage, ok := data.(proto.Message)
	if ok {
		content := protojson.MarshalOptions{
			Multiline:       true,
			EmitUnpopulated: true,
			UseProtoNames:   true, //dart是怎么访问的？骆驼法则 or 小写
		}.Format(pbMessage)
		r.Response.Header().Set("Content-Type", "text/json; charset=utf-8")
		r.Response.Header().Set("Content-Length", gconv.String(len(content)))
		r.Response.Write(content)
		r.Exit()
	} else {
		err := r.Response.WriteJson(data)
		if err != nil {
			OutputServerError(r, err)
		} else {
			r.Exit()
		}
	}
}

// OutputPb 输出Pb数据
func OutputPb(r *ghttp.Request, data proto.Message) {
	bs, err := proto.Marshal(data)
	if err != nil {
		OutputServerError(r, err)
	} else {
		r.Response.Header().Set("Content-Type", "application/protobuf")
		r.Response.Header().Set("Content-Length", gconv.String(len(bs)))
		r.Response.Write(bs)
		r.Exit()
	}
}

// OutputJSONSuccess 返回JSON成功数据并退出当前HTTP执行函数。
func OutputJSONSuccess(r *ghttp.Request, data interface{}, options ...g.MapStrAny) {
	resp := g.MapStrAny{
		"success": true,
		"data":    data,
	}
	if len(options) > 0 {
		for k, v := range options[0] {
			resp[k] = v
		}
	}
	OutputJSON(r, resp)
}

// OutputJSONError 返回JSON失败数据并退出当前HTTP执行函数。
func OutputJSONError(r *ghttp.Request, msg string, options ...g.MapStrAny) {
	resp := g.MapStrAny{
		"success": false,
		"msg":     msg,
	}
	if len(options) > 0 {
		for k, v := range options[0] {
			resp[k] = v
		}
	}
	OutputJSON(r, resp)
}

// OutputJSON 输出JSON数据(为了兼容PHP可能字段不存在情况，这个方法配合proto3的wrappers不会主动输出类型类型默认值的情况)
func OutputJSONNotEmit(r *ghttp.Request, data interface{}) {
	pbMessage, ok := data.(proto.Message)
	if ok {
		content := protojson.MarshalOptions{
			Multiline:     true,
			UseProtoNames: true, //dart是怎么访问的？骆驼法则 or 小写
		}.Format(pbMessage)
		r.Response.Header().Set("Content-Type", "text/json; charset=utf-8")
		r.Response.Header().Set("Content-Length", gconv.String(len(content)))
		r.Response.Write(content)
		r.Exit()
	} else {
		err := r.Response.WriteJson(data)
		if err != nil {
			OutputServerError(r, err)
		} else {
			r.Exit()
		}
	}
}
