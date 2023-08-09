package middleware

import "github.com/gogf/gf/net/ghttp"

// CORS 允许接口跨域请求
func CORS(r *ghttp.Request) {
	corsOptions := r.Response.DefaultCORSOptions()
	//corsOptions.AllowDomain = []string{
	//	"hubeiyihelian.com",
	//	"caihongxq.com",
	//}
	//corsOptions.AllowMethods = "GET,POST"
	//corsOptions.AllowHeaders = "x-requested-with,content-type,user-token,user-language"
	//corsOptions.ExposeHeaders = "date,user-status"
	//corsOptions.AllowCredentials = "true"

	r.Response.CORS(corsOptions)
	r.Middleware.Next()
}
