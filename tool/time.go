package tool

import (
	"time"
)

var Time = &tm{}

type tm struct{}

// 单位秒，纳秒精度
func (*tm) NowFloat() float64 {
	t := time.Now()
	return float64(t.UnixNano()) / 1e9
}
