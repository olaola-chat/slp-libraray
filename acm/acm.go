package acm

import (
	"sync"
)

// KeyCallback 数据变更回调方法
type KeyCallback func(value string) error

// DirCallback 数据变更回调方法
type DirCallback func(kvs map[string]string) error

type Acm interface {
	ListenKey(key string, cb KeyCallback) error
	ListenDir(key string, cb DirCallback) error
}

var _acm *acm
var acmOnce sync.Once

// GetAcm 获取单例
func GetAcm() Acm {
	acmOnce.Do(func() {
		_acm = &acm{}
		_acm.init()
		go _acm.run()
	})
	return _acm
}
