package context

import (
	"github.com/gogf/gf/frame/g"
	"github.com/opentracing/opentracing-go"

	"github.com/olaola-chat/rbp-library/i18n"
)

// Context 注入到http request里面，用于上下文共享
type Context struct {
	User *ContextUser
	Data g.Map
	I18n *i18n.I18n
	Span opentracing.Span
}
