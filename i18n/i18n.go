package i18n

import (
	"github.com/gogf/gf/i18n/gi18n"
)

// NewI18n 实例化一个I18n对象，每个http请求都应该实例一个
func NewI18n() *I18n {
	return &I18n{}
}

// I18n 封装下当前请求用户的language，避免每次格式化都需要传递
type I18n struct {
	language string
}

// SetLanguage 设置当前会话的语言
func (n *I18n) SetLanguage(language string) {
	n.language = language
}

// GetLanguage 设置当前会话的语言
func (n *I18n) GetLanguage() string {
	return n.language
}

// T 根据key返回当前会话的语言
func (n *I18n) T(content string) string {
	return gi18n.Translate(content, n.language)
}

// Tf 根据key返回当前会话的语言，支持格式化
func (n *I18n) Tf(format string, values ...interface{}) string {
	return gi18n.Tfl(n.language, format, values...)
}
