package tool

import (
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	chinese = "中华人民共和国"
	yin     = "zhong hua ren min gong he guo"
)

func TestPinyin(t *testing.T) {
	var total int64 = 0
	var num int64 = 0

	rand.Seed(time.Now().UnixNano())
	var max int = rand.Intn(100)

	wg := sync.WaitGroup{}

	for i := 0; i < max; i++ {
		wg.Add(1)
		go func() {
			pin := Pinyin()
			for j := 0; j < max; j++ {
				atomic.AddInt64(&total, 1)
				val, _ := pin.ConvertWithoutTone(chinese)
				if strings.Join(val, " ") == yin {
					atomic.AddInt64(&num, 1)
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	t.Log("TestPinyin", max, num, num == total)
}
