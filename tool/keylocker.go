package tool

import (
	"sync"
	"time"
)

var DefaultKeyLocker *KeyLocker = &KeyLocker{}

type KeyLocker struct {
	m sync.Map
}

func (k *KeyLocker) TryLock(key interface{}) bool {
	_, ok := k.m.LoadOrStore(key, struct{}{})
	return !ok
}

func (k *KeyLocker) WaitLock(key interface{}, retry ...int) bool {
	num := 1000 //1秒钟
	if len(retry) > 0 {
		num = retry[0]
	}
	for i := 0; i < num; i++ {
		if k.TryLock(key) {
			return true
		} else {
			time.Sleep(time.Millisecond)
		}
	}
	return false
}

func (k *KeyLocker) UnLock(key interface{}) {
	k.m.Delete(key)
}
