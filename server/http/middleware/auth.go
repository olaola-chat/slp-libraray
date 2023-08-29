package middleware

import (
	"encoding/json"

	"github.com/olaola-chat/rbp-library/i18n"
	context2 "github.com/olaola-chat/rbp-library/server/http/context"
	"github.com/olaola-chat/rbp-library/tool"

	"github.com/gogf/gf/net/ghttp"
)

func Auth(r *ghttp.Request) {
	ctxUser, ok := r.GetCtxVar(context2.ContextUserKey).Interface().(*context2.ContextUser)
	if !ok {
		ctxUser = &context2.ContextUser{}
	}
	if ctxUser.IsLogined() {
		r.Middleware.Next()
	} else {
		ctxI18n, ok := r.GetCtxVar(context2.ContextI18nKey).Interface().(*i18n.I18n)
		if !ok {
			ctxI18n = i18n.NewI18n()
		}
		msg := tool.Str.EscapeUnicode(ctxI18n.T("system.need_login"))
		val, _ := json.Marshal(map[string]interface{}{
			"status": 403,
			"msg":    msg,
		})
		r.Response.Header().Set("User-Status", string(val))
		r.Response.WriteExit([]byte(msg))

	}
}
