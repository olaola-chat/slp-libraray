package tool

import (
	"reflect"
	"sync"

	"github.com/gogf/gf/frame/g"
)

var Gr = &goroutine{}

type goroutine struct{}

func (*goroutine) SyncGo(wg *sync.WaitGroup, fn func()) {
	//wg.Add(1)
	func() {
		defer func() {
			if err := recover(); err != nil {
				g.Log().Error("functionPanic Err", err)
			}
			//wg.Done()
		}()
		fn()
	}()
}

// LoopCall 遍历执行结构体所有可导出方法(顺序执行)
func (grt *goroutine) LoopCall(ins interface{}) bool {
	if ins == nil {
		return false
	}
	tk := reflect.TypeOf(ins).Kind()
	if tk != reflect.Ptr && tk != reflect.Struct {
		return false
	}
	values := reflect.ValueOf(ins)
	methodNums := values.NumMethod()
	if methodNums == 0 {
		return false
	}
	wg := &sync.WaitGroup{}
	for i := 0; i < methodNums; i++ {
		values.Method(i).Call(nil)
	}
	wg.Wait()
	return true
}

func (*goroutine) SyncGoFunc(wg *sync.WaitGroup, fn func()) {
	wg.Add(1)
	func() {
		defer func() {
			if err := recover(); err != nil {
				g.Log().Error("SyncGoFunc functionPanic Err", err)
			}
			wg.Done()
		}()
		fn()
	}()
}
