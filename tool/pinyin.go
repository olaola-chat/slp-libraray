package tool

import (
	"github.com/olaola-chat/slp-library/tool/pinyin"

	"path"

	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/gins"
)

// Pinyin 一个单例，防止资源重复加载
func Pinyin() *pinyin.Pinyin {
	instanceKey := "self-pinyin"
	result := gins.GetOrSetFuncLock(instanceKey, func() interface{} {
		dir := Path.ExecRootPath()
		pin := pinyin.New()
		err := pin.Init(path.Join(dir, "config", "pinyin.txt"))
		if err != nil {
			//因配置问题造成，直接 panic
			panic(err)
		}
		return pin
	})
	if pin, ok := result.(*pinyin.Pinyin); ok {
		return pin
	}
	panic(gerror.New("get pinyin instance error"))
}
